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
	"time"

	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/pkg/rpc/dfdaemon"
	"d7y.io/dragonfly/v2/pkg/rpc/dfproxy"
)

type Service struct {
	reqCh chan dfproxy.DaemonProxyServerPacket
	resCh chan dfproxy.DaemonProxyClientPacket
}

// New service instance
func New() *Service {
	return &Service{
		reqCh: make(chan dfproxy.DaemonProxyServerPacket),
		resCh: make(chan dfproxy.DaemonProxyClientPacket),
	}
}

// Dfdaemon handles the proxied dfdaemon requests.
func (s *Service) Dfdaemon(stream dfproxy.DaemonProxy_DfdaemonServer) error {
	defer func() {
		close(s.reqCh)
		close(s.resCh)
	}()

	ctx := stream.Context()
	select {
	case <-ctx.Done():
		logger.Infof("Dfproxy Dfdaemon context was done")
		return ctx.Err()
	}

	// Dfdaemon receives the first packet, which should be ReqType_HeartBeat, subsequent packets
	// are handled by req type handlers accordingly.
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

	// We got initial HeartBeat request, and we're ready to send server packet back to client as
	// dfdaemon requests.
	for {
		select {
		case <-ctx.Done():
			logger.Infof("Dfproxy Dfdaemon context done: %s", ctx.Err())
			return ctx.Err()
		case serverPkt := <-s.reqCh:
			if err := s.handleServerPacket(stream, serverPkt); err != nil {
				logger.Warnf("Failed to handle server req: %s", err.Error())
			}
		}
	}
}

func (s *Service) handleServerPacket(stream dfproxy.DaemonProxy_DfdaemonServer, serverPkt dfproxy.DaemonProxyServerPacket) error {
	if err := stream.Send(&serverPkt); err != nil {
		msg := fmt.Sprintf("failed to send server packet: %s", err.Error())
		logger.Error(msg)
		return errors.New(msg)
	}

	clientPkt, err := stream.Recv()
	if err != nil {
		if err == io.EOF {
			msg := ("Dfproxy Dfdaemon service received unexpected EOF")
			logger.Error(msg)
			return errors.New(msg)
		}
		msg := fmt.Sprintf("Dfproxy Dfdaemon service received error: %s", err.Error())
		logger.Error(msg)
		return errors.New(msg)
	}
	if err := checkReqType(clientPkt.GetType(), serverPkt.GetType()); err != nil {
		return err
	}

	s.resCh <- *clientPkt
	return nil
}

func (s *Service) CheckHealth(ctx context.Context) error {
	reqType := dfproxy.ReqType_CheckHealth
	serverPkt := dfproxy.DaemonProxyServerPacket{
		Type: reqType,
	}

	logger.Info("dfproxy starts to check health")

	start := time.Now()
	s.reqCh <- serverPkt
	select {
	case clientPkt := <-s.resCh:
		if err := handleClientPacket(clientPkt, reqType); err != nil {
			return err
		}
		logger.Infof("check health successfully in %.6f s", time.Since(start).Seconds())
		return nil
	case <-ctx.Done():
		return handleContextDone(ctx, "check health timeout")
	}
}

func (s *Service) StatTask(ctx context.Context, req *dfdaemon.StatTaskRequest) error {
	reqType := dfproxy.ReqType_StatTask
	serverPkt := dfproxy.DaemonProxyServerPacket{
		Type: dfproxy.ReqType_StatTask,
		DaemonReq: &dfproxy.DfDaemonReq{
			StatTask: req,
		},
	}

	wLog := logger.With("Cid", req.Cid, "Tag", req.UrlMeta.Tag)
	wLog.Info("dfproxy starts to stat task")

	start := time.Now()
	s.reqCh <- serverPkt
	select {
	case clientPkt := <-s.resCh:
		if err := handleClientPacket(clientPkt, reqType); err != nil {
			return err
		}
		wLog.Infof("task stat successfully in %.6f s", time.Since(start).Seconds())
		return nil
	case <-ctx.Done():
		return handleContextDone(ctx, "stat timeout")
	}
}

func (s *Service) ImportTask(ctx context.Context, req *dfdaemon.ImportTaskRequest) error {
	reqType := dfproxy.ReqType_ImportTask
	serverPkt := dfproxy.DaemonProxyServerPacket{
		Type: dfproxy.ReqType_ImportTask,
		DaemonReq: &dfproxy.DfDaemonReq{
			ImportTask: req,
		},
	}

	wLog := logger.With("Cid", req.Cid, "Tag", req.UrlMeta.Tag, "file", req.Path)
	wLog.Info("dfproxy starts to import task")

	start := time.Now()
	s.reqCh <- serverPkt
	select {
	case clientPkt := <-s.resCh:
		if err := handleClientPacket(clientPkt, reqType); err != nil {
			return err
		}
		wLog.Infof("task imported successfully in %.6f s", time.Since(start).Seconds())
		return nil
	case <-ctx.Done():
		return handleContextDone(ctx, "import timeout")
	}
}

func (s *Service) ExportTask(ctx context.Context, req *dfdaemon.ExportTaskRequest) error {
	reqType := dfproxy.ReqType_ExportTask
	serverPkt := dfproxy.DaemonProxyServerPacket{
		Type: dfproxy.ReqType_ExportTask,
		DaemonReq: &dfproxy.DfDaemonReq{
			ExportTask: req,
		},
	}

	wLog := logger.With("Cid", req.Cid, "Tag", req.UrlMeta.Tag, "output", req.Output)
	wLog.Info("dfproxy starts to export task")

	start := time.Now()
	s.reqCh <- serverPkt
	select {
	case clientPkt := <-s.resCh:
		if err := handleClientPacket(clientPkt, reqType); err != nil {
			return err
		}
		wLog.Infof("task exported successfully in %.6f s", time.Since(start).Seconds())
		return nil
	case <-ctx.Done():
		return handleContextDone(ctx, "export timeout")
	}
}

func (s *Service) DeleteTask(ctx context.Context, req *dfdaemon.DeleteTaskRequest) error {
	reqType := dfproxy.ReqType_DeleteTask
	serverPkt := dfproxy.DaemonProxyServerPacket{
		Type: dfproxy.ReqType_DeleteTask,
		DaemonReq: &dfproxy.DfDaemonReq{
			DeleteTask: req,
		},
	}

	wLog := logger.With("Cid", req.Cid, "Tag", req.UrlMeta.Tag)
	wLog.Info("dfproxy starts to delete task")

	start := time.Now()
	s.reqCh <- serverPkt
	select {
	case clientPkt := <-s.resCh:
		if err := handleClientPacket(clientPkt, reqType); err != nil {
			return err
		}
		wLog.Infof("task deleted successfully in %.6f s", time.Since(start).Seconds())
		return nil
	case <-ctx.Done():
		return handleContextDone(ctx, "delete timeout")
	}
}

func checkReqType(got, expected dfproxy.ReqType) error {
	if got != expected {
		msg := fmt.Sprintf("invalid req type %d %s, expected %d %s", got, dfproxy.ReqType_name[int32(got)], expected, dfproxy.ReqType_name[int32(expected)])
		logger.Error(msg)
		return errors.New(msg)
	}
	return nil
}

func handleClientPacket(clientPkt dfproxy.DaemonProxyClientPacket, reqType dfproxy.ReqType) error {
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
