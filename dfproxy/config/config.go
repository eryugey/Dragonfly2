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

package config

import (
	"github.com/pkg/errors"

	"d7y.io/dragonfly/v2/cmd/dependency/base"
	"d7y.io/dragonfly/v2/pkg/dfnet"
	"d7y.io/dragonfly/v2/pkg/dfpath"
)

type Config struct {
	// Base options
	base.Options `yaml:",inline" mapstructure:",squash"`

	// Server configuration
	Server *ServerConfig `yaml:"server" mapstructure:"server"`

	// Dfdaemon config
	Daemon *DaemonConfig `yaml:"daemon" mapstructure:"daemon"`
}

type ServerConfig struct {
	// ProxyGRPC defines dfproxy GRPC listen config
	ProxyGRPC *dfnet.NetAddr `yaml:"proxyGRPC" mapstructure:"proxyGRPC"`

	// Server log directory
	LogDir string `yaml:"logDir" mapstructure:"logDir"`
}

type DaemonConfig struct {
	DaemonGRPC *dfnet.NetAddr `yaml:"daemonGRPC" mapstructure:"daemonGRPC"`
}

// New default configuration
func New() *Config {
	return &Config{
		Server: &ServerConfig{
			ProxyGRPC: &dfnet.NetAddr{
				Type: dfnet.UNIX,
				Addr: "/run/dragonfly/dfproxy.sock",
			},
			LogDir: dfpath.DefaultLogDir,
		},
		Daemon: &DaemonConfig{
			DaemonGRPC: &dfnet.NetAddr{
				Type: dfnet.UNIX,
				Addr: "/usr/local/dragonfly/daemon.sock",
			},
		},
	}
}

// Validate config parameters
func (c *Config) Validate() error {
	if c.Server == nil {
		return errors.New("requires parameter server")
	}

	if c.Server.ProxyGRPC == nil {
		return errors.New("server requires parameter server.proxyGRPC")
	}

	if c.Daemon == nil {
		return errors.New("requires parameter daemon")
	}

	if c.Daemon.DaemonGRPC == nil {
		return errors.New("daemon requires parameter daemon.daemonGRPC")
	}

	return nil
}
