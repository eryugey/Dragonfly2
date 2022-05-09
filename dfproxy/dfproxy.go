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

package dfproxy

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"time"

	"d7y.io/dragonfly/v2/dfproxy/config"
	"d7y.io/dragonfly/v2/dfproxy/rpcserver"
	"d7y.io/dragonfly/v2/dfproxy/service"
	"d7y.io/dragonfly/v2/internal/dferrors"
	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/pkg/dfnet"
	"d7y.io/dragonfly/v2/pkg/dfpath"
	"d7y.io/dragonfly/v2/pkg/rpc/base"
	"d7y.io/dragonfly/v2/pkg/rpc/base/common"
	daemonclient "d7y.io/dragonfly/v2/pkg/rpc/dfdaemon/client"
	"d7y.io/dragonfly/v2/pkg/rpc/dfproxy"
	"d7y.io/dragonfly/v2/pkg/rpc/dfproxy/client"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

const (
	// gracefulStopTimeout specifies a time limit for
	// grpc server to complete a graceful shutdown
	gracefulStopTimeout = 5 * time.Second
)

type Client struct {
	config       *config.Config
	proxyClient  dfproxy.DaemonProxyClient
	daemonClient daemonclient.DaemonClient
}

func NewClient(ctx context.Context, cfg *config.Config, d dfpath.Dfpath) (*Client, error) {
	proxyClient, err := client.GetClientByAddr([]dfnet.NetAddr{*cfg.Server.ProxyGRPC})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dfproxy client to addr %s", cfg.Server.ProxyGRPC)
	}

	daemonClient, err := daemonclient.GetClientByAddr([]dfnet.NetAddr{*cfg.Daemon.DaemonGRPC})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get dfdaemon client to addr %s", cfg.Daemon.DaemonGRPC)
	}

	return &Client{
		config:       cfg,
		proxyClient:  proxyClient,
		daemonClient: daemonClient,
	}, nil
}

func (c *Client) Do() error {
	stream, err := c.proxyClient.Dfdaemon(context.Background())
	if err != nil {
		msg := fmt.Sprintf("failed to call Dfdaemon: %s", err.Error())
		logger.Error(msg)
		return errors.New(msg)
	}

	logger.Info("dfproxy client got stream to server")

	// Send initial HeartBeat request to initialize stream connection, server won't send back a
	// reply, any subsequent packets sent by server are actual proxied dfdaemon requests.
	hbPkt := &dfproxy.DaemonProxyClientPacket{
		Type: dfproxy.ReqType_HeartBeat,
	}
	if err := stream.Send(hbPkt); err != nil {
		msg := fmt.Sprintf("failed to send initial heart beat packet: %s", err.Error())
		logger.Error(msg)
		return errors.New(msg)
	}

	// Handle proxied dfdaemon requests from dfproxy server
	ctx := stream.Context()
	for {
		select {
		case <-ctx.Done():
			logger.Info("dfproxy stream context done: %s", ctx.Err())
			return ctx.Err()
		default:
		}

		// Receive requests from dfproxy server
		serverPkt, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				logger.Info("dfproxy received EOF")
				return nil
			}
			msg := fmt.Sprintf("dfproxy received error: %s", err.Error())
			logger.Error(msg)
			return errors.New(msg)
		}

		reqType := serverPkt.GetType()
		switch reqType {
		case dfproxy.ReqType_HeartBeat:
			logger.Warn("unexpected heart beat request from dfproxy server")
			if err := stream.Send(hbPkt); err != nil {
				logger.Errorf("failed to send heart beat packet: %s", err.Error())
			}
		default:
			if err := c.handleServerPacket(stream, serverPkt); err != nil {
				logger.Errorf("failed to handle server packet: %s", err.Error())
			}
		}
	}

	return nil
}

