package zfield

import (
	"context"

	"github.com/grinderz/go-libs/libctx"
	"go.uber.org/zap"
)

var contextKey = libctx.Key("zap_field") //nolint:gochecknoglobals

// Context stores a list of zap.Field in a context.
// If the parent context had any zap.Field's they will be kept.
func Context(ctx context.Context, fields ...zap.Field) context.Context {
	old := GetFields(ctx)
	fields = append(old, fields...)

	return context.WithValue(ctx, contextKey, fields)
}

// GetFields returns the zap.Field's stored in the context or nil if none are found.
func GetFields(ctx context.Context) []zap.Field {
	fields, ok := ctx.Value(contextKey).([]zap.Field)
	if !ok {
		return nil
	}

	return fields
}

// WithContext gets the zap.Field's from a context and adds them to an existing logger.
func WithContext(ctx context.Context, log *zap.Logger) *zap.Logger {
	return log.With(GetFields(ctx)...)
}

// With adds fields to context and return a logger with the same fields added.
func With(ctx context.Context, log *zap.Logger, fields ...zap.Field) (context.Context, *zap.Logger) {
	ctx = Context(ctx, fields...)
	log = log.With(fields...)

	return ctx, log
}
