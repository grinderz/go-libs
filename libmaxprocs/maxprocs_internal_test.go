package libmaxprocs

import (
	"os"
	"runtime"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestSetDirectZeroValueKeepsGOMAXPROCS(t *testing.T) {
	// setDirect honors a GOMAXPROCS env var before checking the value, so the
	// test needs it unset; t.Setenv restores the original on cleanup.
	t.Setenv(maxProcsKey, "")

	if err := os.Unsetenv(maxProcsKey); err != nil {
		t.Fatal(err)
	}

	before := runtime.GOMAXPROCS(0)
	core, logs := observer.New(zapcore.DebugLevel)

	setDirect(&DirectConfig{Value: 0}, zap.New(core))

	if after := runtime.GOMAXPROCS(0); after != before {
		t.Errorf("GOMAXPROCS = %d, want unchanged %d", after, before)
	}

	entries := logs.All()
	if len(entries) != 1 {
		t.Fatalf("got %d log entries, want 1: %v", len(entries), entries)
	}

	if entries[0].Level != zapcore.WarnLevel {
		t.Errorf("level = %v, want %v", entries[0].Level, zapcore.WarnLevel)
	}

	if got := entries[0].ContextMap()["value"]; got != int64(0) {
		t.Errorf("value field = %v, want 0", got)
	}
}
