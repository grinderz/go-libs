package libzap

import "errors"

var (
	ErrEmptyConfig          = errors.New("zap config is empty")
	ErrLoggerAlreadyDefined = errors.New("logger already defined")
)