func (c *Client) handleServerPacket(stream dfproxy.DaemonProxy_DfdaemonClient, serverPkt *dfproxy.DaemonProxyServerPacket) error {
	reqType := serverPkt.GetType()
	reqTypeName := dfproxy.ReqType_name[int32(reqType)]
	logger.Infof("dfproxy handling new server packet %s", reqTypeName)
	if serverPkt.DaemonReq == nil && reqType != dfproxy.ReqType_CheckHealth {
		return errors.New("invalid dfproxy server packet, DfDaemonReq is nil")
	}

	clientPkt := dfproxy.DaemonProxyClientPacket{
		Type:  reqType,
		Error: nil,
	}
	switch reqType {
	case dfproxy.ReqType_CheckHealth:
		clientPkt.Error = handleError(c.daemonClient.CheckHealth(stream.Context(), *c.config.Daemon.DaemonGRPC))
	case dfproxy.ReqType_StatTask:
		statReq := serverPkt.DaemonReq.StatTask
		logger.Infof("dfproxy redirecting server req %v", *statReq)
		if statReq == nil {
			e := common.NewGrpcDfError(base.Code_BadRequest, "StatTaskRequest is empty")
			clientPkt.Error = e
		} else {
			clientPkt.Error = handleError(c.daemonClient.StatTask(stream.Context(), statReq))
		}
	case dfproxy.ReqType_ImportTask:
		importReq := serverPkt.DaemonReq.ImportTask
		logger.Infof("dfproxy redirecting server req %v", *importReq)
		if importReq == nil {
			e := common.NewGrpcDfError(base.Code_BadRequest, "ImportTaskRequest is empty")
			clientPkt.Error = e
		} else {
			clientPkt.Error = handleError(c.daemonClient.ImportTask(stream.Context(), importReq))
		}
	case dfproxy.ReqType_ExportTask:
		exportReq := serverPkt.DaemonReq.ExportTask
		logger.Infof("dfproxy redirecting server req %v", *exportReq)
		if exportReq == nil {
			e := common.NewGrpcDfError(base.Code_BadRequest, "ExportTaskRequest is empty")
			clientPkt.Error = e
		} else {
			clientPkt.Error = handleError(c.daemonClient.ExportTask(stream.Context(), exportReq))
		}
	case dfproxy.ReqType_DeleteTask:
		deleteReq := serverPkt.DaemonReq.DeleteTask
		logger.Infof("dfproxy redirecting server req %v", *deleteReq)
		if deleteReq == nil {
			e := common.NewGrpcDfError(base.Code_BadRequest, "DeleteTaskRequest is empty")
			clientPkt.Error = e
		} else {
			err := c.daemonClient.DeleteTask(stream.Context(), deleteReq)
			clientPkt.Error = handleError(err)
		}
	default:
		msg := fmt.Sprintf("Unknown request type %d", reqType)
		e := common.NewGrpcDfError(base.Code_BadRequest, msg)
		clientPkt.Error = e
	}

	logger.Infof("dfproxy sending back client packet: type %s, error %s", reqTypeName, clientPkt.Error)
	if err := stream.Send(&clientPkt); err != nil {
		msg := fmt.Sprintf("failed to send client packet: %s", err.Error())
		logger.Error(msg)
		return errors.New(msg)
	}
	return nil
}

func handleError(retErr error) *base.GrpcDfError {
	if retErr == nil {
		return nil
	}
	if e, ok := retErr.(*dferrors.DfError); ok {
		return common.NewGrpcDfError(e.Code, e.Message)
	}
	msg := fmt.Sprintf("error msg: %s", retErr.Error())
	return common.NewGrpcDfError(base.Code_UnknownError, msg)
}

type Server struct {
	grpcServer *grpc.Server
	listen     *dfnet.NetAddr
	service    *service.Service
}

func NewServer(listen *dfnet.NetAddr, ready chan<- bool) (*Server, error) {
	// Initialize dfproxy service
	service := service.New(ready)

	// Initialize grpc service
	svr := rpcserver.New(service)

	return &Server{
		grpcServer: svr,
		listen:     listen,
		service:    service,
	}, nil
}

func (s *Server) GetService() *service.Service {
	return s.service
}

func (s *Server) Serve() error {
	if s.listen.Type == dfnet.UNIX {
		if !filepath.IsAbs(s.listen.Addr) {
			return fmt.Errorf("dfdaemon socket path is not absolute path: %s", s.listen.Addr)
		}
		dir, _ := path.Split(s.listen.Addr)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.Wrapf(err, "failed to mkdir %s", dir)
		}
		_ = os.Remove(s.listen.Addr)
	}
	listener, err := net.Listen(string(s.listen.Type), s.listen.Addr)
	if err != nil {
		msg := fmt.Sprintf("failed to listen on %s: %s", s.listen.GetEndpoint(), err.Error())
		logger.Error(msg)
		return errors.New(msg)
	}
	defer listener.Close()

	// Start GRPC server
	logger.Infof("dfproxy starting grpc server at %s", s.listen.GetEndpoint())
	if err := s.grpcServer.Serve(listener); err != nil {
		msg := fmt.Sprintf("dfproxy serve failed: %s", err.Error())
		logger.Error(msg)
		return errors.New(msg)
	}
	return nil
}

func (s *Server) Stop() {
	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		logger.Info("dfproxy grpc server closed under request")
		close(stopped)
	}()

	t := time.NewTimer(gracefulStopTimeout)
	select {
	case <-t.C:
		s.grpcServer.Stop()
	case <-stopped:
		t.Stop()
	}
}
