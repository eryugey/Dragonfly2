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

package logger

import (
	"fmt"
	"os"
	"strconv"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/grpclog"
)

var (
	CoreLogger       *zap.SugaredLogger
	GrpcLogger       *zap.SugaredLogger
	GCLogger         *zap.SugaredLogger
	StorageGCLogger  *zap.SugaredLogger
	JobLogger        *zap.SugaredLogger
	KeepAliveLogger  *zap.SugaredLogger
	StatSeedLogger   *zap.Logger
	DownloaderLogger *zap.Logger

	coreLogLevelEnabler zapcore.LevelEnabler
)

func init() {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	log, err := config.Build(zap.AddCaller(), zap.AddStacktrace(zap.WarnLevel), zap.AddCallerSkip(1))
	if err == nil {
		sugar := log.Sugar()
		SetCoreLogger(sugar)
		SetGrpcLogger(sugar)
		SetGCLogger(sugar)
		SetStorageGCLogger(sugar)
		SetKeepAliveLogger(sugar)
		SetStatSeedLogger(log)
		SetDownloadLogger(log)
		SetJobLogger(sugar)
	}
	levels = append(levels, config.Level)
}

// SetLevel updates all log level
func SetLevel(level zapcore.Level) {
	Infof("change log level to %s", level.String())
	for _, l := range levels {
		l.SetLevel(level)
	}
}

func SetCoreLogger(log *zap.SugaredLogger) {
	CoreLogger = log
	coreLogLevelEnabler = log.Desugar().Core()
}

func SetGCLogger(log *zap.SugaredLogger) {
	GCLogger = log
}

func SetStorageGCLogger(log *zap.SugaredLogger) {
	StorageGCLogger = log
}

func SetKeepAliveLogger(log *zap.SugaredLogger) {
	KeepAliveLogger = log
}

func SetStatSeedLogger(log *zap.Logger) {
	StatSeedLogger = log
}

func SetDownloadLogger(log *zap.Logger) {
	DownloaderLogger = log
}

func SetGrpcLogger(log *zap.SugaredLogger) {
	GrpcLogger = log
	var v int
	vLevel := os.Getenv("GRPC_GO_LOG_VERBOSITY_LEVEL")
	if vl, err := strconv.Atoi(vLevel); err == nil {
		v = vl
	}
	grpclog.SetLoggerV2(&zapGrpc{GrpcLogger, v})
}

func SetJobLogger(log *zap.SugaredLogger) {
	JobLogger = log
}

type SugaredLoggerOnWith struct {
	withArgs []interface{}
}

func With(args ...interface{}) *SugaredLoggerOnWith {
	return &SugaredLoggerOnWith{
		withArgs: args,
	}
}

func WithHostID(hostID string) *SugaredLoggerOnWith {
	return &SugaredLoggerOnWith{
		withArgs: []interface{}{"hostID", hostID},
	}
}

func WithTaskID(taskID string) *SugaredLoggerOnWith {
	return &SugaredLoggerOnWith{
		withArgs: []interface{}{"taskID", taskID},
	}
}

func WithTaskAndPeerID(taskID, peerID string) *SugaredLoggerOnWith {
	return &SugaredLoggerOnWith{
		withArgs: []interface{}{"taskID", taskID, "peerID", peerID},
	}
}

func WithTaskIDAndURL(taskID, url string) *SugaredLoggerOnWith {
	return &SugaredLoggerOnWith{
		withArgs: []interface{}{"taskID", taskID, "url", url},
	}
}

func WithHostnameAndIP(hostname, ip string) *SugaredLoggerOnWith {
	return &SugaredLoggerOnWith{
		withArgs: []interface{}{"hostname", hostname, "ip", ip},
	}
}

func (log *SugaredLoggerOnWith) With(args ...interface{}) *SugaredLoggerOnWith {
	args = append(args, log.withArgs...)
	return &SugaredLoggerOnWith{
		withArgs: args,
	}
}

