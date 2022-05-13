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

package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"

	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/pkg/dfnet"
	"d7y.io/dragonfly/v2/pkg/rpc/base"
	"d7y.io/dragonfly/v2/pkg/rpc/base/common"
	"d7y.io/dragonfly/v2/pkg/rpc/dfdaemon"
	dfdaemonclient "d7y.io/dragonfly/v2/pkg/rpc/dfdaemon/client"
	"d7y.io/dragonfly/v2/pkg/rpc/dfproxy"
	"google.golang.org/grpc"
)

type Service struct {
	ready   chan<- bool
	reqCh   chan dfproxy.DaemonProxyServerPacket
	reqMap  map[uint64](chan dfproxy.DaemonProxyClientPacket)
	nextSeq uint64
}

// New service instance
func New(ready chan<- bool) *Service {
	return &Service{
		ready:   ready,
		reqCh:   make(chan dfproxy.DaemonProxyServerPacket),
		reqMap:  make(map[uint64](chan dfproxy.DaemonProxyClientPacket)),
		nextSeq: 0,
	}
}

// Dfdaemon handles the proxied dfdaemon requests.
func (s *Service) Dfdaemon(stream dfproxy.DaemonProxy_DfdaemonServer) error {
	defer func() {
		close(s.reqCh)
	}()

	ctx := stream.Context()
	select {
	case <-ctx.Done():
		logger.Infof("Dfproxy Dfdaemon context was done")
		return ctx.Err()
	default:
	}

	// Dfdaemon receives the first packet, which should be ReqType_HeartBeat, subsequent packets
	// are handled by req type handlers accordingly.
	logger.Debug("Dfproxy Dfdaemon receving initial HeartBeat packet")
	clientPkt, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			logger.Info("Dfproxy Dfdaemon service received EOF")
			return nil
		}
		msg := fmt.Sprintf("Dfproxy Dfdaemon service received error: %s", err.Error())
		logger.Error(msg)
		return errors.New(msg)
	}

	if err := checkReqType(clientPkt.GetType(), dfproxy.ReqType_HeartBeat); err != nil {
		return err
	}

	// Let others know service is ready
	s.ready <- true
	logger.Debug("Dfproxy Dfdaemon received initial HeartBeat packet, stream is ready")

	// We got initial HeartBeat request, and we're ready to send server packet back to client as
	// dfdaemon requests.
	stoppedCh := make(chan struct{})

	// Start a goroutine to receive ClientPacket from stream and dispatch back to corresponding
	// requester.
	go func() {
		for {
			select {
			case <-stoppedCh:
				logger.Info("Dfproxy Dfdaemon stops receving ClientPacket")
				return
			default:
			}

			logger.Debug("dfproxy service waiting for ClientPacket")
			clientPkt, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					msg := ("Dfproxy Dfdaemon service received EOF")
					logger.Warn(msg)
					return
				}
				msg := fmt.Sprintf("Dfproxy Dfdaemon service received error: %s", err.Error())
				logger.Error(msg)
				return
			}
			resCh, ok := s.reqMap[clientPkt.GetSeqNum()]
			if !ok {
				logger.Warnf("invalid client packet seq number %v", clientPkt.GetSeqNum())
				continue
			}
			resCh <- *clientPkt
		}
	}()

	// Start a loop to receive ServerPacket and handle it in a goroutine
	for {
		select {
		case <-ctx.Done():
			logger.Infof("Dfproxy Dfdaemon context done: %s", ctx.Err())
			s.ready <- false
			close(stoppedCh)
			return ctx.Err()
		case serverPkt := <-s.reqCh:
			resCh, ok := s.reqMap[serverPkt.GetSeqNum()]
			if !ok {
				logger.Warnf("invalid server packet seq number %v", serverPkt.GetSeqNum())
				continue
			}
			go func() {
				s.sendServerPacket(stream, serverPkt, resCh)
			}()
		}
	}
}

