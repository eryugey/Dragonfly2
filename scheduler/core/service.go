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

package core

import (
	"context"
	"sync"
	"time"

	"github.com/pkg/errors"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/util/workqueue"

	"d7y.io/dragonfly/v2/internal/dferrors"
	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/pkg/gc"
	"d7y.io/dragonfly/v2/pkg/rpc/base"
	"d7y.io/dragonfly/v2/pkg/rpc/base/common"
	schedulerRPC "d7y.io/dragonfly/v2/pkg/rpc/scheduler"
	pkgsync "d7y.io/dragonfly/v2/pkg/sync"
	"d7y.io/dragonfly/v2/scheduler/config"
	"d7y.io/dragonfly/v2/scheduler/core/scheduler"
	"d7y.io/dragonfly/v2/scheduler/metrics"
	"d7y.io/dragonfly/v2/scheduler/supervisor"
)

const maxRescheduleTimes = 8

type Options struct {
	openTel    bool
	disableCDN bool
}

type Option func(options *Options)

func WithOpenTel(openTel bool) Option {
	return func(options *Options) {
		options.openTel = openTel
	}
}

func WithDisableCDN(disableCDN bool) Option {
	return func(options *Options) {
		options.disableCDN = disableCDN
	}
}

type SchedulerService struct {
	// CDN manager
	CDN supervisor.CDN
	// task manager
	taskManager supervisor.TaskManager
	// host manager
	hostManager supervisor.HostManager
	// Peer manager
	peerManager supervisor.PeerManager

	sched   scheduler.Scheduler
	worker  worker
	monitor *monitor
	done    chan struct{}
	wg      sync.WaitGroup
	kmu     *pkgsync.Krwmutex

	config        *config.SchedulerConfig
	dynconfig     config.DynconfigInterface
	metricsConfig *config.MetricsConfig
}

func NewSchedulerService(cfg *config.SchedulerConfig, pluginDir string, metricsConfig *config.MetricsConfig, dynConfig config.DynconfigInterface, gc gc.GC, options ...Option) (*SchedulerService, error) {
	ops := &Options{}
	for _, op := range options {
		op(ops)
	}

	hostManager := supervisor.NewHostManager()

	peerManager, err := supervisor.NewPeerManager(cfg.GC, gc, hostManager)
	if err != nil {
		return nil, err
	}

	taskManager, err := supervisor.NewTaskManager(cfg.GC, gc, peerManager)
	if err != nil {
		return nil, err
	}

	sched, err := scheduler.Get(cfg.Scheduler).Build(cfg, &scheduler.BuildOptions{
		PeerManager: peerManager,
		PluginDir:   pluginDir,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "build scheduler %v", cfg.Scheduler)
	}

	work := newEventLoopGroup(cfg.WorkerNum)
	downloadMonitor := newMonitor(cfg.OpenMonitor, peerManager)
	s := &SchedulerService{
		taskManager:   taskManager,
		hostManager:   hostManager,
		peerManager:   peerManager,
		worker:        work,
		monitor:       downloadMonitor,
		sched:         sched,
		config:        cfg,
		metricsConfig: metricsConfig,
		dynconfig:     dynConfig,
		done:          make(chan struct{}),
		wg:            sync.WaitGroup{},
		kmu:           pkgsync.NewKrwmutex(),
	}
	if !ops.disableCDN {
		var opts []grpc.DialOption
		if ops.openTel {
			opts = append(opts, grpc.WithChainUnaryInterceptor(otelgrpc.UnaryClientInterceptor()), grpc.WithChainStreamInterceptor(otelgrpc.StreamClientInterceptor()))
		}
		client, err := supervisor.NewCDNDynmaicClient(dynConfig, peerManager, hostManager, opts)
		if err != nil {
			return nil, errors.Wrap(err, "new refreshable cdn client")
		}

		cdn := supervisor.NewCDN(client, peerManager, hostManager)
		if err != nil {
			return nil, errors.Wrap(err, "new cdn manager")
		}
		s.CDN = cdn
	}
	return s, nil
}

func (s *SchedulerService) Serve() {
	s.wg.Add(2)
	wsdq := workqueue.NewNamedDelayingQueue("wait reSchedule parent")
	go s.runWorkerLoop(wsdq)
	go s.runReScheduleParentLoop(wsdq)
	go s.runMonitor()
	logger.Debugf("start scheduler service successfully")
}

func (s *SchedulerService) runWorkerLoop(wsdq workqueue.DelayingInterface) {
	defer s.wg.Done()
	s.worker.start(newState(s.sched, s.peerManager, s.CDN, wsdq))
}

