package libzap

import (
	"strings"

	"go.uber.org/zap"

	"github.com/grinderz/go-libs/libzap/zerr"
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
		return newEncodingValueError(value)
	}

	*e = encoding

	return nil
}

func (e EncodingEnum) MarshalText() ([]byte, error) {
	if e == EncodingUnknown {
		return nil, newEncodingValueError(e.String())
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

type encodingValueError struct {
	value string
}

func (e *encodingValueError) Error() string {
	return "encoding invalid value: " + e.value
}

func newEncodingValueError(value string) error {
	return zerr.Wrap(&encodingValueError{
		value: value,
	}, zap.String("encoding", value))
}
