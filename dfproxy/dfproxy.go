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

	"d7y.io/dragonfly/v2/dfproxy/config"
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
)

type Client struct {
	config       *config.Config
	proxyClient  dfproxy.DaemonProxyClient
	daemonClient daemonclient.DaemonClient
}

func New(ctx context.Context, cfg *config.Config, d dfpath.Dfpath) (*Client, error) {
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
	if serverPkt.DaemonReq == nil {
		return errors.New("invalid dfproxy server packet, DfDaemonReq is nil")
	}

	reqType := serverPkt.GetType()
	clientPkt := dfproxy.DaemonProxyClientPacket{
		Type:  reqType,
		Error: nil,
	}
	switch reqType {
	case dfproxy.ReqType_StatTask:
		statReq := serverPkt.DaemonReq.StatTask
		if statReq == nil {
			e := common.NewGrpcDfError(base.Code_BadRequest, "StatTaskRequest is empty")
			clientPkt.Error = e
		} else {
			clientPkt.Error = handleError(c.daemonClient.StatTask(stream.Context(), statReq))
		}
	case dfproxy.ReqType_ImportTask:
		importReq := serverPkt.DaemonReq.ImportTask
		if importReq == nil {
			e := common.NewGrpcDfError(base.Code_BadRequest, "ImportTaskRequest is empty")
			clientPkt.Error = e
		} else {
			clientPkt.Error = handleError(c.daemonClient.ImportTask(stream.Context(), importReq))
		}
	case dfproxy.ReqType_ExportTask:
		exportReq := serverPkt.DaemonReq.ExportTask
		if exportReq == nil {
			e := common.NewGrpcDfError(base.Code_BadRequest, "ExportTaskRequest is empty")
			clientPkt.Error = e
		} else {
			clientPkt.Error = handleError(c.daemonClient.ExportTask(stream.Context(), exportReq))
		}
	case dfproxy.ReqType_DeleteTask:
		deleteReq := serverPkt.DaemonReq.DeleteTask
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
