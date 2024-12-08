package liberrors

import (
	"errors"
	"fmt"

	"github.com/grinderz/go-libs/libzap/zerr"
	"go.uber.org/zap"
)

var ErrNotImplemented = errors.New("not implemented")

type castError struct {
	name    string
	srcType string
	dstType string
}

func (e *castError) Error() string {
	return fmt.Sprintf(
		"%s cast %s to %s",
		e.name,
		e.srcType,
		e.dstType,
	)
}

func NewCastError(name string, src any, dstType string) error {
	srcType := fmt.Sprintf("%T", src)

	return zerr.Wrap(
		&castError{
			name:    name,
			srcType: srcType,
			dstType: dstType,
		},
		zap.String("entity_name", name),
		zap.String("src_type", srcType),
		zap.String("dst_type", dstType),
	)
}

func IsCastError(err error) bool {
	if err == nil {
		return false
	}

	var e *castError

	return errors.As(err, &e)
}

type invalidEntityError struct {
	name string
}

func (e *invalidEntityError) Error() string {
	return "invalid " + e.name
}

func IsInvalidEntityError(err error) bool {
	if err == nil {
		return false
	}

	var e *invalidEntityError

	return errors.As(err, &e)
}

func NewInvalidEntityError(name string) error {
	return zerr.Wrap(
		&invalidEntityError{
			name,
		},
		zap.String("entity_name", name),
	)
}

type invalidIntEntityError struct {
	err   error
	value int
}

func (e *invalidIntEntityError) Error() string {
	return fmt.Sprintf("%v: %d", e.err, e.value)
}

func (e *invalidIntEntityError) Unwrap() error {
	return e.err
}

func NewInvalidIntEntityError(name string, value int) error {
	return zerr.Wrap(
		&invalidIntEntityError{
			NewInvalidEntityError(name),
			value,
		},
		zap.Int("entity_value", value),
	)
}

type invalidInt64EntityError struct {
	err   error
	value int64
}

func (e *invalidInt64EntityError) Error() string {
	return fmt.Sprintf("%v: %d", e.err, e.value)
}

func (e *invalidInt64EntityError) Unwrap() error {
	return e.err
}

func NewInvalidInt64EntityError(name string, value int64) error {
	return zerr.Wrap(
		&invalidInt64EntityError{
			NewInvalidEntityError(name),
			value,
		},
		zap.Int64("entity_value", value),
	)
}

type invalidStringEntityError struct {
	err   error
	value string
}

func (e *invalidStringEntityError) Error() string {
	return fmt.Sprintf("%v: %s", e.err, e.value)
}

func (e *invalidStringEntityError) Unwrap() error {
	return e.err
}

func NewInvalidStringEntityError(name string, value string) error {
	return zerr.Wrap(
		&invalidStringEntityError{
			NewInvalidEntityError(name),
			value,
		},
		zap.String("entity_value", value),
	)
}
