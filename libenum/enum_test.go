package libenum_test

import (
	"testing"

	"github.com/grinderz/go-libs/libenum"
)

type color int

const (
	colorUnknown color = iota
	colorRed
	colorGreen
)

func (c color) String() string {
	switch c {
	case colorRed:
		return redName
	case colorGreen:
		return "green"
	case colorUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

const redName = "red"

var colorNames = map[string]color{ //nolint:gochecknoglobals
	redName: colorRed,
	"green": colorGreen,
}

func colorFromString(value string) color {
	return libenum.FromString(colorNames, value)
}

func TestFromString(t *testing.T) {
	t.Parallel()

	if got := colorFromString("RED"); got != colorRed {
		t.Errorf("case-insensitive lookup failed: %v", got)
	}

	if got := colorFromString("nope"); got != colorUnknown {
		t.Errorf("miss must yield the unknown sentinel: %v", got)
	}
}

func TestSetValue(t *testing.T) {
	t.Parallel()

	var dst color
	if err := libenum.SetValue(&dst, "color", "green", colorFromString); err != nil || dst != colorGreen {
		t.Errorf("SetValue = %v, dst %v", err, dst)
	}

	if err := libenum.SetValue(&dst, "color", "nope", colorFromString); err == nil {
		t.Error("unknown value must fail")
	}

	if dst != colorGreen {
		t.Errorf("failed SetValue must not modify dst: %v", dst)
	}
}

func TestMarshalText(t *testing.T) {
	t.Parallel()

	text, err := libenum.MarshalText(colorRed, "color")
	if err != nil || string(text) != redName {
		t.Errorf("MarshalText = %q, %v", text, err)
	}

	if _, err := libenum.MarshalText(colorUnknown, "color"); err == nil {
		t.Error("unknown sentinel must fail")
	}
}
