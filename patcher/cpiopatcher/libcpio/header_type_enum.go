package libcpio

import (
	"bytes"
	"fmt"
	"io"

	"github.com/grinderz/go-libs/libenum"
	"github.com/grinderz/go-libs/libzap/zerr"
	"go.uber.org/zap"
)

const MaxMagicSize = 6

var (
	cpioMagic = []byte{ //nolint:gochecknoglobals
		0x30, 0x37, 0x30, 0x37, 0x30, 0x31,
	}

	xzMagic = []byte{ //nolint:gochecknoglobals
		0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00,
	}

	gzMagic = []byte{ //nolint:gochecknoglobals
		0x1F, 0x8B,
	}
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=HeaderTypeEnum -linecomment -output header_type_enum_string.go
type HeaderTypeEnum int //nolint:recvcheck

const (
	HeaderTypeUnknown HeaderTypeEnum = iota // unknown
	HeaderTypeCPIO    HeaderTypeEnum = iota // cpio
	HeaderTypeXZ      HeaderTypeEnum = iota // xz
	HeaderTypeGZ      HeaderTypeEnum = iota // gz
)

var headerTypeNames = map[string]HeaderTypeEnum{ //nolint:gochecknoglobals
	"cpio": HeaderTypeCPIO,
	"xz":   HeaderTypeXZ,
	"gz":   HeaderTypeGZ,
}

func (ht *HeaderTypeEnum) SetValue(value string) error {
	return libenum.SetValue(ht, "cpio_header_type", value, HeaderTypeFromString)
}

func (ht HeaderTypeEnum) MarshalText() ([]byte, error) {
	return libenum.MarshalText(ht, "cpio_header_type")
}

func (ht *HeaderTypeEnum) UnmarshalText(text []byte) error {
	return ht.SetValue(string(text))
}

func HeaderTypeFromString(value string) HeaderTypeEnum {
	return libenum.FromString(headerTypeNames, value)
}

func HeaderTypeFromReader(r io.Reader) (HeaderTypeEnum, error) {
	buff := make([]byte, MaxMagicSize)
	if _, err := io.ReadFull(r, buff); err != nil {
		return HeaderTypeUnknown, fmt.Errorf("read reader: %w", err)
	}

	if bytes.Equal(buff, cpioMagic) {
		return HeaderTypeCPIO, nil
	}

	if bytes.Equal(buff, xzMagic) {
		return HeaderTypeXZ, nil
	}

	if bytes.Equal(buff[:len(gzMagic)], gzMagic) {
		return HeaderTypeGZ, nil
	}

	return HeaderTypeUnknown, newHeaderTypeUnsupportedFormatError(buff)
}

type headerTypeUnsupportedFormatError struct {
	format []byte
}

func (e *headerTypeUnsupportedFormatError) Error() string {
	return fmt.Sprintf("cpio header unsupported format %x", e.format)
}

func newHeaderTypeUnsupportedFormatError(format []byte) error {
	return zerr.Wrap(
		&headerTypeUnsupportedFormatError{
			format: format,
		},
		zap.ByteString("format", format),
	)
}
