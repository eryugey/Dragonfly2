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

package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"d7y.io/dragonfly/v2/client/config"
	"d7y.io/dragonfly/v2/cmd/dependency"
	"d7y.io/dragonfly/v2/internal/dflog/logcore"
	"d7y.io/dragonfly/v2/pkg/source"
)

var (
	cacheDescription  = `TODO: cacheDescription`
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
		// Initialize daemon dfpath
		d, err := initDfgetDfpath(dfgetConfig)
		if err != nil {
			return err
		}

		// Initialize logger
		if err := logcore.InitDfget(dfgetConfig.Console, d.LogDir()); err != nil {
			return errors.Wrap(err, "init client dfget logger")
		}

		// update plugin directory
		source.UpdatePluginDir(d.PluginDir())

		fmt.Printf("dfgetConfig: \"%v\"\n", dfgetConfig)

		// Convert config
		if err := dfgetConfig.Convert(args); err != nil {
			fmt.Printf("convert failed: %v\n", err)
			return err
		}

		// Validate config
		if err := dfgetConfig.Validate(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	cacheCmd.AddCommand(importCmd)
	rootCmd.AddCommand(cacheCmd)

	// Initialize default dfget config
	dfgetConfig = config.NewDfgetConfig()
	dfgetConfig.IsCache = true

	// Initialize cobra
	dependency.InitCobra(importCmd, false, dfgetConfig)

	// Add flags
	addImportCmdFlags(importCmd)
}

func addImportCmdFlags(cmd *cobra.Command) {
	flagSet := cmd.Flags()

	flagSet.StringP("input", "I", dfgetConfig.Input,
		"Import file into cache system, equivalent to the command's first position argument")

	flagSet.StringP("inputid", "i", dfgetConfig.InputID,
		"Identification of the imported file, could be a hash or a URL")

	flagSet.StringP("tag", "t", dfgetConfig.Tag,
		"Different tags for the same ID will be divided into different P2P overlay, it conflicts with --digest")

	flagSet.StringP("callsystem", "c", dfgetConfig.CallSystem, "The caller name which is mainly used for statistics and access control")

	flagSet.StringP("workhome", "w", dfgetConfig.WorkHome, "Dfget working directory")

	flagSet.StringP("logdir", "l", dfgetConfig.LogDir, "Dfget log directory")

	if err := viper.BindPFlags(flagSet); err != nil {
		panic(errors.Wrap(err, "bind dfget flags to viper"))
	}
}
