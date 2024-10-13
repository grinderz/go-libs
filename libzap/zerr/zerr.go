// Package zerr forked from https://github.com/yzzyx/zerr by Elias Norberg
// MIT license
package zerr

import (
	"errors"

	"go.uber.org/zap"
)

type Error struct {
	err      error
	fields   []zap.Field
	hasStack bool
}

func (e *Error) Error() string {
	if e.err == nil {
		return ""
	}

	return e.err.Error()
}

func (e *Error) Unwrap() error {
	return e.err
}

func IsError(err error) bool {
	var e *Error
	return errors.As(err, &e)
}

func (e *Error) Fields() []zap.Field {
	fields := e.fields
	err := e

	for {
		if !errors.As(err.err, &err) {
			break
		}

		fields = append(fields, err.fields...)
	}

	return fields
}

func (e *Error) WithField(f zap.Field, fields ...zap.Field) *Error {
	return &Error{
		err:      e,
		fields:   append(fields, f),
		hasStack: e.hasStack,
	}
}

func Wrap(err error, fields ...zap.Field) *Error {
	return wrapWithStack(1, err, fields...)
}

func wrapWithStack(lvl int, err error, fields ...zap.Field) *Error {
	var zerr *Error
	if errors.As(err, &zerr) && len(fields) == 0 {
		return zerr
	}

	hasStack := false
	if errors.As(err, &zerr) {
		hasStack = zerr.hasStack
	}

	if !hasStack {
		fields = append(fields, zap.StackSkip("zerr_stacktrace", lvl+1))
	}

	return &Error{
		err:      err,
		fields:   fields,
		hasStack: true,
	}
}

func WrapNoStack(err error, fields ...zap.Field) *Error {
	return &Error{
		err:    err,
		fields: fields,
	}
}

func Fields(err error) []zap.Field {
	var zerr *Error
	if errors.As(err, &zerr) {
		return zerr.Fields()
	}

	return nil
}
