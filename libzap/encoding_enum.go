package libzap

import (
	"github.com/grinderz/go-libs/libenum"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=EncodingEnum -linecomment -output encoding_enum_string.go
type EncodingEnum int //nolint:recvcheck

const (
	EncodingUnknown EncodingEnum = iota // unknown
	EncodingConsole EncodingEnum = iota // console
	EncodingJSON    EncodingEnum = iota // json
)

var encodingNames = map[string]EncodingEnum{ //nolint:gochecknoglobals
	"console": EncodingConsole,
	"json":    EncodingJSON,
}

func (e *EncodingEnum) SetValue(value string) error {
	return libenum.SetValue(e, "encoding", value, EncodingFromString)
}

func (e EncodingEnum) MarshalText() ([]byte, error) {
	return libenum.MarshalText(e, "encoding")
}

func (e *EncodingEnum) UnmarshalText(text []byte) error {
	return e.SetValue(string(text))
}

func EncodingFromString(value string) EncodingEnum {
	return libenum.FromString(encodingNames, value)
}