func (s *Service) sendServerPacket(stream dfproxy.DaemonProxy_DfdaemonServer, serverPkt dfproxy.DaemonProxyServerPacket, clientPktCh chan<- dfproxy.DaemonProxyClientPacket) {
	logger.Debugf("dfproxy service handling new ServerPacket: %v", serverPkt)
	clientPkt := dfproxy.DaemonProxyClientPacket{
		Type: serverPkt.GetType(),
	}
	if err := stream.Send(&serverPkt); err != nil {
		msg := fmt.Sprintf("failed to send server packet: %s", err.Error())
		e := common.NewGrpcDfError(base.Code_UnknownError, msg)
		clientPkt.Error = e
		clientPktCh <- clientPkt
	}
	return
}

func (s *Service) getNextSeq() uint64 {
	return atomic.AddUint64(&s.nextSeq, 1)
}

func (s *Service) newServerPacket(reqType dfproxy.ReqType) (dfproxy.DaemonProxyServerPacket, chan dfproxy.DaemonProxyClientPacket) {
	resCh := make(chan dfproxy.DaemonProxyClientPacket)
	serverPkt := dfproxy.DaemonProxyServerPacket{
		Type:   reqType,
		SeqNum: s.getNextSeq(),
	}
	s.reqMap[serverPkt.GetSeqNum()] = resCh
	return serverPkt, resCh
}

func (s *Service) CheckHealth(ctx context.Context, _target dfnet.NetAddr, _opts ...grpc.CallOption) error {
	reqType := dfproxy.ReqType_CheckHealth
	serverPkt, resCh := s.newServerPacket(reqType)

	logger.Info("dfproxy starts to check health")

	start := time.Now()
	s.reqCh <- serverPkt
	select {
	case clientPkt := <-resCh:
		if err := handleClientPacket(clientPkt, reqType); err != nil {
			return err
		}
		logger.Infof("check health successfully in %.6f s", time.Since(start).Seconds())
		return nil
	case <-ctx.Done():
		return handleContextDone(ctx, "check health timeout")
	}
}

func (s *Service) StatTask(ctx context.Context, req *dfdaemon.StatTaskRequest, _opts ...grpc.CallOption) error {
	reqType := dfproxy.ReqType_StatTask
	serverPkt, resCh := s.newServerPacket(reqType)

	wLog := logger.With("Cid", req.Cid, "Tag", req.UrlMeta.Tag)
	wLog.Info("dfproxy starts to stat task")

	start := time.Now()
	s.reqCh <- serverPkt
	select {
	case clientPkt := <-resCh:
		if err := handleClientPacket(clientPkt, reqType); err != nil {
			return err
		}
		wLog.Infof("task stat successfully in %.6f s", time.Since(start).Seconds())
		return nil
	case <-ctx.Done():
		return handleContextDone(ctx, "stat timeout")
	}
}

func (s *Service) ImportTask(ctx context.Context, req *dfdaemon.ImportTaskRequest, _opts ...grpc.CallOption) error {
	reqType := dfproxy.ReqType_ImportTask
	serverPkt, resCh := s.newServerPacket(reqType)

	wLog := logger.With("Cid", req.Cid, "Tag", req.UrlMeta.Tag, "file", req.Path)
	wLog.Info("dfproxy starts to import task")

	start := time.Now()
	s.reqCh <- serverPkt
	select {
	case clientPkt := <-resCh:
		if err := handleClientPacket(clientPkt, reqType); err != nil {
			return err
		}
		wLog.Infof("task imported successfully in %.6f s", time.Since(start).Seconds())
		return nil
	case <-ctx.Done():
		return handleContextDone(ctx, "import timeout")
	}
}

