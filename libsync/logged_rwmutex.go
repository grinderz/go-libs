package libsync

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/grinderz/go-libs/liberrors"
	"github.com/grinderz/go-libs/libzap"
	"github.com/grinderz/go-libs/libzap/zerr"
	"go.uber.org/zap"
)

const unlockersChanSize = 1024

var _ IRWMutex = (*LoggedRWMutex)(nil)

type LoggedRWMutex struct {
	sync.RWMutex

	cfg *Config
	// debug is snapshotted at construction: re-reading a shared flag per
	// operation would let a runtime toggle desync Lock from its Unlock.
	debug  bool
	holder atomic.Value
	logger *zap.Logger

	readHolders    map[int][]holder
	readHoldersMut sync.Mutex

	// waitingWriters counts writers blocked in Lock: a bool would be reset
	// by the first writer to acquire while another one is still waiting.
	waitingWriters atomic.Int32
	rUnlockers     chan holder
}

func NewLoggedRWMutex(cfg *Config) *LoggedRWMutex {
	mutex := &LoggedRWMutex{
		cfg:    cfg,
		debug:  cfg.Log.Debug,
		logger: libzap.Logger().With(libzap.FieldPkg("sync_rwmutex")),
	}

	// The tracking structures cost ~48KB per mutex (1024-slot holder
	// channel), so they exist only in debug mode.
	if mutex.debug {
		mutex.readHolders = make(map[int][]holder)
		mutex.rUnlockers = make(chan holder, unlockersChanSize)
	}

	mutex.holder.Store(holder{})

	return mutex
}

func (m *LoggedRWMutex) Lock() {
	if !m.debug {
		m.RWMutex.Lock()
		return
	}

	start := time.Now()

	m.waitingWriters.Add(1)
	m.RWMutex.Lock()
	m.waitingWriters.Add(-1)

	holder := getHolder()
	m.holder.Store(holder)

	duration := holder.time.Sub(start)

	if duration >= m.cfg.Log.Threshold {
		var unlockerStrings []string

	loop:

		for {
			select {
			case holder := <-m.rUnlockers:
				unlockerStrings = append(unlockerStrings, holder.String())
			default:
				break loop
			}
		}

		m.logger.Debug(
			fmt.Sprintf(
				"rwmutex took %v to lock, locked at %s, runlockers while locking: [%s]",
				duration,
				holder.at,
				strings.Join(unlockerStrings, " | "),
			),
			zap.Duration("took_duration", duration),
			zap.String("locked_at", holder.at),
			zap.Strings("runlockers", unlockerStrings),
		)
	}
}

func (m *LoggedRWMutex) Unlock() {
	if !m.debug {
		m.RWMutex.Unlock()
		return
	}

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
				fmt.Sprintf(
					"rwmutex held for %v, locked at %s unlocked at %s",
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
	m.RWMutex.Unlock()
}

func (m *LoggedRWMutex) RLock() {
	m.RWMutex.RLock()

	if !m.debug {
		return
	}

	holder := getHolder()

	m.readHoldersMut.Lock()
	m.readHolders[holder.goID] = append(m.readHolders[holder.goID], holder)
	m.readHoldersMut.Unlock()
}

func (m *LoggedRWMutex) RUnlock() {
	if !m.debug {
		m.RWMutex.RUnlock()
		return
	}

	id := goID()

	m.readHoldersMut.Lock()

	current := m.readHolders[id]

	switch {
	case len(current) > 1:
		m.readHolders[id] = current[:len(current)-1]
	case len(current) == 1:
		delete(m.readHolders, id)
	}

	m.readHoldersMut.Unlock()

	if m.waitingWriters.Load() > 0 {
		holder := getHolder()
		select {
		case m.rUnlockers <- holder:
		default:
			m.logger.Debug(
				fmt.Sprintf("dropped holder %s as channel full", holder),
				zap.Stringer("holder", holder),
			)
		}
	}

	m.RWMutex.RUnlock()
}

func (m *LoggedRWMutex) Holders() string {
	holderRaw := m.holder.Load()

	holder, ok := holderRaw.(holder)
	if !ok {
		zerr.Wrap(
			liberrors.NewCastError("holder", holderRaw, "holder"),
		).LogError(m.logger, "")

		return ""
	}

	var output strings.Builder
	output.WriteString(holder.String())
	output.WriteString(" (writer)")

	m.readHoldersMut.Lock()

	for _, holders := range m.readHolders {
		for _, holder := range holders {
			output.WriteString(" | ")
			output.WriteString(holder.String())
			output.WriteString(" (reader)")
		}
	}

	m.readHoldersMut.Unlock()

	return output.String()
}