func (log *SugaredLoggerOnWith) Infof(template string, args ...interface{}) {
	if !coreLogLevelEnabler.Enabled(zap.InfoLevel) {
		return
	}
	CoreLogger.Infow(fmt.Sprintf(template, args...), log.withArgs...)
}

func (log *SugaredLoggerOnWith) Info(args ...interface{}) {
	if !coreLogLevelEnabler.Enabled(zap.InfoLevel) {
		return
	}
	CoreLogger.Infow(fmt.Sprint(args...), log.withArgs...)
}

func (log *SugaredLoggerOnWith) Warnf(template string, args ...interface{}) {
	if !coreLogLevelEnabler.Enabled(zap.WarnLevel) {
		return
	}
	CoreLogger.Warnw(fmt.Sprintf(template, args...), log.withArgs...)
}

func (log *SugaredLoggerOnWith) Warn(args ...interface{}) {
	if !coreLogLevelEnabler.Enabled(zap.WarnLevel) {
		return
	}
	CoreLogger.Warnw(fmt.Sprint(args...), log.withArgs...)
}

func (log *SugaredLoggerOnWith) Errorf(template string, args ...interface{}) {
	if !coreLogLevelEnabler.Enabled(zap.ErrorLevel) {
		return
	}
	CoreLogger.Errorw(fmt.Sprintf(template, args...), log.withArgs...)
}

func (log *SugaredLoggerOnWith) Error(args ...interface{}) {
	if !coreLogLevelEnabler.Enabled(zap.ErrorLevel) {
		return
	}
	CoreLogger.Errorw(fmt.Sprint(args...), log.withArgs...)
}

func (log *SugaredLoggerOnWith) Debugf(template string, args ...interface{}) {
	if !coreLogLevelEnabler.Enabled(zap.DebugLevel) {
		return
	}
	CoreLogger.Debugw(fmt.Sprintf(template, args...), log.withArgs...)
}

func (log *SugaredLoggerOnWith) Debug(args ...interface{}) {
	if !coreLogLevelEnabler.Enabled(zap.DebugLevel) {
		return
	}
	CoreLogger.Debugw(fmt.Sprint(args...), log.withArgs...)
}

func Infof(template string, args ...interface{}) {
	CoreLogger.Infof(template, args...)
}

func Info(args ...interface{}) {
	CoreLogger.Info(args...)
}

func Warnf(template string, args ...interface{}) {
	CoreLogger.Warnf(template, args...)
}

func Warn(args ...interface{}) {
	CoreLogger.Warn(args...)
}

func Errorf(template string, args ...interface{}) {
	CoreLogger.Errorf(template, args...)
}

func Error(args ...interface{}) {
	CoreLogger.Error(args...)
}

func Debug(args ...interface{}) {
	CoreLogger.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	CoreLogger.Debugf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	CoreLogger.Fatalf(template, args...)
}

func Fatal(args ...interface{}) {
	CoreLogger.Fatal(args...)
}

type zapGrpc struct {
	*zap.SugaredLogger
	verbose int
}

func (z *zapGrpc) Infoln(args ...interface{}) {
	z.SugaredLogger.Info(args...)
}

func (z *zapGrpc) Warning(args ...interface{}) {
	z.SugaredLogger.Warn(args...)
}

func (z *zapGrpc) Warningln(args ...interface{}) {
	z.SugaredLogger.Warn(args...)
}

func (z *zapGrpc) Warningf(format string, args ...interface{}) {
	z.SugaredLogger.Warnf(format, args...)
}

func (z *zapGrpc) Errorln(args ...interface{}) {
	z.SugaredLogger.Error(args...)
}

func (z *zapGrpc) Fatalln(args ...interface{}) {
	z.SugaredLogger.Fatal(args...)
}

func (z *zapGrpc) V(level int) bool {
	return level <= z.verbose
}
