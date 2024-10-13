package libzap

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/grinderz/go-libs/libzap/zerr"
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

	parseKeys(presetCfg, &zcfg)

	if err := parseLevel(presetCfg, &zcfg); err != nil {
		return nil, err
	}

	if err := parseLevelEncoder(presetCfg, &zcfg); err != nil {
		return nil, err
	}

	if err := parseTimeEncoder(presetCfg, &zcfg); err != nil {
		return nil, err
	}

	if err := parseDurationEncoder(presetCfg, &zcfg); err != nil {
		return nil, err
	}

	if err := parseCallerEncoder(presetCfg, &zcfg); err != nil {
		return nil, err
	}

	if err := parseOutputs(appID, presetCfg, &zcfg); err != nil {
		return nil, err
	}

	return zcfg.Build() //nolint:wrapcheck
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

func parseLevel(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if len(presetCfg.Level) == 0 {
		return nil
	}

	lvl, err := zap.ParseAtomicLevel(presetCfg.Level)
	if err != nil {
		return zerr.Wrap(
			fmt.Errorf("parse log level: %w", err),
			zap.String("level", presetCfg.Level),
		)
	}

	zcfg.Level = lvl

	return nil
}

func parseLevelEncoder(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if len(presetCfg.LevelEncoder) == 0 {
		return nil
	}

	var lvlEncoder zapcore.LevelEncoder

	if err := lvlEncoder.UnmarshalText([]byte(presetCfg.LevelEncoder)); err != nil {
		return zerr.Wrap(
			fmt.Errorf("parse log level encoder: %w", err),
			zap.String("level_encoder", presetCfg.LevelEncoder),
		)
	}

	zcfg.EncoderConfig.EncodeLevel = lvlEncoder

	return nil
}

func parseTimeEncoder(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if len(presetCfg.TimeEncoder) > 0 {
		var tsEncoder zapcore.TimeEncoder

		if err := tsEncoder.UnmarshalText([]byte(presetCfg.TimeEncoder)); err != nil {
			return zerr.Wrap(
				fmt.Errorf("parse log time encoder: %w", err),
				zap.String("time_encoder", presetCfg.TimeEncoder),
			)
		}

		zcfg.EncoderConfig.EncodeTime = tsEncoder
	} else if len(presetCfg.TimeLayout) > 0 {
		zcfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(presetCfg.TimeLayout)
	}

	return nil
}

func parseDurationEncoder(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if len(presetCfg.DurationEncoder) == 0 {
		return nil
	}

	var durEncoder zapcore.DurationEncoder

	if err := durEncoder.UnmarshalText([]byte(presetCfg.DurationEncoder)); err != nil {
		return zerr.Wrap(
			fmt.Errorf("parse log duration encoder: %w", err),
			zap.String("duration_encoder", presetCfg.DurationEncoder),
		)
	}

	zcfg.EncoderConfig.EncodeDuration = durEncoder

	return nil
}

func parseCallerEncoder(presetCfg *PresetConfig, zcfg *zap.Config) error {
	if len(presetCfg.CallerEncoder) == 0 {
		return nil
	}

	var callerEncoder zapcore.CallerEncoder

	if err := callerEncoder.UnmarshalText([]byte(presetCfg.CallerEncoder)); err != nil {
		return zerr.Wrap(
			fmt.Errorf("parse log caller encoder: %w", err),
			zap.String("caller_encoder", presetCfg.CallerEncoder),
		)
	}

	zcfg.EncoderConfig.EncodeCaller = callerEncoder

	return nil
}

func parseOutputs(appID string, presetCfg *PresetConfig, zcfg *zap.Config) error {
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
		if err := parseFileOutput(appID, presetCfg, zcfg); err != nil {
			return err
		}
	}

	return nil
}

func parseFileOutput(appID string, presetCfg *PresetConfig, zcfg *zap.Config) error {
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

func parseKeys(presetCfg *PresetConfig, zcfg *zap.Config) {
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
