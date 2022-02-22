/*
 *     Copyright 2020 The Dragonfly Authors
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
	"fmt"
	"io"
	"sync"

	"github.com/go-http-utils/headers"
	"github.com/pkg/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/time/rate"

	"d7y.io/dragonfly/v2/client/config"
	"d7y.io/dragonfly/v2/client/daemon/metrics"
	"d7y.io/dragonfly/v2/client/daemon/storage"
	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/pkg/idgen"
	"d7y.io/dragonfly/v2/pkg/rpc/base"
	"d7y.io/dragonfly/v2/pkg/rpc/scheduler"
	schedulerclient "d7y.io/dragonfly/v2/pkg/rpc/scheduler/client"
)

// TaskManager processes all peer tasks request
type TaskManager interface {
	// StartFileTask starts a peer task to download a file
	// return a progress channel for request download progress
	// tiny stands task file is tiny and task is done
	StartFileTask(ctx context.Context, req *FileTaskRequest) (
		progress chan *FileTaskProgress, tiny *TinyData, err error)
	// StartStreamTask starts a peer task with stream io
	// tiny stands task file is tiny and task is done
	StartStreamTask(ctx context.Context, req *StreamTaskRequest) (
		readCloser io.ReadCloser, attribute map[string]string, err error)

	IsPeerTaskRunning(id string) bool

	// Check if the given task exists in P2P network
	StatPeerTask(ctx context.Context, taskID string) (*base.GrpcDfResult, error)

	// AnnouncePeerTask announces peer task info to P2P network
	AnnouncePeerTask(ctx context.Context, meta storage.PeerTaskMetadata, cid string, urlMeta *base.UrlMeta) error

	GetPieceManager() PieceManager

	// Stop stops the PeerTaskManager
	Stop(ctx context.Context) error
}

//go:generate mockgen -source peertask_manager.go -package peer -self_package d7y.io/dragonfly/v2/client/daemon/peer -destination peertask_manager_mock_test.go
//go:generate mockgen -source peertask_manager.go -destination ../test/mock/peer/peertask_manager.go
// Task represents common interface to operate a peer task
type Task interface {
	Logger
	Context() context.Context
	Log() *logger.SugaredLoggerOnWith

	GetStorage() storage.TaskStorageDriver

	GetPeerID() string
	GetTaskID() string

	GetTotalPieces() int32
	SetTotalPieces(int32)

	GetContentLength() int64
	SetContentLength(int64)

	AddTraffic(uint64)
	GetTraffic() uint64

	SetPieceMd5Sign(string)
	GetPieceMd5Sign() string

	PublishPieceInfo(pieceNum int32, size uint32)
	ReportPieceResult(request *DownloadPieceRequest, result *DownloadPieceResult, err error)
}

type Logger interface {
	Log() *logger.SugaredLoggerOnWith
}

type TinyData struct {
	TaskID  string
	PeerID  string
	Content []byte
}

var tracer trace.Tracer

func init() {
	tracer = otel.Tracer("dfget-daemon")
}

type peerTaskManager struct {
	host            *scheduler.PeerHost
	schedulerClient schedulerclient.SchedulerClient
	schedulerOption config.SchedulerOption
	pieceManager    PieceManager
	storageManager  storage.Manager

	conductorLock    sync.Locker
	runningPeerTasks sync.Map

	perPeerRateLimit rate.Limit

	// enableMultiplex indicates to reuse the data of completed peer tasks
	enableMultiplex bool
	// enablePrefetch indicates to prefetch the whole files of ranged requests
	enablePrefetch bool

	calculateDigest bool

	getPiecesMaxRetry int
}

func NewPeerTaskManager(
	host *scheduler.PeerHost,
	pieceManager PieceManager,
	storageManager storage.Manager,
	schedulerClient schedulerclient.SchedulerClient,
	schedulerOption config.SchedulerOption,
	perPeerRateLimit rate.Limit,
	multiplex bool,
	prefetch bool,
	calculateDigest bool,
	getPiecesMaxRetry int) (TaskManager, error) {

	ptm := &peerTaskManager{
		host:              host,
		runningPeerTasks:  sync.Map{},
		conductorLock:     &sync.Mutex{},
		pieceManager:      pieceManager,
		storageManager:    storageManager,
		schedulerClient:   schedulerClient,
		schedulerOption:   schedulerOption,
		perPeerRateLimit:  perPeerRateLimit,
		enableMultiplex:   multiplex,
		enablePrefetch:    prefetch,
		calculateDigest:   calculateDigest,
		getPiecesMaxRetry: getPiecesMaxRetry,
	}
	return ptm, nil
}

var _ TaskManager = (*peerTaskManager)(nil)

func (ptm *peerTaskManager) findPeerTaskConductor(taskID string) (*peerTaskConductor, bool) {
	pt, ok := ptm.runningPeerTasks.Load(taskID)
	if !ok {
		return nil, false
	}
	return pt.(*peerTaskConductor), true
}

func (ptm *peerTaskManager) getPeerTaskConductor(ctx context.Context,
	taskID string,
	request *scheduler.PeerTaskRequest,
	limit rate.Limit) (*peerTaskConductor, error) {
	ptc, created, err := ptm.getOrCreatePeerTaskConductor(ctx, taskID, request, limit)
	if err != nil {
		return nil, err
	}
	if created {
		if err = ptc.start(); err != nil {
			return nil, err
		}
	}
	return ptc, err
}

// getOrCreatePeerTaskConductor will get or create a peerTaskConductor,
// if created, return (ptc, true, nil), otherwise return (ptc, false, nil)
func (ptm *peerTaskManager) getOrCreatePeerTaskConductor(
	ctx context.Context,
	taskID string,
	request *scheduler.PeerTaskRequest,
	limit rate.Limit) (*peerTaskConductor, bool, error) {
	if ptc, ok := ptm.findPeerTaskConductor(taskID); ok {
		logger.Debugf("peer task found: %s/%s", ptc.taskID, ptc.peerID)
		return ptc, false, nil
	}
	ptc := ptm.newPeerTaskConductor(ctx, request, limit)

	ptm.conductorLock.Lock()
	// double check
	if p, ok := ptm.findPeerTaskConductor(taskID); ok {
		ptm.conductorLock.Unlock()
		logger.Debugf("peer task found: %s/%s", p.taskID, p.peerID)
		metrics.PeerTaskCacheHitCount.Add(1)
		return p, false, nil
	}
	ptm.runningPeerTasks.Store(taskID, ptc)
	ptm.conductorLock.Unlock()
	metrics.PeerTaskCount.Add(1)
	return ptc, true, nil
}

func (ptm *peerTaskManager) prefetch(request *scheduler.PeerTaskRequest) {
	req := &scheduler.PeerTaskRequest{
		Url:         request.Url,
		PeerId:      request.PeerId,
		PeerHost:    ptm.host,
		HostLoad:    request.HostLoad,
		IsMigrating: request.IsMigrating,
		UrlMeta: &base.UrlMeta{
			Digest: request.UrlMeta.Digest,
			Tag:    request.UrlMeta.Tag,
			Filter: request.UrlMeta.Filter,
			Header: map[string]string{},
		},
	}
	for k, v := range request.UrlMeta.Header {
		if k == headers.Range {
			continue
		}
		req.UrlMeta.Header[k] = v
	}
	taskID := idgen.TaskID(req.Url, req.UrlMeta)
	req.PeerId = idgen.PeerID(req.PeerHost.Ip)

	var limit = rate.Inf
	if ptm.perPeerRateLimit > 0 {
		limit = ptm.perPeerRateLimit
	}

	logger.Infof("prefetch peer task %s/%s", taskID, req.PeerId)
	prefetch, err := ptm.getPeerTaskConductor(context.Background(), taskID, req, limit)
	if err != nil {
		logger.Errorf("prefetch peer task %s/%s error: %s", prefetch.taskID, prefetch.peerID, err)
	}
	if prefetch != nil && prefetch.peerID == req.PeerId {
		metrics.PrefetchTaskCount.Add(1)
	}
}

func (ptm *peerTaskManager) StartFileTask(ctx context.Context, req *FileTaskRequest) (chan *FileTaskProgress, *TinyData, error) {
	if ptm.enableMultiplex {
		progress, ok := ptm.tryReuseFilePeerTask(ctx, req)
		if ok {
			metrics.PeerTaskCacheHitCount.Add(1)
			return progress, nil, nil
		}
	}
	// TODO ensure scheduler is ok first
	var limit = rate.Inf
	if ptm.perPeerRateLimit > 0 {
		limit = ptm.perPeerRateLimit
	}
	if req.Limit > 0 {
		limit = rate.Limit(req.Limit)
	}
	ctx, pt, err := ptm.newFileTask(ctx, req, limit)
	if err != nil {
		return nil, nil, err
	}

	// FIXME when failed due to schedulerClient error, relocate schedulerClient and retry
	progress, err := pt.Start(ctx)
	return progress, nil, err
}

func (ptm *peerTaskManager) StartStreamTask(ctx context.Context, req *StreamTaskRequest) (io.ReadCloser, map[string]string, error) {
	peerTaskRequest := &scheduler.PeerTaskRequest{
		Url:         req.URL,
		UrlMeta:     req.URLMeta,
		PeerId:      req.PeerID,
		PeerHost:    ptm.host,
		HostLoad:    nil,
		IsMigrating: false,
	}

	if ptm.enableMultiplex {
		r, attr, ok := ptm.tryReuseStreamPeerTask(ctx, req)
		if ok {
			metrics.PeerTaskCacheHitCount.Add(1)
			return r, attr, nil
		}
	}

	pt, err := ptm.newStreamTask(ctx, peerTaskRequest)
	if err != nil {
		return nil, nil, err
	}

	// FIXME when failed due to schedulerClient error, relocate schedulerClient and retry
	readCloser, attribute, err := pt.Start(ctx)
	return readCloser, attribute, err
}

func (ptm *peerTaskManager) Stop(ctx context.Context) error {
	// TODO
	return nil
}

func (ptm *peerTaskManager) PeerTaskDone(taskID string) {
	logger.Debugf("delete done task %s in running tasks", taskID)
	ptm.runningPeerTasks.Delete(taskID)
}

func (ptm *peerTaskManager) IsPeerTaskRunning(taskID string) bool {
	_, ok := ptm.runningPeerTasks.Load(taskID)
	return ok
}

func (ptm *peerTaskManager) StatPeerTask(ctx context.Context, taskID string) (*base.GrpcDfResult, error) {
	req := &scheduler.StatPeerTaskRequest{
		TaskId: taskID,
	}

	return ptm.schedulerClient.StatPeerTask(ctx, req)
}

func (ptm *peerTaskManager) GetPieceManager() PieceManager {
	return ptm.pieceManager
}

func (ptm *peerTaskManager) AnnouncePeerTask(ctx context.Context,
	meta storage.PeerTaskMetadata, cid string, urlMeta *base.UrlMeta) error {
	log := logger.With("function", "AnnouncePeerTask", "taskID", meta.TaskID, "peerID", meta.PeerID, "CID", cid)

	// Check if the given task is completed in local storageManager
	if ptm.storageManager.FindCompletedTask(meta.TaskID) == nil {
		msg := fmt.Sprintf("task %s not found in local storage", meta.TaskID)
		log.Errorf(msg)
		return errors.New(msg)
	}

	// prepare AnnounceTaskRequest
	totalPieces, err := ptm.storageManager.GetTotalPieces(ctx, &meta)
	if err != nil {
		msg := fmt.Sprintf("get total pieces failed: %s", err)
		log.Error(msg)
		return errors.New(msg)
	}
	pieceTaskRequest := &base.PieceTaskRequest{
		TaskId:   meta.TaskID,
		DstPid:   meta.PeerID,
		StartNum: 0,
		Limit:    uint32(totalPieces),
	}
	piecePacket, err := ptm.storageManager.GetPieces(ctx, pieceTaskRequest)
	if err != nil {
		msg := fmt.Sprintf("get pieces info failed: %s", err)
		log.Error(msg)
		return errors.New(msg)
	}
	piecePacket.DstAddr = fmt.Sprintf("%s:%d", ptm.host.Ip, ptm.host.DownPort)
	req := &scheduler.AnnounceTaskRequest{
		TaskId:      meta.TaskID,
		Cid:         cid,
		UrlMeta:     urlMeta,
		PeerHost:    ptm.host,
		PiecePacket: piecePacket,
	}

	// Announce peer task to scheduler
	res, err := ptm.schedulerClient.AnnounceTask(ctx, req)
	if err != nil {
		msg := fmt.Sprintf("announce peer task failed: %s", err)
		log.Error(msg)
		return errors.New(msg)
	}
	if res.Code != base.Code_Success {
		msg := fmt.Sprintf("announce peer task got unexpected result: %v", res)
		log.Error(msg)
		return errors.New(msg)
	}
	return nil
}
