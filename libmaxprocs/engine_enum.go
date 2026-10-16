package libmaxprocs

import (
	"github.com/grinderz/go-libs/libenum"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=EngineEnum -linecomment -output engine_enum_string.go
type EngineEnum int //nolint:recvcheck

const (
	EngineUnknown  EngineEnum = iota // unknown
	EngineDisabled EngineEnum = iota // disabled
	EngineAuto     EngineEnum = iota // auto
	EngineDirect   EngineEnum = iota // direct
)

var engineNames = map[string]EngineEnum{ //nolint:gochecknoglobals
	"disabled": EngineDisabled,
	"auto":     EngineAuto,
	"direct":   EngineDirect,
}

func (e *EngineEnum) SetValue(value string) error {
	return libenum.SetValue(e, "libmaxprocs_engine", value, EngineFromString)
}

func (e EngineEnum) MarshalText() ([]byte, error) {
	return libenum.MarshalText(e, "libmaxprocs_engine")
}

func (e *EngineEnum) UnmarshalText(text []byte) error {
	return e.SetValue(string(text))
}

func EngineFromString(value string) EngineEnum {
	return libenum.FromString(engineNames, value)
}
