package libzap

import (
	"github.com/grinderz/go-libs/libenum"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=PresetEnum -linecomment -output preset_enum_string.go
type PresetEnum int //nolint:recvcheck

const (
	PresetUnknown     PresetEnum = iota // unknown
	PresetDevelopment PresetEnum = iota // development
	PresetProduction  PresetEnum = iota // production
)

var presetNames = map[string]PresetEnum{ //nolint:gochecknoglobals
	"development": PresetDevelopment,
	"production":  PresetProduction,
}

func (e *PresetEnum) SetValue(value string) error {
	return libenum.SetValue(e, "preset", value, PresetFromString)
}

func (e PresetEnum) MarshalText() ([]byte, error) {
	return libenum.MarshalText(e, "preset")
}

func (e *PresetEnum) UnmarshalText(text []byte) error {
	return e.SetValue(string(text))
}

func PresetFromString(value string) PresetEnum {
	return libenum.FromString(presetNames, value)
}
