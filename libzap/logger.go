package libzap

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/grinderz/go-libs/liberrors"
	"github.com/grinderz/go-libs/libzap/zerr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _logger *zap.Logger //nolint:gochecknoglobals

// Logger returns the logger installed by Setup/SetupFromLogger, or a no-op
// logger before that — callers chaining .With() must not panic at package
// init time.
func Logger() *zap.Logger {
	if _logger == nil {
		return zap.NewNop()
	}

	return _logger
}

func New(appID string, cfg *Config, runtimeCfg *RuntimeConfig) (*zap.Logger, error) {
	var (
		zcfg      zap.Config
		presetCfg *PresetConfig
	)

	switch cfg.Preset {
	case PresetDevelopment:
		zcfg = zap.NewDevelopmentConfig()
		presetCfg = &cfg.Development
	case PresetUnknown:
		fallthrough
	case PresetProduction:
		zcfg = zap.NewProductionConfig()
		presetCfg = &cfg.Production
	}

	zcfg.DisableCaller = presetCfg.DisableCaller
	zcfg.DisableStacktrace = presetCfg.DisableStacktrace
	zcfg.Encoding = presetCfg.Encoding.String()
	zcfg.Development = presetCfg.Development
	zcfg.EncoderConfig.SkipLineEnding = presetCfg.SkipLineEnding
	zcfg.EncoderConfig.LineEnding = presetCfg.LineEnding
	zcfg.EncoderConfig.ConsoleSeparator = presetCfg.ConsoleSeparator

	setKeys(presetCfg, &zcfg)

	if err := setLevelEncoder(presetCfg, &zcfg); err != nil {
		return nil, fmt.Errorf("set level encoder: %w", err)
	}

	if err := setTimeEncoder(presetCfg, &zcfg); err != nil {
		return nil, fmt.Errorf("set time encoder: %w", err)
	}

	if err := setDurationEncoder(presetCfg, &zcfg); err != nil {
		return nil, fmt.Errorf("set duration encoder: %w", err)
	}

	if err := setCallerEncoder(presetCfg, &zcfg); err != nil {
		return nil, fmt.Errorf("set caller encoder: %w", err)
	}

	if err := setOutputs(appID, presetCfg, &zcfg, runtimeCfg); err != nil {
		return nil, fmt.Errorf("set outputs encoder: %w", err)
	}

	if runtimeCfg != nil {
		zcfg.Level = zap.NewAtomicLevelAt(runtimeCfg.Level)
	} else {
		if err := setLevel(presetCfg, &zcfg); err != nil {
			return nil, fmt.Errorf("set level: %w", err)
		}
	}

	logger, err := zcfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build: %w", err)
	}

	return logger, nil
}

func Setup(appID string, cfg *Config) error {
	if _logger != nil {
		return ErrLoggerAlreadyDefined
	}

	if cfg == nil {
		return ErrEmptyConfig
	}

	zp, err := New(appID, cfg, nil)
	if err != nil {
		return err
	}

	_logger = zp

	return nil
}

func SetupFromLogger(logger *zap.Logger) error {
	if _logger != nil {
		return ErrLoggerAlreadyDefined
	}

	_logger = logger

	return nil
}

func setLevel(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if presetCfg.Level == "" {
		return nil
	}

	lvl, err := zap.ParseAtomicLevel(presetCfg.Level)
	if err != nil {
		return zerr.Wrap(
			fmt.Errorf("parse atomic level: %w", err),
			zap.String("level", presetCfg.Level),
		)
	}

	zcfg.Level = lvl

	return nil
}

func setLevelEncoder(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if presetCfg.LevelEncoder == "" {
		return nil
	}

	var lvlEncoder zapcore.LevelEncoder

	if err := lvlEncoder.UnmarshalText([]byte(presetCfg.LevelEncoder)); err != nil {
		return zerr.Wrap(
			fmt.Errorf("unmarshal level encoder: %w", err),
			zap.String("level_encoder", presetCfg.LevelEncoder),
		)
	}

	zcfg.EncoderConfig.EncodeLevel = lvlEncoder

	return nil
}

func setTimeEncoder(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if len(presetCfg.TimeEncoder) > 0 {
		var tsEncoder zapcore.TimeEncoder

		if err := tsEncoder.UnmarshalText([]byte(presetCfg.TimeEncoder)); err != nil {
			return zerr.Wrap(
				fmt.Errorf("unmarshal time encoder: %w", err),
				zap.String("time_encoder", presetCfg.TimeEncoder),
			)
		}

		zcfg.EncoderConfig.EncodeTime = tsEncoder
	} else if len(presetCfg.TimeLayout) > 0 {
		zcfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(presetCfg.TimeLayout)
	}

	return nil
}

