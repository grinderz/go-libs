package libzap

import (
	"strings"

	"github.com/grinderz/go-libs/liberrors"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=EncodingEnum -linecomment -output encoding_enum_string.go
type EncodingEnum int

const (
	EncodingUnknown EncodingEnum = iota // unknown
	EncodingConsole EncodingEnum = iota // console
	EncodingJSON    EncodingEnum = iota // json
)

func (e *EncodingEnum) SetValue(value string) error {
	encoding := EncodingFromString(value)
	if encoding == EncodingUnknown {
		return liberrors.NewInvalidStringEntityError("encoding", value)
	}

	*e = encoding

	return nil
}

func (e EncodingEnum) MarshalText() ([]byte, error) {
	if e == EncodingUnknown {
		return nil, liberrors.NewInvalidStringEntityError("encoding", e.String())
	}

	return []byte(e.String()), nil
}

func (e *EncodingEnum) UnmarshalText(text []byte) error {
	return e.SetValue(string(text))
}

func EncodingFromString(value string) EncodingEnum {
	switch strings.ToLower(value) {
	case "console":
		return EncodingConsole
	case "json":
		return EncodingJSON
	default:
		return EncodingUnknown
	}
}
