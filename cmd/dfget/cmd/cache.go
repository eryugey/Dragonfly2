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
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"d7y.io/dragonfly/v2/client/config"
	"d7y.io/dragonfly/v2/client/dfget"
	"d7y.io/dragonfly/v2/cmd/dependency"
	logger "d7y.io/dragonfly/v2/internal/dflog"
	"d7y.io/dragonfly/v2/internal/dflog/logcore"
	"d7y.io/dragonfly/v2/pkg/rpc/dfdaemon/client"
	"d7y.io/dragonfly/v2/pkg/source"
	"d7y.io/dragonfly/v2/version"
)

var (
	cacheConfig       *config.DfgetConfig
	cacheDescription  = `TODO: cacheDescription`
	importDescription = `TODO: importDescription`
	exportDescription = `TODO: exportDescription`
	statDescription   = `TODO: statDescription`
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
		start := time.Now()

		// Initialize daemon dfpath
		d, err := initDfgetDfpath(cacheConfig)
		if err != nil {
			return err
		}

		// Initialize logger
		if err := logcore.InitDfget(cacheConfig.Console, d.LogDir()); err != nil {
			return errors.Wrap(err, "init client dfget logger")
		}

		// update plugin directory
		source.UpdatePluginDir(d.PluginDir())

		fmt.Printf("cacheConfig %v\n", cacheConfig)

		// Convert config
		if err := cacheConfig.Convert(args); err != nil {
			fmt.Printf("convert failed: %v\n", err)
			return err
		}

		// Validate config
		if err := cacheConfig.Validate(); err != nil {
			return err
		}

		logger.Infof("cacheConfig: %v", cacheConfig)

		// save file to cache system
		var errInfo string
		err = runDfgetImport(d.DfgetLockPath(), d.DaemonSockPath())
		if err != nil {
			errInfo = fmt.Sprintf("error: %v", err)
		}

		msg := fmt.Sprintf("import success: %t cost: %d ms %s", err == nil, time.Now().Sub(start).Milliseconds(), errInfo)
		logger.With("file", cacheConfig.Input).Info(msg)
		fmt.Println(msg)

		return errors.Wrapf(err, "import file: %s", cacheConfig.Input)
	},
}

// stat represents the cache stat command
var statCmd = &cobra.Command{
	// TODO: stat Use and Short
	Use:                "stat",
	Short:              "stat checks if given file exists in cache system",
	Long:               statDescription,
	Args:               cobra.MaximumNArgs(0),
	DisableAutoGenTag:  true,
	SilenceUsage:       true,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		// Initialize daemon dfpath
		d, err := initDfgetDfpath(cacheConfig)
		if err != nil {
			return err
		}

		// Initialize logger
		if err := logcore.InitDfget(cacheConfig.Console, d.LogDir()); err != nil {
			return errors.Wrap(err, "init client dfget logger")
		}

		// update plugin directory
		source.UpdatePluginDir(d.PluginDir())

		fmt.Printf("cacheConfig %v\n", cacheConfig)

		// Convert config
		if err := cacheConfig.Convert(args); err != nil {
			fmt.Printf("convert failed: %v\n", err)
			return err
		}

		// Validate config
		if err := cacheConfig.Validate(); err != nil {
			return err
		}

		logger.Infof("cacheConfig: %v", cacheConfig)

		// save file to cache system
		var errInfo string
		err = runDfgetStat(d.DfgetLockPath(), d.DaemonSockPath())
		if err != nil {
			errInfo = fmt.Sprintf("error: %v", err)
		}

		msg := fmt.Sprintf("stat success: %t cost: %d ms %s", err == nil, time.Now().Sub(start).Milliseconds(), errInfo)
		logger.With("statID", cacheConfig.StatID).Info(msg)
		fmt.Println(msg)

		return errors.Wrapf(err, "stat file: %s", cacheConfig.StatID)
	},
}

