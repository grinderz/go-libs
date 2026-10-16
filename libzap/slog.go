package libzap

import (
	"log/slog"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
)

// NewSlogLogger returns a slog.Logger that writes into the core of log, for
// libraries that take a *slog.Logger. Fields added with log.With, the level
// and the logger name carry over; caller and stack trace options of log do
// not, as a *zap.Logger does not expose them. The handler adds the caller and
// takes stack traces at error level by default, opts override that.
func NewSlogLogger(log *zap.Logger, opts ...zapslog.HandlerOption) *slog.Logger {
	handlerOpts := append([]zapslog.HandlerOption{
		zapslog.WithName(log.Name()),
		zapslog.WithCaller(true),
	}, opts...)

	return slog.New(zapslog.NewHandler(log.Core(), handlerOpts...))
}
