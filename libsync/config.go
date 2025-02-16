package libsync

import (
	"time"
)

type LogConfig struct {
	Debug     bool          `yaml:"debug"     env:"DEBUG"     env-default:"false" env-description:"Enable logging for sync."`
	Threshold time.Duration `yaml:"threshold" env:"THRESHOLD" env-default:"100ms" env-description:"Sync logs threshold."`
}

type Config struct {
	Log LogConfig `yaml:"log" env-prefix:"LOG_"`
}
