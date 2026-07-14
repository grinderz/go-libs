package zerr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/grinderz/go-libs/libzap/zerr"
	"go.uber.org/zap"
)

var errTest = errors.New("x")

func hasStackField(fields []zap.Field) bool {
	for _, f := range fields {
		if f.Key == "zerr_stacktrace" {
			return true
		}
	}

	return false
}

func TestWrapKeepsOuterContext(t *testing.T) {
	t.Parallel()

	outer := fmt.Errorf("context: %w", zerr.Wrap(errTest))

	if got := zerr.Wrap(outer).Error(); got != "context: x" {
		t.Errorf("outer context lost: got %q", got)
	}
}

func TestWrapIdentity(t *testing.T) {
	t.Parallel()

	z := zerr.Wrap(errTest)
	if got := zerr.Wrap(z); any(got) != any(z) {
		t.Error("Wrap of *Error with stack must return it unchanged")
	}
}

func TestWrapAddsStackOverNoStack(t *testing.T) {
	t.Parallel()

	z := zerr.Wrap(zerr.WrapNoStack(errTest))
	if !hasStackField(z.Fields()) {
		t.Error("Wrap over WrapNoStack must add a stacktrace")
	}
}

func TestWrapNilReturnsNil(t *testing.T) {
	t.Parallel()

	// Wrap(nil) and WrapNoStack(nil) return a nil *Error.
	if got := zerr.Wrap(nil); got != nil {
		t.Errorf("Wrap(nil) = %v, want nil", got)
	}

	if got := zerr.WrapNoStack(nil); got != nil {
		t.Errorf("WrapNoStack(nil) = %v, want nil", got)
	}
}

func TestWrapTypedNilNoPanic(t *testing.T) {
	t.Parallel()

	var e *zerr.Error
	if zerr.Wrap(e) == nil {
		t.Error("Wrap(typed-nil) must return a new wrapper")
	}

	if zerr.Wrap(e, zap.String("a", "1")) == nil {
		t.Error("Wrap(typed-nil, fields) must return a new wrapper")
	}
}

func TestTypedNilMethodsSafe(t *testing.T) {
	t.Parallel()

	var typedNil *zerr.Error

	if got := typedNil.Error(); got != "" {
		t.Errorf("nil.Error() = %q, want empty", got)
	}

	if typedNil.Unwrap() != nil {
		t.Error("nil.Unwrap() must be nil")
	}

	if typedNil.Fields() != nil {
		t.Error("nil.Fields() must be nil")
	}

	if zerr.Fields(typedNil) != nil {
		t.Error("Fields(typed-nil) must be nil")
	}

	if got := zerr.Wrap(typedNil).Error(); got != "" {
		t.Errorf("Wrap(typed-nil).Error() = %q, want empty", got)
	}

	z := typedNil.WithField(zap.String("a", "1"))
	if z == nil || len(z.Fields()) != 1 {
		t.Errorf("nil.WithField() = %v", z)
	}
}

func TestFieldsCollectedThroughStdErrorWrap(t *testing.T) {
	t.Parallel()

	// zerr.Error -> standard error -> zerr.Error
	inner := zerr.Wrap(errTest, zap.String("inner", "1"))
	mid := fmt.Errorf("mid: %w", inner)
	outer := zerr.Wrap(mid, zap.String("outer", "2"))

	if got := outer.Error(); got != "mid: x" {
		t.Errorf("Error() = %q, want %q", got, "mid: x")
	}

	keys := make(map[string]bool)
	for _, f := range outer.Fields() {
		keys[f.Key] = true
	}

	for _, want := range []string{"outer", "inner", "zerr_stacktrace"} {
		if !keys[want] {
			t.Errorf("missing field %q, got %v", want, keys)
		}
	}
}

func TestFieldsDoesNotMutateCallerSlice(t *testing.T) {
	t.Parallel()

	backing := [4]zap.Field{zap.String("a", "1")}
	buf := backing[:1]

	z := zerr.WrapNoStack(errTest, buf...)
	_ = z.Fields()
	_ = z.WithField(zap.String("b", "2"), buf...)

	var zero zap.Field
	if backing[1] != zero {
		t.Errorf("caller backing array mutated: %v", backing[1])
	}
}
