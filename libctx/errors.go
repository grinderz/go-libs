package libctx

import (
	"errors"
	"fmt"

	"github.com/grinderz/go-libs/libzap/zerr"
	"go.uber.org/zap"
)

var ErrNilCtx = errors.New("nil context")

type keyNotFoundInCtxError struct {
	key Key
}

func (e *keyNotFoundInCtxError) Error() string {
	return fmt.Sprintf("%s key not found in context", e.key)
}

func NewKeyNotFoundInCtxError(key Key) error {
	return zerr.Wrap(
		&keyNotFoundInCtxError{
			key,
		},
		zap.String("ctx_key", string(key)),
	)
}
