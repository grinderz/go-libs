package libzap

import (
	"github.com/grinderz/go-libs/libenum"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=OutputEnum -linecomment -output output_enum_string.go
type OutputEnum int //nolint:recvcheck

const (
	OutputUnknown OutputEnum = iota // unknown
	OutputStdout  OutputEnum = iota // stdout
	OutputStderr  OutputEnum = iota // stderr
	OutputFile    OutputEnum = iota // file
)

var outputNames = map[string]OutputEnum{ //nolint:gochecknoglobals
	"stdout": OutputStdout,
	"stderr": OutputStderr,
	"file":   OutputFile,
}

func (e *OutputEnum) SetValue(value string) error {
	return libenum.SetValue(e, "output", value, OutputFromString)
}

func (e OutputEnum) MarshalText() ([]byte, error) {
	return libenum.MarshalText(e, "output")
}

func (e *OutputEnum) UnmarshalText(text []byte) error {
	return e.SetValue(string(text))
}

func OutputFromString(value string) OutputEnum {
	return libenum.FromString(outputNames, value)
}
