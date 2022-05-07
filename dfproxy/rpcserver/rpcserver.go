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

package rpcserver

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"d7y.io/dragonfly/v2/dfproxy/service"
	"d7y.io/dragonfly/v2/pkg/rpc"
	"d7y.io/dragonfly/v2/pkg/rpc/dfproxy"
)

// Server is grpc server
type Server struct {
	// Service interface
	service *service.Service

	// GRPC UnimplementedDaemonProxyServer interface
	dfproxy.UnimplementedDaemonProxyServer
}

// New returns a new transparent dfproxy server from the given options
func New(service *service.Service, opts ...grpc.ServerOption) *grpc.Server {
	svr := &Server{service: service}
	grpcServer := grpc.NewServer(append(rpc.DefaultServerOptions(), opts...)...)

	// Register servers on grpc server
	dfproxy.RegisterDaemonProxyServer(grpcServer, svr)
	healthpb.RegisterHealthServer(grpcServer, health.NewServer())
	return grpcServer
}

// Dfdaemon proxies dfdaemon requests
func (s *Server) Dfdaemon(stream dfproxy.DaemonProxy_DfdaemonServer) error {
	return s.service.Dfdaemon(stream)
}