// exportCmd represents the cache export command
var exportCmd = &cobra.Command{
	// TODO: export Use and Short
	Use:                "export",
	Short:              "export file from cache system",
	Long:               exportDescription,
	Args:               cobra.MaximumNArgs(1),
	DisableAutoGenTag:  true,
	SilenceUsage:       true,
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	RunE: func(cmd *cobra.Command, args []string) error {
		start := time.Now()

		// Initialize daemon dfpath
		d, err := initDfgetDfpath(cacheConfig)
		if err != nil {
			return err
		}

		// Initialize logger
		if err := logcore.InitDfget(cacheConfig.Console, d.LogDir()); err != nil {
			return errors.Wrap(err, "init client dfget logger")
		}

		// update plugin directory
		source.UpdatePluginDir(d.PluginDir())

		fmt.Printf("cacheConfig %v\n", cacheConfig)

		// Convert config
		if err := cacheConfig.Convert(args); err != nil {
			fmt.Printf("convert failed: %v\n", err)
			return err
		}

		// Validate config
		if err := cacheConfig.Validate(); err != nil {
			return err
		}

		logger.Infof("cacheConfig: %v", cacheConfig)

		// export file from cache system
		var errInfo string
		err = runDfgetExport(d.DfgetLockPath(), d.DaemonSockPath())
		if err != nil {
			errInfo = fmt.Sprintf("error: %v", err)
		}

		msg := fmt.Sprintf("export success: %t cost: %d ms %s", err == nil, time.Now().Sub(start).Milliseconds(), errInfo)
		logger.With("ID", cacheConfig.ExportID).Info(msg)
		fmt.Println(msg)

		return errors.Wrapf(err, "export file: %s", cacheConfig.ExportID)
	},
}

func init() {
	cacheCmd.AddCommand(importCmd)
	cacheCmd.AddCommand(statCmd)
	cacheCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(cacheCmd)

	// Initialize default dfget config
	cacheConfig = new(config.DfgetConfig)
	cacheConfig.IsCache = true

	// Initialize cobra
	dependency.InitCobra(importCmd, false, cacheConfig)
	dependency.InitCobra(statCmd, false, cacheConfig)

	// Add flags
	addImportCmdFlags(importCmd)
	addStatCmdFlags(statCmd)
	addExportCmdFlags(exportCmd)
}

func addImportCmdFlags(cmd *cobra.Command) {
	flagSet := cmd.Flags()

	flagSet.StringP("input", "I", cacheConfig.Input,
		"Import file into cache system, equivalent to the command's first position argument")

	flagSet.StringP("inputid", "i", cacheConfig.InputID,
		"Identification of the imported file, could be a hash or a URL")

	// FIXME: following flags are not working
	flagSet.StringP("tag", "t", cacheConfig.Tag,
		"Different tags for the same ID will be divided into different P2P overlay, it conflicts with --digest")

	flagSet.StringP("callsystem", "c", cacheConfig.CallSystem, "The caller name which is mainly used for statistics and access control")

	flagSet.StringP("workhome", "w", cacheConfig.WorkHome, "Dfget working directory")

	flagSet.StringP("logdir", "l", cacheConfig.LogDir, "Dfget log directory")

	if err := viper.BindPFlags(flagSet); err != nil {
		panic(errors.Wrap(err, "bind dfget flags to viper"))
	}
}

func addStatCmdFlags(cmd *cobra.Command) {
	flagSet := cmd.Flags()

	flagSet.StringP("statid", "i", cacheConfig.StatID,
		"Identification of the stated file, could be a hash or a URL")

	flagSet.BoolP("local", "l", cacheConfig.LocalOnly,
		"Only find local cache, skip P2P network")

	// FIXME: following flags are not working
	flagSet.StringP("tag", "t", cacheConfig.Tag,
		"Different tags for the same ID will be divided into different P2P overlay, it conflicts with --digest")

	flagSet.StringP("callsystem", "c", cacheConfig.CallSystem, "The caller name which is mainly used for statistics and access control")

	flagSet.StringP("workhome", "w", cacheConfig.WorkHome, "Dfget working directory")

	//flagSet.StringP("logdir", "l", cacheConfig.LogDir, "Dfget log directory")

	if err := viper.BindPFlags(flagSet); err != nil {
		panic(errors.Wrap(err, "bind dfget flags to viper"))
	}
}

