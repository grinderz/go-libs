package libzap

import (
	"fmt"

	"go.uber.org/zap"
)

var temporalCallerSkip = 1 //nolint:gochecknoglobals

type ITemporalLogger interface {
	Debug(s string, i ...any)
	Info(s string, i ...any)
	Warn(s string, i ...any)
	Error(s string, i ...any)
}

type ITemporalLoggerWithSkipCallers interface {
	WithCallerSkip(skip int) ITemporalLogger
}

type ITemporalLoggerWithLogger interface {
	With(i ...any) ITemporalLogger
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

func NewTemporalLoggerWith(log *zap.Logger, i ...any) *TemporalLogger {
	return &TemporalLogger{
		log: log.Sugar().With(i...).Desugar(),
	}
}

func (l *TemporalLogger) Error(s string, i ...any) {
	l.log.Error(fmt.Sprintf(s, i...))
}

func (l *TemporalLogger) Warn(s string, i ...any) {
	l.log.Warn(fmt.Sprintf(s, i...))
}

func (l *TemporalLogger) Info(s string, i ...any) {
	l.log.Info(fmt.Sprintf(s, i...))
}

func (l *TemporalLogger) Debug(s string, i ...any) {
	l.log.Debug(fmt.Sprintf(s, i...))
}

func (l *TemporalLogger) WithCallerSkip(i int) ITemporalLogger { //nolint:ireturn
	return NewTemporalLoggerWithCallerSkip(l.log, i)
}

func (l *TemporalLogger) With(i ...any) ITemporalLogger { //nolint:ireturn
	return NewTemporalLoggerWith(l.log, i...)
}

var (
	_ ITemporalLogger                = &TemporalLogger{nil}
	_ ITemporalLoggerWithSkipCallers = &TemporalLogger{nil}
	_ ITemporalLoggerWithLogger      = &TemporalLogger{nil}
)
