package libzap

import (
	"strings"

	"github.com/grinderz/go-libs/liberrors"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=PresetEnum -linecomment -output preset_enum_string.go
type PresetEnum int //nolint:recvcheck

const (
	PresetUnknown     PresetEnum = iota // unknown
	PresetDevelopment PresetEnum = iota // development
	PresetProduction  PresetEnum = iota // production
)

func (e *PresetEnum) SetValue(value string) error {
	preset := PresetFromString(value)
	if preset == PresetUnknown {
		return liberrors.NewInvalidStringEntityError("preset", value)
	}

	*e = preset

	return nil
}

func (e PresetEnum) MarshalText() ([]byte, error) {
	if e == PresetUnknown {
		return nil, liberrors.NewInvalidStringEntityError("preset", e.String())
	}

	return []byte(e.String()), nil
}

func (e *PresetEnum) UnmarshalText(text []byte) error {
	return e.SetValue(string(text))
}

func PresetFromString(value string) PresetEnum {
	switch strings.ToLower(value) {
	case "development":
		return PresetDevelopment
	case "production":
		return PresetProduction
	default:
		return PresetUnknown
	}
}
