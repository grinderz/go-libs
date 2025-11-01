package libzap

import (
	"fmt"

	"go.uber.org/zap"
)

var badgerCallerSkip = 2 //nolint:gochecknoglobals

type IBadgerLogger interface {
	Errorf(s string, i ...any)
	Warningf(s string, i ...any)
	Infof(s string, i ...any)
	Debugf(s string, i ...any)
}

type BadgerLogger struct {
	log *zap.Logger
}

func NewBadgerLogger(log *zap.Logger) *BadgerLogger {
	return &BadgerLogger{
		log: log.WithOptions(zap.AddCallerSkip(badgerCallerSkip)).With(FieldPkg("badger")),
	}
}

func (l *BadgerLogger) Errorf(s string, i ...any) {
	l.log.Error(fmt.Sprintf(s, i...))
}

func (l *BadgerLogger) Warningf(s string, i ...any) {
	l.log.Warn(fmt.Sprintf(s, i...))
}

func (l *BadgerLogger) Infof(s string, i ...any) {
	l.log.Info(fmt.Sprintf(s, i...))
}

func (l *BadgerLogger) Debugf(s string, i ...any) {
	l.log.Debug(fmt.Sprintf(s, i...))
}

var _ IBadgerLogger = &BadgerLogger{nil}
