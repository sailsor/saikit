package log

import (
	"context"
	"os"
	"runtime"
	"time"

	"code.jshyjdtech.com/godev/hykit/config"
	tracerid "code.jshyjdtech.com/godev/hykit/pkg/tracer-id"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type EsimZap struct {
	Config

	conf config.Config

	*zap.Logger

	debug bool

	json bool

	// 日志等级，参考zapcore.Level.
	logLevel zapcore.Level
}

type ZapOption func(c *EsimZap)

func NewEsimZap(options ...ZapOption) *EsimZap {
	ez := &EsimZap{}

	ez.logLevel = zap.InfoLevel
	for _, option := range options {
		option(ez)
	}

	if ez.conf == nil {
		ez.conf = config.NewMemConfig()
	}

	// 文件配置
	ez.Config.fillWithDefaultConfig(ez.conf)

	hook := &lumberjack.Logger{
		Filename:   ez.Config.File,
		MaxSize:    ez.Config.MaxSize,
		MaxBackups: ez.Config.BackupCount,
		MaxAge:     ez.Config.MaxAge,
		Compress:   false,
	}

	var writer []zapcore.WriteSyncer
	switch {
	case ez.Config.IsBothFileStdout():
		writer = append(writer, zapcore.AddSync(hook), zapcore.AddSync(os.Stdout))
	case ez.Config.IsOutFile():
		writer = append(writer, zapcore.AddSync(hook))
	case ez.Config.IsOutStdout():
		writer = append(writer, zapcore.AddSync(os.Stdout))
	}

	var opts = make([]zap.Option, 0)

	/*此处文件如已配置级别，则使用配置文件级别为准*/
	if ez.Config.Level != "" {
		ez.logLevel = ParseLevel(ez.Config.Level)
	} else if ez.debug {
		ez.logLevel = zap.DebugLevel
	}

	/*此处文件如已配置文件格式，则使用文件格式为准*/
	if ez.Config.Format == "" {
		if ez.json {
			ez.Config.Format = "json"
		} else {
			ez.Config.Format = "text"
		}
	}

	if ez.Config.ReportCaller {
		opts = append(opts, zap.AddCaller())
	}

	if ez.Config.Stacktrace {
		opts = append(opts, zap.AddStacktrace(ez.logLevel))
	}

	var core []zapcore.Core
	for _, w := range writer {
		core = append(core, zapcore.NewCore(ez.buildEncoder(), w, zap.NewAtomicLevelAt(ez.logLevel)))
	}

	ez.Logger = zap.New(zapcore.NewTee(core...), opts...)

	return ez
}

func WithEsimZapConf(conf config.Config) ZapOption {
	return func(ez *EsimZap) {
		ez.conf = conf
	}
}

func WithEsimZapDebug(debug bool) ZapOption {
	return func(ez *EsimZap) {
		ez.debug = debug
	}
}

func WithEsimZapJSON(json bool) ZapOption {
	return func(ez *EsimZap) {
		ez.json = json
	}
}

func WithLogLevel(level zapcore.Level) ZapOption {
	return func(ez *EsimZap) {
		ez.logLevel = level
	}
}

func (ez *EsimZap) getArgs(ctx context.Context, logLevel zapcore.Level) []interface{} {
	args := make([]interface{}, 0)

	//info 及以下级别不需要打印调用函数位置
	if logLevel > zap.InfoLevel {
		args = append(args, "caller", ez.getCaller(runtime.Caller(2)))
	}

	tracerID := tracerid.ExtractTracerID(ctx)
	if tracerID != "" {
		args = append(args, "tracer_id", tracerID)
	}
	if fld, ok := ctx.Value(externalFieldKey).(Field); ok {
		for k, v := range fld {
			args = append(args, k, v)
		}
	}

	return args
}

func (ez *EsimZap) getCaller(pc uintptr, file string, line int, ok bool) string {
	return zapcore.NewEntryCaller(pc, file, line, ok).TrimmedPath()
}

func (ez *EsimZap) buildEncoder() zapcore.Encoder {
	var (
		encoder zapcore.EncoderConfig
	)

	if ez.debug {
		encoder = zap.NewDevelopmentEncoderConfig()
	} else {
		encoder = zap.NewProductionEncoderConfig()
	}

	encoder.TimeKey = "time"
	encoder.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.999999"))
	}
	encoder.EncodeCaller = zapcore.FullCallerEncoder
	encoder.EncodeName = zapcore.FullNameEncoder

	if ez.Config.Format == "json" {
		return zapcore.NewJSONEncoder(encoder)
	}
	return zapcore.NewConsoleEncoder(encoder)
}
