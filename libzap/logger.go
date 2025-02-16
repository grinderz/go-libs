package libzap

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/grinderz/go-libs/libzap/zerr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger //nolint:gochecknoglobals

func New(appID string, cfg *Config) (*zap.Logger, error) {
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

	if err := setLevel(presetCfg, &zcfg); err != nil {
		return nil, fmt.Errorf("set level: %w", err)
	}

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

	if err := setOutputs(appID, presetCfg, &zcfg); err != nil {
		return nil, fmt.Errorf("set outputs encoder: %w", err)
	}

	logger, err := zcfg.Build()
	if err != nil {
		return nil, fmt.Errorf("build: %w", err)
	}

	return logger, nil
}

func Setup(appID string, cfg *Config) {
	if cfg == nil {
		panic("empty zap config")
	}

	zp, err := New(appID, cfg)
	if err != nil {
		panic(err)
	}

	Logger = zp
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

func setOutputs(appID string, presetCfg *PresetConfig, zcfg *zap.Config) error {
	if len(presetCfg.Outputs) == 0 {
		return nil
	}

	outputs := make([]string, 0, len(presetCfg.Outputs))
	fileEnabled := false

	for output, enabled := range presetCfg.Outputs {
		if !enabled {
			continue
		}

		if output == OutputFile && len(appID) > 0 {
			fileEnabled = true
			continue
		}

		outputs = append(outputs, output.String())
	}

	if len(outputs) > 0 {
		zcfg.OutputPaths = outputs
		zcfg.ErrorOutputPaths = outputs
	}

	if fileEnabled {
		if err := setFileOutput(appID, presetCfg, zcfg); err != nil {
			return fmt.Errorf("set file output: %w", err)
		}
	}

	return nil
}

func setFileOutput(appID string, presetCfg *PresetConfig, zcfg *zap.Config) error {
	var dir string

	if filepath.IsLocal(presetCfg.OutputFile.Dir) {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("detect working directory: %w", err)
		}

		dir = filepath.Join(cwd, presetCfg.OutputFile.Dir)
	} else if filepath.IsAbs(presetCfg.OutputFile.Dir) {
		dir = presetCfg.OutputFile.Dir
	}

	if len(dir) > 0 {
		runTS := time.Now().Format(presetCfg.OutputFile.TimeLayout)
		location := filepath.Join(dir, fmt.Sprintf("%s-%s.log", appID, runTS))

		if len(location) > 0 {
			zcfg.OutputPaths = append(zcfg.OutputPaths, location)
		}
	}

	return nil
}

func setKeys(presetCfg *PresetConfig, zcfg *zap.Config) {
	if presetCfg.Encoding == EncodingJSON {
		zcfg.EncoderConfig.TimeKey = presetCfg.JSONTimeKey
		zcfg.EncoderConfig.MessageKey = presetCfg.JSONMessageKey
		zcfg.EncoderConfig.StacktraceKey = presetCfg.JSONStacktraceKey
		zcfg.EncoderConfig.CallerKey = presetCfg.JSONCallerKey
		zcfg.EncoderConfig.LevelKey = presetCfg.JSONLevelKey
		zcfg.EncoderConfig.FunctionKey = presetCfg.JSONFunctionKey
		zcfg.EncoderConfig.NameKey = presetCfg.JSONNameKey
	}
}
