package libsync

import (
	"time"
)

type LogConfig struct {
	Debug     bool          `yaml:"debug"     env:"DEBUG"     env-default:"false" env-description:"enable logging for sync"`
	Threshold time.Duration `yaml:"threshold" env:"THRESHOLD" env-default:"100ms" env-description:"sync logs threshold"`
}

type Config struct {
	Log LogConfig `yaml:"log" env-prefix:"LOG_"`
}
