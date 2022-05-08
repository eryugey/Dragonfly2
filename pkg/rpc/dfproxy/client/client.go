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

package client

import (
	"context"
	"time"

	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/pkg/dfnet"
	"d7y.io/dragonfly/v2/pkg/rpc"
	"d7y.io/dragonfly/v2/pkg/rpc/dfproxy"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func GetClientByAddr(addrs []dfnet.NetAddr, opts ...grpc.DialOption) (DaemonProxyClient, error) {
	if len(addrs) == 0 {
		return nil, errors.New("address list of dfproxy is empty")
	}
	pc := &daemonProxyClient{
		rpc.NewConnection(context.Background(), "dfproxy-static", addrs, []rpc.ConnOption{
			rpc.WithConnExpireTime(0 * time.Second),
			rpc.WithDialOption(opts),
		}),
	}
	logger.Infof("dfproxy server list: %s", addrs)
	return pc, nil
}

//go:generate mockgen -package mocks -source client.go -destination ./mocks/client_mock.go
// DaemonProxyClient see dfproxy.DaemonProxyClient
type DaemonProxyClient interface {
	// Dfdaemon proxy service
	Dfdaemon(ctx context.Context, opts ...grpc.CallOption) (dfproxy.DaemonProxy_DfdaemonClient, error)
	Close() error
}

type daemonProxyClient struct {
	*rpc.Connection
}

func (pc *daemonProxyClient) getDaemonProxyClient(key string, stick bool) (dfproxy.DaemonProxyClient, string, error) {
	clientConn, err := pc.Connection.GetClientConn(key, stick)
	if err != nil {
		return nil, "", err
	}
	return dfproxy.NewDaemonProxyClient(clientConn), clientConn.Target(), nil
}

func (pc *daemonProxyClient) Dfdaemon(ctx context.Context, opts ...grpc.CallOption) (dfproxy.DaemonProxy_DfdaemonClient, error) {
	client, target, err := pc.getDaemonProxyClient("dfproxy-static", false)
	if err != nil {
		return nil, err
	}
	stream, err := client.Dfdaemon(ctx, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "Dfdaemon failed on target %s", target)
	}
	logger.Infof("Start dfproxy Dfdaemon stream on target %s", target)
	return stream, nil
}

var _ DaemonProxyClient = (*daemonProxyClient)(nil)
