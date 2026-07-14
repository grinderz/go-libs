// Package libenum provides shared helpers for integer-backed enums whose
// zero value is the "unknown" sentinel.
package libenum

import (
	"fmt"
	"strings"

	"github.com/grinderz/go-libs/liberrors"
)

// FromString resolves value against names case-insensitively; a miss yields
// the zero ("unknown") enum value.
func FromString[E ~int](names map[string]E, value string) E { //nolint:ireturn // type param, not an interface
	return names[strings.ToLower(value)]
}

// SetValue parses value with fromString and stores it into dst, rejecting
// the unknown sentinel.
func SetValue[E ~int](dst *E, entity, value string, fromString func(string) E) error {
	parsed := fromString(value)
	if parsed == 0 {
		return liberrors.NewInvalidStringEntityError(entity, value)
	}

	*dst = parsed

	return nil
}

// MarshalText renders the enum, rejecting the unknown sentinel.
func MarshalText[E interface {
	~int
	fmt.Stringer
}](value E, entity string) ([]byte, error) {
	if value == 0 {
		return nil, liberrors.NewInvalidStringEntityError(entity, value.String())
	}

	return []byte(value.String()), nil
}
