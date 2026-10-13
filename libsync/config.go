package libsync

import (
	"time"
)

type LogConfig struct {
	Debug     bool          `yaml:"debug"     env:"DEBUG"     env-default:"false" env-description:"Enable holder tracking and slow-lock logging for sync (adds runtime.Caller overhead per lock)."`
	Threshold time.Duration `yaml:"threshold" env:"THRESHOLD" env-default:"100ms" env-description:"Sync logs threshold."`
}

type Config struct {
	Log LogConfig `yaml:"log" env-prefix:"LOG__"`
}
