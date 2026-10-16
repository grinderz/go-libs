package libzap

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const renamedKey = "renamed"

func TestSetKeysConsole(t *testing.T) {
	t.Parallel()

	zcfg := zap.NewProductionConfig()
	stockMessageKey := zcfg.EncoderConfig.MessageKey

	setKeys(&PresetConfig{
		Encoding:       EncodingConsole,
		JSONTimeKey:    omitKeySentinel,
		JSONMessageKey: renamedKey,
	}, &zcfg)

	if zcfg.EncoderConfig.TimeKey != zapcore.OmitKey {
		t.Errorf("console '-' must omit the entry, got %q", zcfg.EncoderConfig.TimeKey)
	}

	if zcfg.EncoderConfig.MessageKey != stockMessageKey {
		t.Errorf("console must keep stock keys, got %q", zcfg.EncoderConfig.MessageKey)
	}

	if zcfg.EncoderConfig.LevelKey == zapcore.OmitKey {
		t.Error("console unset key must not omit the entry")
	}
}

func TestSetOutputsFileOnly(t *testing.T) {
	t.Parallel()

	zcfg := zap.NewProductionConfig()
	presetCfg := &PresetConfig{
		Outputs:    map[OutputEnum]bool{OutputFile: true},
		OutputFile: OutputFileConfig{Dir: t.TempDir(), TimeLayout: "2006"},
	}

	if err := setOutputs("app", presetCfg, &zcfg, nil); err != nil {
		t.Fatal(err)
	}

	if len(zcfg.OutputPaths) != 1 {
		t.Fatalf("file-only config must produce exactly the file path, got %v", zcfg.OutputPaths)
	}

	if zcfg.OutputPaths[0] == OutputStderr.String() || zcfg.OutputPaths[0] == OutputFile.String() {
		t.Errorf("unexpected output path %q", zcfg.OutputPaths[0])
	}
}

func TestSetOutputsFileEmptyAppID(t *testing.T) {
	t.Parallel()

	zcfg := zap.NewProductionConfig()
	presetCfg := &PresetConfig{
		Outputs:    map[OutputEnum]bool{OutputFile: true, OutputStdout: true},
		OutputFile: OutputFileConfig{Dir: t.TempDir(), TimeLayout: "2006"},
	}

	if err := setOutputs("", presetCfg, &zcfg, nil); err != nil {
		t.Fatal(err)
	}

	for _, path := range zcfg.OutputPaths {
		if path == OutputFile.String() {
			t.Error("literal enum name leaked into output paths")
		}
	}
}

func TestSetKeysJSON(t *testing.T) {
	t.Parallel()

	zcfg := zap.NewProductionConfig()

	setKeys(&PresetConfig{
		Encoding:       EncodingJSON,
		JSONTimeKey:    omitKeySentinel,
		JSONMessageKey: renamedKey,
		JSONLevelKey:   "",
	}, &zcfg)

	if zcfg.EncoderConfig.TimeKey != zapcore.OmitKey {
		t.Errorf("json '-' must omit the entry, got %q", zcfg.EncoderConfig.TimeKey)
	}

	if zcfg.EncoderConfig.MessageKey != renamedKey {
		t.Errorf("json key must be applied, got %q", zcfg.EncoderConfig.MessageKey)
	}

	if zcfg.EncoderConfig.LevelKey != zapcore.OmitKey {
		t.Errorf("json empty key must omit the entry, got %q", zcfg.EncoderConfig.LevelKey)
	}
}

func TestLoggerFallbackBeforeSetup(t *testing.T) {
	t.Parallel()

	if _logger != nil {
		t.Skip("logger already set up")
	}

	logger := Logger()
	if logger == nil {
		t.Fatal("Logger() before setup must return a fallback, got nil")
	}

	if !logger.Core().Enabled(zapcore.WarnLevel) || logger.Core().Enabled(zapcore.InfoLevel) {
		t.Error("fallback logger must log warnings and above only")
	}
}

func TestSetFileOutputParentRelativeDir(t *testing.T) {
	t.Parallel()

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	presetCfg := &PresetConfig{
		OutputFile: OutputFileConfig{Dir: "../logs", TimeLayout: "2006"},
	}
	zcfg := &zap.Config{}

	if err := setFileOutput("app", presetCfg, zcfg); err != nil {
		t.Fatal(err)
	}

	want := filepath.Join(filepath.Dir(cwd), "logs", "app-"+time.Now().Format("2006")+".log")
	if len(zcfg.OutputPaths) != 1 || zcfg.OutputPaths[0] != want {
		t.Errorf("OutputPaths = %v, want [%s]", zcfg.OutputPaths, want)
	}
}
