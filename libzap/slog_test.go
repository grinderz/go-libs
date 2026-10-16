package libzap_test

import (
	"log/slog"
	"path/filepath"
	"testing"

	"github.com/grinderz/go-libs/libzap"
	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewSlogLogger(t *testing.T) {
	t.Parallel()

	core, logs := observer.New(zapcore.InfoLevel)
	zlog := zap.New(core).Named("svc").With(zap.String("app", "test"))

	log := libzap.NewSlogLogger(zlog)
	log.Debug("dropped")
	log.WithGroup("req").Info("handled", slog.Int("code", 200))

	entries := logs.AllUntimed()
	if len(entries) != 1 {
		t.Fatalf("debug must be dropped by the zap level, got %d entries", len(entries))
	}

	entry := entries[0]
	if entry.Message != "handled" || entry.Level != zapcore.InfoLevel {
		t.Errorf("unexpected entry %q at %s", entry.Message, entry.Level)
	}

	if entry.LoggerName != "svc" {
		t.Errorf("logger name must carry over, got %q", entry.LoggerName)
	}

	fields := entry.ContextMap()
	if fields["app"] != "test" {
		t.Errorf("zap fields must carry over, got %v", fields)
	}

	req, ok := fields["req"].(map[string]any)
	if !ok || req["code"] != int64(200) {
		t.Errorf("slog group must become a namespace, got %v", fields)
	}

	if !entry.Caller.Defined || filepath.Base(entry.Caller.File) != "slog_test.go" {
		t.Errorf("caller must point to the slog call site, got %v", entry.Caller)
	}
}

func TestNewSlogLoggerOptions(t *testing.T) {
	t.Parallel()

	core, logs := observer.New(zapcore.InfoLevel)

	log := libzap.NewSlogLogger(zap.New(core), zapslog.WithCaller(false), zapslog.WithName("override"))
	log.Info("msg")

	entries := logs.AllUntimed()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].Caller.Defined {
		t.Error("opts must override the default caller")
	}

	if entries[0].LoggerName != "override" {
		t.Errorf("opts must override the default name, got %q", entries[0].LoggerName)
	}
}
