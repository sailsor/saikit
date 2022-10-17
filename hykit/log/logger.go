package log

import (
	"context"

	"go.uber.org/zap"
)

type logger struct {
	debug bool

	json bool

	ez *EsimZap

	logger *zap.Logger

	sugar *zap.SugaredLogger
}

type Field map[string]interface{}

var externalFieldKey struct{}

type LoggerOptions struct{}

type Option func(c *logger)

func NewLogger(options ...Option) Logger {
	logger := &logger{}

	for _, option := range options {
		option(logger)
	}

	if logger.ez == nil {
		logger.ez = NewEsimZap(
			WithEsimZapDebug(true),
			WithEsimZapJSON(true),
		)
	}

	logger.logger = logger.ez.Logger
	logger.sugar = logger.ez.Logger.Sugar()

	return logger
}

func WithDebug(debug bool) Option {
	return func(l *logger) {
		l.debug = debug
	}
}

func WithJSON(json bool) Option {
	return func(l *logger) {
		l.json = json
	}
}

func WithEsimZap(ez *EsimZap) Option {
	return func(l *logger) {
		l.ez = ez
	}
}

func (log *logger) Error(msg string) {
	log.logger.Error(msg)
}

func (log *logger) Printf(template string, args ...interface{}) {
	log.Debugf(template, args...)
}

func (log *logger) Debugf(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO(), zap.DebugLevel)...).Debugf(template, args...)
}

func (log *logger) Infof(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO(), zap.InfoLevel)...).Infof(template, args...)
}

func (log *logger) Warnf(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO(), zap.WarnLevel)...).Warnf(template, args...)
}

func (log *logger) Errorf(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO(), zap.ErrorLevel)...).Errorf(template, args...)
}

func (log *logger) DPanicf(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO(), zap.DPanicLevel)...).DPanicf(template, args...)
}

func (log *logger) Panicf(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO(), zap.PanicLevel)...).Panicf(template, args...)
}

func (log *logger) Fatalf(template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(context.TODO(), zap.FatalLevel)...).Fatalf(template, args...)
}

func (log *logger) Debugc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx, zap.DebugLevel)...).Debugf(template, args...)
}

func (log *logger) Infoc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx, zap.InfoLevel)...).Infof(template, args...)
}

func (log *logger) Warnc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx, zap.WarnLevel)...).Warnf(template, args...)
}

func (log *logger) Errorc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx, zap.ErrorLevel)...).Errorf(template, args...)
}

func (log *logger) DPanicc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx, zap.DPanicLevel)...).DPanicf(template, args...)
}

func (log *logger) Panicc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx, zap.PanicLevel)...).Panicf(template, args...)
}

func (log *logger) Fatalc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.ez.getArgs(ctx, zap.FatalLevel)...).Fatalf(template, args...)
}

func (log *logger) SetFields(ctx context.Context, field Field) context.Context {
	return context.WithValue(ctx, externalFieldKey, field)
}

func (log *logger) WithFields(ctx context.Context, field Field) Logger {
	args := make([]interface{}, 0)
	if len(field) > 0 {
		for k, v := range field {
			args = append(args, k, v)
		}
	}
	l := &logger{
		debug:  log.debug,
		json:   log.json,
		ez:     log.ez,
		logger: log.logger,
		sugar:  log.sugar.With(args...),
	}
	return l
}

//仅用于外部传入相关日志打印键值
func SetFields(ctx context.Context, field Field) context.Context {
	return context.WithValue(ctx, externalFieldKey, field)
}
