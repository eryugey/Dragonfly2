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

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"

	"github.com/gofrs/flock"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"d7y.io/dragonfly/v2/client/config"
	"d7y.io/dragonfly/v2/client/dfget"
	"d7y.io/dragonfly/v2/cmd/dependency"
	"d7y.io/dragonfly/v2/internal/constants"
	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/internal/dflog/logcore"
	"d7y.io/dragonfly/v2/internal/dfnet"
	"d7y.io/dragonfly/v2/pkg/basic"
	"d7y.io/dragonfly/v2/pkg/dfpath"
	"d7y.io/dragonfly/v2/pkg/rpc/dfdaemon/client"
	"d7y.io/dragonfly/v2/pkg/source"
	"d7y.io/dragonfly/v2/pkg/unit"
	"d7y.io/dragonfly/v2/pkg/util/net/iputils"
	"d7y.io/dragonfly/v2/version"
)

var (
	dfgetConfig *config.DfgetConfig
)

var (
	cacheDescription = `TODO: cacheDescription`
	importDescription = `TODO: importDescription`
)

// cacheCmd represents the cache command, and it requires sub-commands, so no Run or RunE provided
var cacheCmd = &cobra.Command{
	// TODO: cache Use and Short
	Use:                "cache",
	Short:              "cache system",
	Long:               cacheDescription,
	Args:               cobra.MaximumNArgs(1),
	DisableAutoGenTag:  true,
	SilenceUsage:       true,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
}

// importCmd represents the cache import command
var importCmd = &cobra.Command{
	// TODO: import Use and Short
	Use:                "import",
	Short:              "import file to cache system",
	Long:               importDescription,
	Args:               cobra.MaximumNArgs(1),
	DisableAutoGenTag:  true,
	SilenceUsage:       true,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE: func(cmd *cobra.Command, args []string) error {
	},
}

func init() {
	cacheCmd.AddCommand(importCmd)
	rootCmd.AddCommand(cacheCmd)

	// Initialize default dfget config
	dfgetConfig = config.NewDfgetConfig()
	// Initialize cobra
	dependency.InitCobra(importCmd, false, dfgetConfig)

	// Add flags
	addImportCmdFlags(importCmd)
}

func addImportCmdFlags(cmd *cobra.Command) {
	importFlag := cmd.Flags()

	flagSet.StringP("file", "f", dfgetConfig.Output,
		"Import file into cache system, equivalent to the command's first position argument")

	flagSet.StringP("id", "i", dfgetConfig.URL,
		"Destination path which is used to store the downloaded file, it must be a full path")

	flagSet.String("tag", dfgetConfig.Tag,
		"Different tags for the same url will be divided into different P2P overlay, it conflicts with --digest")

	flagSet.StringSliceP("header", "H", dfgetConfig.Header, "url header, eg: --header='Accept: *' --header='Host: abc'")

	flagSet.String("callsystem", dfgetConfig.CallSystem, "The caller name which is mainly used for statistics and access control")

	flagSet.String("workhome", dfgetConfig.WorkHome, "Dfget working directory")

	flagSet.String("logdir", dfgetConfig.LogDir, "Dfget log directory")
}
