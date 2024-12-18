package libsync

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grinderz/go-libs/liberrors"
	"github.com/grinderz/go-libs/libzap"
	"github.com/grinderz/go-libs/libzap/zerr"
	"go.uber.org/zap"
)

var _ IMutex = (*LoggedMutex)(nil)

type LoggedMutex struct {
	sync.Mutex

	cfg    *Config
	holder atomic.Value
	logger *zap.Logger
}

func NewLoggedMutex(cfg *Config, logger *zap.Logger) *LoggedMutex {
	mutex := &LoggedMutex{
		cfg:    cfg,
		logger: logger.With(libzap.FieldPkg("sync_mutex")),
	}
	mutex.holder.Store(holder{})

	return mutex
}

func (m *LoggedMutex) Lock() {
	m.Mutex.Lock()
	m.holder.Store(getHolder())
}

func (m *LoggedMutex) Unlock() {
	currentHolderRaw := m.holder.Load()

	currentHolder, ok := currentHolderRaw.(holder)
	if !ok {
		zerr.Wrap(
			liberrors.NewCastError("holder", currentHolderRaw, "holder"),
		).LogError(m.logger, "")
	} else {
		duration := time.Since(currentHolder.time)
		if duration >= m.cfg.Log.Threshold {
			holder := getHolder()
			m.logger.Debug(
				fmt.Sprintf("mutex held for %v, locked at %s unlocked at %s",
					duration,
					currentHolder.at,
					holder.at,
				),
				zap.Duration("lock_duration", duration),
				zap.String("locked_at", currentHolder.at),
				zap.String("unlocked_at", holder.at),
			)
		}
	}

	m.holder.Store(holder{})
	m.Mutex.Unlock()
}

func (m *LoggedMutex) Holders() string {
	holderRaw := m.holder.Load()

	holder, ok := holderRaw.(holder)
	if !ok {
		zerr.Wrap(
			liberrors.NewCastError("holder", holderRaw, "holder"),
		).LogError(m.logger, "")

		return ""
	}

	return holder.String()
}