func addExportCmdFlags(cmd *cobra.Command) {
	flagSet := cmd.Flags()

	flagSet.StringP("exportoutput", "E", cacheConfig.ExportOutput,
		"Destination of exported file from cache system, equivalent to the command's first position argument")

	flagSet.StringP("exportid", "e", cacheConfig.ExportID,
		"Identification of the export file, could be a hash or a URL")

	flagSet.BoolP("local", "l", cacheConfig.LocalOnly,
		"Only find local cache, skip P2P network")

	if err := viper.BindPFlags(flagSet); err != nil {
		panic(errors.Wrap(err, "bind dfget flags to viper"))
	}
}

// runDfgetImport does some init operations and starts to import file.
func runDfgetImport(dfgetLockPath, daemonSockPath string) error {
	logger.Infof("Version:\n%s", version.Version())

	// Dfget config values
	s, _ := yaml.Marshal(cacheConfig)
	logger.Infof("client dfget configuration:\n%s", string(s))

	ff := dependency.InitMonitor(cacheConfig.Verbose, cacheConfig.PProfPort, cacheConfig.Telemetry)
	defer ff()

	var (
		daemonClient client.DaemonClient
		err          error
	)

	logger.Info("start to check and spawn daemon")
	if daemonClient, err = checkAndSpawnDaemon(dfgetLockPath, daemonSockPath); err != nil {
		logger.Errorf("check and spawn daemon error: %v", err)
	} else {
		logger.Info("check and spawn daemon success")
	}

	return dfget.Import(cacheConfig, daemonClient)
}

// runDfgetStat does some init operations and starts to stat file.
func runDfgetStat(dfgetLockPath, daemonSockPath string) error {
	logger.Infof("Version:\n%s", version.Version())

	// Dfget config values
	s, _ := yaml.Marshal(cacheConfig)
	logger.Infof("client dfget configuration:\n%s", string(s))

	ff := dependency.InitMonitor(cacheConfig.Verbose, cacheConfig.PProfPort, cacheConfig.Telemetry)
	defer ff()

	var (
		daemonClient client.DaemonClient
		err          error
	)

	logger.Info("start to check and spawn daemon")
	if daemonClient, err = checkAndSpawnDaemon(dfgetLockPath, daemonSockPath); err != nil {
		logger.Errorf("check and spawn daemon error: %v", err)
	} else {
		logger.Info("check and spawn daemon success")
	}

	return dfget.Stat(cacheConfig, daemonClient)
}

// runDfgetExport does some init operations and starts to export file.
func runDfgetExport(dfgetLockPath, daemonSockPath string) error {
	logger.Infof("Version:\n%s", version.Version())

	// Dfget config values
	s, _ := yaml.Marshal(cacheConfig)
	logger.Infof("client dfget configuration:\n%s", string(s))

	ff := dependency.InitMonitor(cacheConfig.Verbose, cacheConfig.PProfPort, cacheConfig.Telemetry)
	defer ff()

	var (
		daemonClient client.DaemonClient
		err          error
	)

	logger.Info("start to check and spawn daemon")
	if daemonClient, err = checkAndSpawnDaemon(dfgetLockPath, daemonSockPath); err != nil {
		logger.Errorf("check and spawn daemon error: %v", err)
	} else {
		logger.Info("check and spawn daemon success")
	}

	return dfget.Export(cacheConfig, daemonClient)
}
