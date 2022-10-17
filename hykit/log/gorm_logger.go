package log

import (
	"context"
	"runtime"
	"time"

	tracerid "code.jshyjdtech.com/godev/hykit/pkg/tracer-id"

	"go.uber.org/zap"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

type gormLogger struct {
	sugar *zap.SugaredLogger

	logLevel glogger.LogLevel

	ez *EsimZap
}

type GormLoggerOptions struct{}

type GormLoggerOption func(c *gormLogger)

func NewGormLogger(options ...GormLoggerOption) glogger.Interface {
	glog := &gormLogger{}

	for _, option := range options {
		option(glog)
	}

	glog.logLevel = glogger.Error
	glog.sugar = glog.ez.Logger.Sugar()

	return glog
}

func WithGLogEsimZap(ez *EsimZap) GormLoggerOption {
	return func(gl *gormLogger) {
		gl.ez = ez
	}
}

func (gl *gormLogger) getArgs(ctx context.Context) []interface{} {
	args := make([]interface{}, 0)

	args = append(args, "caller", gl.ez.getCaller(runtime.Caller(5)))
	tracerID := tracerid.ExtractTracerID(ctx)
	if tracerID != "" {
		args = append(args, "tracer_id", tracerID)
	}
	return args
}
func (gl *gormLogger) LogMode(logLevel glogger.LogLevel) glogger.Interface {
	gl.logLevel = logLevel
	return gl
}

func (gl *gormLogger) Info(ctx context.Context, template string, args ...interface{}) {
	gl.sugar.With(gl.getArgs(ctx)...).Debugf(template, args...)
}

func (gl *gormLogger) Warn(ctx context.Context, template string, args ...interface{}) {
	gl.sugar.With(gl.getArgs(ctx)...).Warnf(template, args...)
}

func (gl *gormLogger) Error(ctx context.Context, template string, args ...interface{}) {
	gl.sugar.With(gl.getArgs(ctx)...).Errorf(template, args...)
}

func (gl *gormLogger) Trace(ctx context.Context, begin time.Time,
	fc func() (string, int64), err error) {
	if gl.logLevel > 0 && fc != nil {
		elapsed := time.Since(begin)
		switch {
		case err != nil && err != gorm.ErrRecordNotFound && gl.logLevel == glogger.Error:
			sql, rows := fc()
			gl.Error(ctx, "%.2fms [%d] %s : %s",
				float64(elapsed.Nanoseconds())/1e6, rows, sql, err.Error())
		case gl.logLevel == glogger.Warn:
			sql, rows := fc()
			gl.Warn(ctx, "%.2fms [%d] %s",
				float64(elapsed.Nanoseconds())/1e6, rows, sql)
		case gl.logLevel == glogger.Info:
			sql, rows := fc()
			gl.Info(ctx, "%.2fms [%d] %s",
				float64(elapsed.Nanoseconds())/1e6, rows, sql)
		default:
			sql, rows := fc()
			gl.Info(ctx, "%.2fms [%d] %s",
				float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