func setDurationEncoder(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if presetCfg.DurationEncoder == "" {
		return nil
	}

	var durEncoder zapcore.DurationEncoder

	if err := durEncoder.UnmarshalText([]byte(presetCfg.DurationEncoder)); err != nil {
		return zerr.Wrap(
			fmt.Errorf("unmarshal duration encoder: %w", err),
			zap.String("duration_encoder", presetCfg.DurationEncoder),
		)
	}

	zcfg.EncoderConfig.EncodeDuration = durEncoder

	return nil
}

func setCallerEncoder(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if presetCfg.CallerEncoder == "" {
		return nil
	}

	var callerEncoder zapcore.CallerEncoder

	if err := callerEncoder.UnmarshalText([]byte(presetCfg.CallerEncoder)); err != nil {
		return zerr.Wrap(
			fmt.Errorf("unmarshal caller encoder: %w", err),
			zap.String("caller_encoder", presetCfg.CallerEncoder),
		)
	}

	zcfg.EncoderConfig.EncodeCaller = callerEncoder

	return nil
}

func setOutputs(appID string, presetCfg *PresetConfig, zcfg *zap.Config, rcfg *RuntimeConfig) error {
	if len(presetCfg.Outputs) == 0 {
		return nil
	}

	outputs := make([]string, 0, len(presetCfg.Outputs))
	fileEnabled := false

	for output, enabled := range presetCfg.Outputs {
		if !enabled {
			continue
		}

		// The file path is appended by setFileOutput; the literal enum name
		// must never become an output path. Without an appID there is no
		// file name to build, so the file output is skipped.
		if output == OutputFile {
			fileEnabled = len(appID) > 0 && (rcfg == nil || rcfg.OutputFileEnabled)
			continue
		}

		outputs = append(outputs, output.String())
	}

	// Replace the preset defaults even when only the file output is enabled,
	// so logs are not duplicated to the default stderr.
	if len(outputs) > 0 || fileEnabled {
		zcfg.OutputPaths = outputs
	}

	if len(outputs) > 0 {
		zcfg.ErrorOutputPaths = outputs
	}

	if fileEnabled {
		if err := setFileOutput(appID, presetCfg, zcfg); err != nil {
			return fmt.Errorf("set file output: %w", err)
		}

		// setFileOutput skips an unset dir; never leave zero outputs.
		if len(zcfg.OutputPaths) == 0 {
			zcfg.OutputPaths = []string{OutputStderr.String()}
		}
	}

	return nil
}

func setFileOutput(appID string, presetCfg *PresetConfig, zcfg *zap.Config) error {
	// An unset dir (zero-value config) skips the file output instead of
	// failing logger construction.
	if presetCfg.OutputFile.Dir == "" {
		return nil
	}

	var dir string

	switch {
	case filepath.IsLocal(presetCfg.OutputFile.Dir):
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("detect working directory: %w", err)
		}

		dir = filepath.Join(cwd, presetCfg.OutputFile.Dir)
	case filepath.IsAbs(presetCfg.OutputFile.Dir):
		dir = presetCfg.OutputFile.Dir
	default:
		return liberrors.NewInvalidStringEntityError("output_file_dir", presetCfg.OutputFile.Dir)
	}

	runTS := time.Now().Format(presetCfg.OutputFile.TimeLayout)
	zcfg.OutputPaths = append(zcfg.OutputPaths, filepath.Join(dir, fmt.Sprintf("%s-%s.log", appID, runTS)))

	return nil
}

// omitKeySentinel is an explicit "omit this entry" config value. Unlike an
// empty key, it is distinguishable from an unset field, so it also works for
// console encoding, where keys are only presence switches.
const omitKeySentinel = "-"

// setKeys applies the configured entry keys. JSON encoding uses them as
// field names (an empty key omits the entry). Console encoding keeps the
// stock layout untouched except for the explicit "-" sentinel, because an
// unset config key is indistinguishable from an explicitly empty one.
func setKeys(presetCfg *PresetConfig, zcfg *zap.Config) {
	jsonEncoding := presetCfg.Encoding == EncodingJSON

	setKey(&zcfg.EncoderConfig.TimeKey, presetCfg.JSONTimeKey, jsonEncoding)
	setKey(&zcfg.EncoderConfig.MessageKey, presetCfg.JSONMessageKey, jsonEncoding)
	setKey(&zcfg.EncoderConfig.StacktraceKey, presetCfg.JSONStacktraceKey, jsonEncoding)
	setKey(&zcfg.EncoderConfig.CallerKey, presetCfg.JSONCallerKey, jsonEncoding)
	setKey(&zcfg.EncoderConfig.LevelKey, presetCfg.JSONLevelKey, jsonEncoding)
	setKey(&zcfg.EncoderConfig.FunctionKey, presetCfg.JSONFunctionKey, jsonEncoding)
	setKey(&zcfg.EncoderConfig.NameKey, presetCfg.JSONNameKey, jsonEncoding)
}

func setKey(dst *string, value string, jsonEncoding bool) {
	switch {
	case value == omitKeySentinel:
		*dst = zapcore.OmitKey
	case jsonEncoding:
		*dst = value
	}
}
