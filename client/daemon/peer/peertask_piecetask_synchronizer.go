/*
 *     Copyright 2022 The Dragonfly Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package peer

import (
	"context"
	"io"
	"sync"

	"go.uber.org/atomic"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/pkg/rpc/base"
	"d7y.io/dragonfly/v2/pkg/rpc/dfdaemon"
	dfclient "d7y.io/dragonfly/v2/pkg/rpc/dfdaemon/client"
	"d7y.io/dragonfly/v2/pkg/rpc/scheduler"
)

type pieceTaskSyncManager struct {
	sync.RWMutex
	ctx               context.Context
	ctxCancel         context.CancelFunc
	peerTaskConductor *peerTaskConductor
	pieceRequestCh    chan *DownloadPieceRequest
	workers           map[string]*pieceTaskSynchronizer
}

type pieceTaskSynchronizer struct {
	*logger.SugaredLoggerOnWith
	client            dfdaemon.Daemon_SyncPieceTasksClient
	dstPeer           *scheduler.PeerPacket_DestPeer
	error             atomic.Value
	peerTaskConductor *peerTaskConductor
	pieceRequestCh    chan *DownloadPieceRequest
}

type pieceTaskSynchronizerError struct {
	err error
}

// FIXME for compatibility, sync will be called after the dfclient.GetPieceTasks deprecated and the pieceTaskPoller removed
func (s *pieceTaskSyncManager) sync(pp *scheduler.PeerPacket, request *base.PieceTaskRequest) error {
	var (
		peers  = map[string]bool{}
		errors []error
	)
	peers[pp.MainPeer.PeerId] = true
	// TODO if the worker failed, reconnect and retry
	s.Lock()
	defer s.Unlock()
	if _, ok := s.workers[pp.MainPeer.PeerId]; !ok {
		err := s.newPieceTaskSynchronizer(s.ctx, pp.MainPeer, request)
		if err != nil {
			s.peerTaskConductor.Errorf("main peer SyncPieceTasks error: %s", err)
			errors = append(errors, err)
		}
	}
	for _, p := range pp.StealPeers {
		peers[p.PeerId] = true
		if _, ok := s.workers[p.PeerId]; !ok {
			err := s.newPieceTaskSynchronizer(s.ctx, p, request)
			if err != nil {
				s.peerTaskConductor.Errorf("steal peer SyncPieceTasks error: %s", err)
				errors = append(errors, err)
			}
		}
	}

	// cancel old workers
	if len(s.workers) != len(peers) {
		for p, worker := range s.workers {
			if !peers[p] {
				worker.close()
			}
		}
	}

	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func (s *pieceTaskSyncManager) cleanStaleWorker(destPeers []*scheduler.PeerPacket_DestPeer) {
	var (
		peers = map[string]bool{}
	)
	for _, p := range destPeers {
		peers[p.PeerId] = true
	}

	// cancel old workers
	if len(s.workers) != len(peers) {
		var peersToRemove []string
		for p, worker := range s.workers {
			if !peers[p] {
				worker.close()
				peersToRemove = append(peersToRemove, p)
			}
		}
		for _, p := range peersToRemove {
			delete(s.workers, p)
		}
	}
}

func (s *pieceTaskSyncManager) newPieceTaskSynchronizer(
	ctx context.Context,
	dstPeer *scheduler.PeerPacket_DestPeer,
	request *base.PieceTaskRequest) error {

	request.DstPid = dstPeer.PeerId
	if worker, ok := s.workers[dstPeer.PeerId]; ok {
		// worker is okay, keep it go on
		if worker.error.Load() == nil {
			s.peerTaskConductor.Infof("reuse PieceTaskSynchronizer %s", dstPeer.PeerId)
			return nil
		}
		// clean error worker
		delete(s.workers, dstPeer.PeerId)
	}

	request.DstPid = dstPeer.PeerId
	client, err := dfclient.SyncPieceTasks(ctx, dstPeer, request)
	if err != nil {
		s.peerTaskConductor.Errorf("call SyncPieceTasks error: %s, dest peer: %s", err, dstPeer.PeerId)
		return err
	}

	// TODO the codes.Unimplemented is received only in client.Recv()
	// when remove legacy get piece grpc, can move this check into synchronizer.receive
	piecePacket, err := client.Recv()
	if err != nil {
		s.peerTaskConductor.Warnf("receive from SyncPieceTasksClient error: %s, dest peer: %s", err, dstPeer.PeerId)
		_ = client.CloseSend()
		return err
	}

	synchronizer := &pieceTaskSynchronizer{
		peerTaskConductor:   s.peerTaskConductor,
		pieceRequestCh:      s.pieceRequestCh,
		client:              client,
		dstPeer:             dstPeer,
		error:               atomic.Value{},
		SugaredLoggerOnWith: s.peerTaskConductor.With("targetPeerID", request.DstPid),
	}
	s.workers[dstPeer.PeerId] = synchronizer
	go synchronizer.receive(piecePacket)
	return nil
}

func (s *pieceTaskSyncManager) newMultiPieceTaskSynchronizer(
	destPeers []*scheduler.PeerPacket_DestPeer,
	lastNum int32) (legacyPeers []*scheduler.PeerPacket_DestPeer) {
	s.Lock()
	defer s.Unlock()
	for _, peer := range destPeers {
		request := &base.PieceTaskRequest{
			TaskId:   s.peerTaskConductor.taskID,
			SrcPid:   s.peerTaskConductor.peerID,
			DstPid:   "",
			StartNum: uint32(lastNum),
			Limit:    16,
		}
		err := s.newPieceTaskSynchronizer(s.ctx, peer, request)
		if err == nil {
			s.peerTaskConductor.Infof("connected to peer: %s", peer.PeerId)
			continue
		}
		legacyPeers = append(legacyPeers, peer)
		// when err is codes.Unimplemented, fallback to legacy get piece grpc
		stat, ok := status.FromError(err)
		if ok && stat.Code() == codes.Unimplemented {
			s.peerTaskConductor.Warnf("connect peer %s error: %s, fallback to legacy get piece grpc", peer.PeerId, err)
		} else {
			s.reportError(peer)
			s.peerTaskConductor.Errorf("connect peer %s error: %s, not codes.Unimplemented", peer.PeerId, err)
		}
	}
	s.cleanStaleWorker(destPeers)
	return legacyPeers
}

func compositePieceResult(peerTaskConductor *peerTaskConductor, destPeer *scheduler.PeerPacket_DestPeer) *scheduler.PieceResult {
	return &scheduler.PieceResult{
		TaskId:        peerTaskConductor.taskID,
		SrcPid:        peerTaskConductor.peerID,
		DstPid:        destPeer.PeerId,
		PieceInfo:     &base.PieceInfo{},
		Success:       false,
		Code:          base.Code_ClientPieceRequestFail,
		HostLoad:      nil,
		FinishedCount: peerTaskConductor.readyPieces.Settled(),
	}
}

func (s *pieceTaskSyncManager) reportError(destPeer *scheduler.PeerPacket_DestPeer) {
	sendError := s.peerTaskConductor.sendPieceResult(compositePieceResult(s.peerTaskConductor, destPeer))
	if sendError != nil {
		s.peerTaskConductor.cancel(base.Code_SchedError, sendError.Error())
		s.peerTaskConductor.Errorf("connect peer %s failed and send piece result with error: %s", destPeer.PeerId, sendError)
	}
}

// acquire send the target piece to other peers
func (s *pieceTaskSyncManager) acquire(request *base.PieceTaskRequest) (attempt int, success int) {
	s.RLock()
	for _, p := range s.workers {
		attempt++
		if p.acquire(request) == nil {
			success++
		}
	}
	s.RUnlock()
	return
}

func (s *pieceTaskSyncManager) cancel() {
	s.ctxCancel()
	s.Lock()
	for _, p := range s.workers {
		p.close()
	}
	s.workers = map[string]*pieceTaskSynchronizer{}
	s.Unlock()
}

func (s *pieceTaskSynchronizer) close() {
	if err := s.client.CloseSend(); err != nil {
		s.error.Store(&pieceTaskSynchronizerError{err})
		s.Debugf("close send error: %s, dest peer: %s", err, s.dstPeer.PeerId)
	}
}

func (s *pieceTaskSynchronizer) dispatchPieceRequest(piecePacket *base.PiecePacket) {
	s.peerTaskConductor.updateMetadata(piecePacket)

	pieceCount := len(piecePacket.PieceInfos)
	s.Debugf("dispatch piece request, piece count: %d, dest peer: %s", pieceCount, s.dstPeer.PeerId)
	// fix cdn return zero piece info, but with total piece count and content length
	if pieceCount == 0 {
		finished := s.peerTaskConductor.isCompleted()
		if finished {
			s.peerTaskConductor.Done()
		}
		return
	}
	for _, piece := range piecePacket.PieceInfos {
		s.Infof("got piece %d from %s/%s, digest: %s, start: %d, size: %d",
			piece.PieceNum, piecePacket.DstAddr, piecePacket.DstPid, piece.PieceMd5, piece.RangeStart, piece.RangeSize)
		// FIXME when set total piece but no total digest, fetch again
		s.peerTaskConductor.requestedPiecesLock.Lock()
		if !s.peerTaskConductor.requestedPieces.IsSet(piece.PieceNum) {
			s.peerTaskConductor.requestedPieces.Set(piece.PieceNum)
		}
		s.peerTaskConductor.requestedPiecesLock.Unlock()
		req := &DownloadPieceRequest{
			storage: s.peerTaskConductor.GetStorage(),
			piece:   piece,
			log:     s.peerTaskConductor.Log(),
			TaskID:  s.peerTaskConductor.GetTaskID(),
			PeerID:  s.peerTaskConductor.GetPeerID(),
			DstPid:  piecePacket.DstPid,
			DstAddr: piecePacket.DstAddr,
		}
		select {
		case s.pieceRequestCh <- req:
		case <-s.peerTaskConductor.successCh:
			s.Infof("peer task success, stop dispatch piece request, dest peer: %s", s.dstPeer.PeerId)
		case <-s.peerTaskConductor.failCh:
			s.Warnf("peer task fail, stop dispatch piece request, dest peer: %s", s.dstPeer.PeerId)
		}
	}
}

func (s *pieceTaskSynchronizer) receive(piecePacket *base.PiecePacket) {
	var err error
	for {
		s.dispatchPieceRequest(piecePacket)
		piecePacket, err = s.client.Recv()
		if err != nil {
			break
		}
	}

	if err == io.EOF {
		s.Debugf("synchronizer receives io.EOF")
	} else if s.canceled(err) {
		s.error.Store(&pieceTaskSynchronizerError{err})
		s.Debugf("synchronizer receives canceled")
	} else {
		s.error.Store(&pieceTaskSynchronizerError{err})
		s.reportError()
		s.Errorf("synchronizer receives with error: %s", err)
	}
}

func (s *pieceTaskSynchronizer) acquire(request *base.PieceTaskRequest) error {
	if s.error.Load() != nil {
		err := s.error.Load().(*pieceTaskSynchronizerError).err
		s.Debugf("synchronizer already error %s, skip acquire more pieces", err)
		return err
	}
	err := s.client.Send(request)
	if err != nil {
		s.error.Store(&pieceTaskSynchronizerError{err})
		if s.canceled(err) {
			s.Debugf("synchronizer sends canceled")
		} else {
			s.Errorf("synchronizer sends with error: %s", err)
			s.reportError()
		}
	}
	return err
}

func (s *pieceTaskSynchronizer) reportError() {
	sendError := s.peerTaskConductor.sendPieceResult(compositePieceResult(s.peerTaskConductor, s.dstPeer))
	if sendError != nil {
		s.peerTaskConductor.cancel(base.Code_SchedError, sendError.Error())
		s.Errorf("sync piece info failed and send piece result with error: %s", sendError)
	}
}

func (s *pieceTaskSynchronizer) canceled(err error) bool {
	if err == context.Canceled {
		s.Debugf("context canceled, dst peer: %s", s.dstPeer.PeerId)
		return true
	}
	if stat, ok := err.(interface{ GRPCStatus() *status.Status }); ok {
		if stat.GRPCStatus().Code() == codes.Canceled {
			s.Debugf("grpc canceled, dst peer: %s", s.dstPeer.PeerId)
			return true
		}
	}
	return false
}
