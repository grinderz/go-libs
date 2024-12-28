package libmaxprocs

import (
	"strings"

	"github.com/grinderz/go-libs/liberrors"
)

//go:generate go run golang.org/x/tools/cmd/stringer@latest -type=EngineEnum -linecomment -output engine_enum_string.go
type EngineEnum int //nolint:recvcheck

const (
	EngineUnknown  EngineEnum = iota // unknown
	EngineDisabled EngineEnum = iota // disabled
	EngineAuto     EngineEnum = iota // auto
	EngineDirect   EngineEnum = iota // direct
)

func (e *EngineEnum) SetValue(value string) error {
	engine := EngineFromString(value)
	if engine == EngineUnknown {
		return liberrors.NewInvalidStringEntityError("libmaxprocs_engine", value)
	}

	*e = engine

	return nil
}

func (e EngineEnum) MarshalText() ([]byte, error) {
	if e == EngineUnknown {
		return nil, liberrors.NewInvalidStringEntityError("libmaxprocs_engine", e.String())
	}

	return []byte(e.String()), nil
}

func (e *EngineEnum) UnmarshalText(text []byte) error {
	return e.SetValue(string(text))
}

func EngineFromString(value string) EngineEnum {
	switch strings.ToLower(value) {
	case "disabled":
		return EngineDisabled
	case "auto":
		return EngineAuto
	case "direct":
		return EngineDirect
	default:
		return EngineUnknown
	}
}