func (s *Service) ExportTask(ctx context.Context, req *dfdaemon.ExportTaskRequest, _opts ...grpc.CallOption) error {
	reqType := dfproxy.ReqType_ExportTask
	serverPkt, resCh := s.newServerPacket(reqType)

	wLog := logger.With("Cid", req.Cid, "Tag", req.UrlMeta.Tag, "output", req.Output)
	wLog.Info("dfproxy starts to export task")

	start := time.Now()
	s.reqCh <- serverPkt
	select {
	case clientPkt := <-resCh:
		if err := handleClientPacket(clientPkt, reqType); err != nil {
			return err
		}
		wLog.Infof("task exported successfully in %.6f s", time.Since(start).Seconds())
		return nil
	case <-ctx.Done():
		return handleContextDone(ctx, "export timeout")
	}
}

func (s *Service) DeleteTask(ctx context.Context, req *dfdaemon.DeleteTaskRequest, _opts ...grpc.CallOption) error {
	reqType := dfproxy.ReqType_DeleteTask
	serverPkt, resCh := s.newServerPacket(reqType)

	wLog := logger.With("Cid", req.Cid, "Tag", req.UrlMeta.Tag)
	wLog.Info("dfproxy starts to delete task")

	start := time.Now()
	s.reqCh <- serverPkt
	select {
	case clientPkt := <-resCh:
		if err := handleClientPacket(clientPkt, reqType); err != nil {
			return err
		}
		wLog.Infof("task deleted successfully in %.6f s", time.Since(start).Seconds())
		return nil
	case <-ctx.Done():
		return handleContextDone(ctx, "delete timeout")
	}
}

// The following DaemonClient interfaces are not implemented yet.
func (s *Service) Download(ctx context.Context, req *dfdaemon.DownRequest, opts ...grpc.CallOption) (*dfdaemonclient.DownResultStream, error) {
	return nil, errors.New("function not implemented")
}

func (s *Service) GetPieceTasks(ctx context.Context, addr dfnet.NetAddr, ptr *base.PieceTaskRequest, opts ...grpc.CallOption) (*base.PiecePacket, error) {
	return nil, errors.New("function not implemented")
}

func (s *Service) SyncPieceTasks(ctx context.Context, addr dfnet.NetAddr, ptr *base.PieceTaskRequest, opts ...grpc.CallOption) (dfdaemon.Daemon_SyncPieceTasksClient, error) {
	return nil, errors.New("function not implemented")
}

func (s *Service) Close() error {
	return nil
}

func checkReqType(got, expected dfproxy.ReqType) error {
	if got != expected {
		msg := fmt.Sprintf("invalid req type %d %s, expected %d %s", got, dfproxy.ReqType_name[int32(got)], expected, dfproxy.ReqType_name[int32(expected)])
		logger.Error(msg)
		return errors.New(msg)
	}
	return nil
}

func checkClientPacket(clientPkt *dfproxy.DaemonProxyClientPacket, serverPkt *dfproxy.DaemonProxyServerPacket) error {
	if err := checkReqType(clientPkt.GetType(), serverPkt.GetType()); err != nil {
		return err
	}
	if clientPkt.GetSeqNum() != serverPkt.GetSeqNum() {
		return fmt.Errorf("invalid req seq number %v, expected %v", clientPkt.GetSeqNum(), serverPkt.GetSeqNum())
	}
	return nil
}

func handleClientPacket(clientPkt dfproxy.DaemonProxyClientPacket, reqType dfproxy.ReqType) error {
	logger.Debugf("dfproxy service handling new ClientPacket: %v", clientPkt)
	if err := checkReqType(clientPkt.GetType(), reqType); err != nil {
		return err
	}
	if clientPkt.Error != nil {
		return errors.New(clientPkt.Error.String())
	}
	return nil
}

func handleContextDone(ctx context.Context, msg string) error {
	if ctx.Err() == context.DeadlineExceeded {
		logger.Warn(msg)
		return errors.New(msg)
	} else if ctx.Err() != nil {
		return ctx.Err()
	}
	msg = "context done without receving clientPkt"
	logger.Warn(msg)
	return errors.New(msg)
}

var _ dfdaemonclient.DaemonClient = (*Service)(nil)
