package libzap

import (
	"strings"

	"go.uber.org/zap"

	"github.com/grinderz/go-libs/libzap/zerr"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=OutputEnum -linecomment -output output_enum_string.go
type OutputEnum int

const (
	OutputUnknown OutputEnum = iota // unknown
	OutputStdout  OutputEnum = iota // stdout
	OutputStderr  OutputEnum = iota // stderr
	OutputFile    OutputEnum = iota // file
)

func (e *OutputEnum) SetValue(value string) error {
	output := OutputFromString(value)
	if output == OutputUnknown {
		return newOutputValueError(value)
	}

	*e = output

	return nil
}

func (e OutputEnum) MarshalText() ([]byte, error) {
	if e == OutputUnknown {
		return nil, newOutputValueError(e.String())
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

type outputValueError struct {
	value string
}

func (e *outputValueError) Error() string {
	return "output invalid value: " + e.value
}

func newOutputValueError(value string) error {
	return zerr.Wrap(&outputValueError{
		value: value,
	}, zap.String("output", value))
}
