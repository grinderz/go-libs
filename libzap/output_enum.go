package libzap

import (
	"strings"

	"github.com/grinderz/go-libs/liberrors"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=OutputEnum -linecomment -output output_enum_string.go
type OutputEnum int //nolint:recvcheck

const (
	OutputUnknown OutputEnum = iota // unknown
	OutputStdout  OutputEnum = iota // stdout
	OutputStderr  OutputEnum = iota // stderr
	OutputFile    OutputEnum = iota // file
)

func (e *OutputEnum) SetValue(value string) error {
	output := OutputFromString(value)
	if output == OutputUnknown {
		return liberrors.NewInvalidStringEntityError("output", value)
	}

	*e = output

	return nil
}

func (e OutputEnum) MarshalText() ([]byte, error) {
	if e == OutputUnknown {
		return nil, liberrors.NewInvalidStringEntityError("output", e.String())
	}

	return []byte(e.String()), nil
}

func (e *OutputEnum) UnmarshalText(text []byte) error {
	return e.SetValue(string(text))
}

func OutputFromString(value string) OutputEnum {
	switch strings.ToLower(value) {
	case "stdout":
		return OutputStdout
	case "stderr":
		return OutputStderr
	case "file":
		return OutputFile
	default:
		return OutputUnknown
	}
}
