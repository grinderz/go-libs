package libsync

import (
	"fmt"
	"sync"
	"time"

	"github.com/grinderz/go-libs/libzap"
	"go.uber.org/zap"
)

var _ IWaitGroup = (*LoggedWaitGroup)(nil)

type LoggedWaitGroup struct {
	sync.WaitGroup

	cfg    *Config
	logger *zap.Logger
}

func NewLoggedWaitGroup(cfg *Config) *LoggedWaitGroup {
	return &LoggedWaitGroup{
		cfg:    cfg,
		logger: libzap.Logger().With(libzap.FieldPkg("sync_waitgroup")),
	}
}

func (wg *LoggedWaitGroup) Wait() {
	start := time.Now()

	wg.WaitGroup.Wait()

	duration := time.Since(start)
	if duration >= wg.cfg.Log.Threshold {
		holder := getHolder()
		wg.logger.Debug(
			fmt.Sprintf("waitgroup took %v at %s", duration, holder),
			zap.Duration("took_duration", duration),
			zap.Stringer("holder", holder),
		)
	}
}