func (s *SchedulerService) runReScheduleParentLoop(wsdq workqueue.DelayingInterface) {
	for {
		select {
		case <-s.done:
			wsdq.ShutDown()
			return
		default:
			v, shutdown := wsdq.Get()
			if shutdown {
				logger.Infof("wait schedule delay queue is shutdown")
				break
			}
			rsPeer := v.(*rsPeer)
			peer := rsPeer.peer
			wsdq.Done(v)
			if rsPeer.times > maxRescheduleTimes {
				if peer.CloseChannelWithError(dferrors.Newf(base.Code_SchedNeedBackSource, "reschedule parent for peer %s already reaches max reschedule times",
					peer.ID)) == nil {
					peer.Task.AddBackToSourcePeer(peer.ID)
				}
				continue
			}
			if peer.Task.ContainsBackToSourcePeer(peer.ID) {
				logger.WithTaskAndPeerID(peer.Task.ID, peer.ID).Debugf("runReScheduleLoop: peer is back source client, no need to reschedule it")
				continue
			}
			if peer.IsDone() || peer.IsLeave() {
				peer.Log().Debugf("runReScheduleLoop: peer has left from waitScheduleParentPeerQueue because peer is done or leave, peer status is %s, "+
					"isLeave %t", peer.GetStatus(), peer.IsLeave())
				continue
			}
			s.worker.send(reScheduleParentEvent{rsPeer: rsPeer})
		}
	}
}

func (s *SchedulerService) runMonitor() {
	defer s.wg.Done()
	if s.monitor != nil {
		s.monitor.start(s.done)
	}
}

func (s *SchedulerService) Stop() {
	close(s.done)
	if s.worker != nil {
		s.worker.stop()
	}
	s.wg.Wait()
}

func (s *SchedulerService) SelectParent(peer *supervisor.Peer) (parent *supervisor.Peer, err error) {
	parent, _, hasParent := s.sched.ScheduleParent(peer, sets.NewString())
	if !hasParent || parent == nil {
		return nil, errors.Errorf("no parent peer available for peer %s", peer.ID)
	}
	return parent, nil
}

func (s *SchedulerService) GetPeer(id string) (*supervisor.Peer, bool) {
	return s.peerManager.Get(id)
}

func (s *SchedulerService) RegisterTask(req *schedulerRPC.PeerTaskRequest, task *supervisor.Task) *supervisor.Peer {
	// get or create host
	peerHost := req.PeerHost
	host, ok := s.hostManager.Get(peerHost.Uuid)
	if !ok {
		var options []supervisor.HostOption
		if clientConfig, ok := s.dynconfig.GetSchedulerClusterClientConfig(); ok {
			options = []supervisor.HostOption{
				supervisor.WithTotalUploadLoad(clientConfig.LoadLimit),
			}
		}

		host = supervisor.NewClientHost(peerHost.Uuid, peerHost.Ip, peerHost.HostName, peerHost.RpcPort, peerHost.DownPort,
			peerHost.SecurityDomain, peerHost.Location, peerHost.Idc, options...)
		s.hostManager.Add(host)
	}
	// get or creat PeerTask
	peer, ok := s.peerManager.Get(req.PeerId)
	if ok {
		logger.Warnf("peer %s has already registered", peer.ID)
		return peer
	}
	peer = supervisor.NewPeer(req.PeerId, task, host)
	s.peerManager.Add(peer)
	return peer
}

