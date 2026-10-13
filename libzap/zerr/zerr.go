// Package zerr forked from https://github.com/yzzyx/zerr by Elias Norberg
// MIT license
package zerr

import (
	"errors"
	"fmt"
	"os"
	"slices"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Error struct {
	err      error
	fields   []zap.Field
	hasStack bool
}

func (e *Error) Error() string {
	if e == nil || e.err == nil {
		return ""
	}

	return e.err.Error()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.err
}

func IsError(err error) bool {
	_, ok := errors.AsType[*Error](err)

	return ok
}

func (e *Error) Fields() []zap.Field {
	if e == nil {
		return nil
	}

	// Clone so appends never write into the shared backing array of e.fields.
	fields := slices.Clone(e.fields)
	err := e

	// The err != nil guard stops on a typed-nil *Error in the chain.
	for errors.As(err.err, &err) && err != nil {
		fields = append(fields, err.fields...)
	}

	return fields
}

// WithField wraps e with extra fields. A nil e yields nil, the same as
// Wrap(nil): there is no error to attach the fields to.
func (e *Error) WithField(field zap.Field, fields ...zap.Field) *Error {
	if e == nil {
		return nil
	}

	// Clone so the append never writes into a caller-owned backing array.
	return &Error{
		err:      e,
		fields:   append(slices.Clone(fields), field),
		hasStack: e.hasStack,
	}
}

func (e *Error) LogError(logger *zap.Logger, message string) {
	e.log(logger, zapcore.ErrorLevel, message)
}

func (e *Error) LogWarn(logger *zap.Logger, message string) {
	e.log(logger, zapcore.WarnLevel, message)
}

func (e *Error) log(logger *zap.Logger, level zapcore.Level, message string) {
	if e == nil {
		return
	}

	if logger == nil {
		fmt.Fprintf(os.Stderr, "[%s] %s %+v %+v\n", level, message, e.Error(), e.Fields())
		return
	}

	if message == "" {
		logger.Log(level, e.Error(), e.Fields()...)
	} else {
		logger.Log(level, fmt.Sprintf("%s: %v", message, e.Error()), e.Fields()...)
	}
}

// Wrap wraps err into *Error, attaching fields and a stacktrace unless the
// chain already carries one. Wrap(nil) returns a nil *Error: returned as an
// error interface it is non-nil, so check err before wrapping it in
// `return zerr.Wrap(err)`.
func Wrap(err error, fields ...zap.Field) *Error {
	if err == nil {
		return nil
	}

	return wrapWithStack(1, err, fields...)
}

func wrapWithStack(lvl int, err error, fields ...zap.Field) *Error {
	// Identity check only: errors.As here would return an inner *Error and
	// drop the outer wrapping context.
	zerr, ok := err.(*Error) //nolint:errorlint // identity check is intentional
	if ok && zerr != nil && len(fields) == 0 && zerr.hasStack {
		return zerr
	}

	hasStack := false
	if inner, ok := errors.AsType[*Error](err); ok && inner != nil {
		hasStack = inner.hasStack
	}

	// Clone so the append never writes into a caller-owned backing array.
	fields = slices.Clone(fields)

	if !hasStack {
		fields = append(fields, zap.StackSkip("zerr_stacktrace", lvl+1))
	}

	return &Error{
		err:      err,
		fields:   fields,
		hasStack: true,
	}
}

// WrapNoStack wraps err into *Error without attaching a stacktrace.
// WrapNoStack(nil) returns a nil *Error, with the same caveat as Wrap.
func WrapNoStack(err error, fields ...zap.Field) *Error {
	if err == nil {
		return nil
	}

	// Clone so later caller appends never mutate the stored fields.
	return &Error{
		err:      err,
		fields:   slices.Clone(fields),
		hasStack: false,
	}
}

func Fields(err error) []zap.Field {
	if zerr, ok := errors.AsType[*Error](err); ok {
		return zerr.Fields()
	}

	return nil
}
