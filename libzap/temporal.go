package libzap

import (
	"fmt"

	"go.uber.org/zap"
)

var temporalCallerSkip = 1 //nolint:gochecknoglobals

type ITemporalLogger interface {
	Debug(s string, i ...interface{})
	Info(s string, i ...interface{})
	Warn(s string, i ...interface{})
	Error(s string, i ...interface{})
}

type ITemporalLoggerWithSkipCallers interface {
	WithCallerSkip(skip int) ITemporalLogger
}

type ITemporalLoggerWithLogger interface {
	With(i ...interface{}) ITemporalLogger
}

type TemporalLogger struct {
	log *zap.Logger
}

func NewTemporalLogger(log *zap.Logger) *TemporalLogger {
	return &TemporalLogger{
		log: log.WithOptions(zap.AddCallerSkip(temporalCallerSkip)).With(FieldPkg("temporal_sdk")),
	}
}

func NewTemporalLoggerWithCallerSkip(log *zap.Logger, i int) *TemporalLogger {
	return &TemporalLogger{
		log: log.WithOptions(zap.AddCallerSkip(i)),
	}
}

func NewTemporalLoggerWith(log *zap.Logger, i ...interface{}) *TemporalLogger {
	return &TemporalLogger{
		log: log.Sugar().With(i...).Desugar(),
	}
}

func (l *TemporalLogger) Error(s string, i ...interface{}) {
	l.log.Error(fmt.Sprintf(s, i...))
}

func (l *TemporalLogger) Warn(s string, i ...interface{}) {
	l.log.Warn(fmt.Sprintf(s, i...))
}

func (l *TemporalLogger) Info(s string, i ...interface{}) {
	l.log.Info(fmt.Sprintf(s, i...))
}

func (l *TemporalLogger) Debug(s string, i ...interface{}) {
	l.log.Debug(fmt.Sprintf(s, i...))
}

func (l *TemporalLogger) WithCallerSkip(i int) ITemporalLogger { //nolint:ireturn
	return NewTemporalLoggerWithCallerSkip(l.log, i)
}

func (l *TemporalLogger) With(i ...interface{}) ITemporalLogger { //nolint:ireturn
	return NewTemporalLoggerWith(l.log, i...)
}

var (
	_ ITemporalLogger                = &TemporalLogger{nil}
	_ ITemporalLoggerWithSkipCallers = &TemporalLogger{nil}
	_ ITemporalLoggerWithLogger      = &TemporalLogger{nil}
)