func (s *SchedulerService) GetOrAddTask(ctx context.Context, task *supervisor.Task, noTrigger bool) *supervisor.Task {
	span := trace.SpanFromContext(ctx)

	s.kmu.RLock(task.ID)
	task, ok := s.taskManager.GetOrAdd(task)
	if ok {
		span.SetAttributes(config.AttributeTaskStatus.String(task.GetStatus().String()))
		span.SetAttributes(config.AttributeLastTriggerTime.String(task.LastTriggerAt.Load().String()))
		if task.LastTriggerAt.Load().Add(s.config.AccessWindow).After(time.Now()) || task.IsHealth() {
			span.SetAttributes(config.AttributeNeedSeedCDN.Bool(false))
			s.kmu.RUnlock(task.ID)
			return task
		}
	} else {
		task.Log().Infof("add new task %s", task.ID)
	}
	s.kmu.RUnlock(task.ID)

	if noTrigger {
		return task
	}

	s.kmu.Lock(task.ID)
	defer s.kmu.Unlock(task.ID)

	// do trigger
	span.SetAttributes(config.AttributeTaskStatus.String(task.GetStatus().String()))
	span.SetAttributes(config.AttributeLastTriggerTime.String(task.LastTriggerAt.Load().String()))
	if task.IsHealth() {
		span.SetAttributes(config.AttributeNeedSeedCDN.Bool(false))
		return task
	}

	task.LastTriggerAt.Store(time.Now())
	task.SetStatus(supervisor.TaskStatusRunning)
	if s.CDN == nil {
		// client back source
		span.SetAttributes(config.AttributeClientBackSource.Bool(true))
		task.BackToSourceWeight.Store(s.config.BackSourceCount)
		return task
	}
	span.SetAttributes(config.AttributeNeedSeedCDN.Bool(true))

	go func() {
		if cdnPeer, err := s.CDN.StartSeedTask(ctx, task); err != nil {
			// fall back to client back source
			task.Log().Errorf("seed task failed: %v", err)
			span.AddEvent(config.EventCDNFailBackClientSource, trace.WithAttributes(config.AttributeTriggerCDNError.String(err.Error())))
			task.BackToSourceWeight.Store(s.config.BackSourceCount)
			if ok = s.worker.send(taskSeedFailEvent{task}); !ok {
				logger.Error("send taskSeed fail event failed, eventLoop is shutdown")
			}
		} else {
			if ok = s.worker.send(peerDownloadSuccessEvent{cdnPeer, nil}); !ok {
				logger.Error("send taskSeed success event failed, eventLoop is shutdown")
			}
			logger.Infof("successfully obtain seeds from cdn, task: %#v", task)
		}
	}()

	return task
}

func (s *SchedulerService) GetTask(ctx context.Context, taskID string) *supervisor.Task {
	span := trace.SpanFromContext(ctx)

	s.kmu.RLock(taskID)
	defer s.kmu.RUnlock(taskID)
	task, ok := s.taskManager.Get(taskID)
	if ok {
		span.SetAttributes(config.AttributeTaskStatus.String(task.GetStatus().String()))
		return task
	} else {
		return nil
	}
}

func (s *SchedulerService) HandlePieceResult(ctx context.Context, peer *supervisor.Peer, pieceResult *schedulerRPC.PieceResult) error {
	peer.Touch()
	if pieceResult.Success && s.metricsConfig != nil && s.metricsConfig.EnablePeerHost {
		// TODO parse PieceStyle
		metrics.PeerHostTraffic.WithLabelValues("download", peer.Host.UUID, peer.Host.IP).Add(float64(pieceResult.PieceInfo.RangeSize))
		if p, ok := s.peerManager.Get(pieceResult.DstPid); ok {
			metrics.PeerHostTraffic.WithLabelValues("upload", p.Host.UUID, p.Host.IP).Add(float64(pieceResult.PieceInfo.RangeSize))
		} else {
			logger.Warnf("dst peer %s not found for pieceResult %#v, pieceInfo %#v", pieceResult.DstPid, pieceResult, pieceResult.PieceInfo)
		}
	}
	if pieceResult.PieceInfo != nil && pieceResult.PieceInfo.PieceNum == common.EndOfPiece {
		return nil
	} else if pieceResult.PieceInfo != nil && pieceResult.PieceInfo.PieceNum == common.ZeroOfPiece {
		s.worker.send(startReportPieceResultEvent{ctx, peer})
		return nil
	} else if pieceResult.PieceInfo != nil && pieceResult.PieceInfo.PieceNum == common.QueryOfPiece {
		s.worker.send(peerQueryPieceResultEvent{ctx, peer})
		return nil
	} else if pieceResult.Success {
		s.worker.send(peerDownloadPieceSuccessEvent{
			ctx:  ctx,
			peer: peer,
			pr:   pieceResult,
		})
		return nil
	} else if pieceResult.Code != base.Code_Success {
		s.worker.send(peerDownloadPieceFailEvent{
			ctx:  ctx,
			peer: peer,
			pr:   pieceResult,
		})
		return nil
	}
	return nil
}

func (s *SchedulerService) HandlePeerResult(ctx context.Context, peer *supervisor.Peer, peerResult *schedulerRPC.PeerResult) error {
	peer.Touch()
	if peerResult.Success {
		if !s.worker.send(peerDownloadSuccessEvent{peer: peer, peerResult: peerResult}) {
			logger.Errorf("send peer download success event failed")
		}
	} else if !s.worker.send(peerDownloadFailEvent{peer: peer, peerResult: peerResult}) {
		logger.Errorf("send peer download fail event failed")
	}
	return nil
}

func (s *SchedulerService) HandleLeaveTask(ctx context.Context, peer *supervisor.Peer) error {
	peer.Touch()
	if !s.worker.send(peerLeaveEvent{
		ctx:  ctx,
		peer: peer,
	}) {
		logger.Errorf("send peer leave event failed")
	}
	return nil
}
