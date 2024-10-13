package libzap

import (
	"strings"

	"go.uber.org/zap"

	"github.com/grinderz/go-libs/libzap/zerr"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=PresetEnum -linecomment -output preset_enum_string.go
type PresetEnum int

const (
	PresetUnknown     PresetEnum = iota // unknown
	PresetDevelopment PresetEnum = iota // development
	PresetProduction  PresetEnum = iota // production
)

func (e *PresetEnum) SetValue(value string) error {
	preset := PresetFromString(value)
	if preset == PresetUnknown {
		return newPresetValueError(value)
	}

	*e = preset

	return nil
}

func (e PresetEnum) MarshalText() ([]byte, error) {
	if e == PresetUnknown {
		return nil, newPresetValueError(e.String())
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

type presetValueError struct {
	value string
}

func (e *presetValueError) Error() string {
	return "preset invalid value: " + e.value
}

func newPresetValueError(value string) error {
	return zerr.Wrap(&presetValueError{
		value: value,
	}, zap.String("preset", value))
}
