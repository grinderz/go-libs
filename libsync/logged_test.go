package libsync_test

import (
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/grinderz/go-libs/libsync"
	"github.com/grinderz/go-libs/libzap"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	if err := libzap.SetupFromLogger(zap.NewNop()); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func exercise(t *testing.T, debug bool) {
	t.Helper()

	cfg := &libsync.Config{Log: libsync.LogConfig{Debug: debug, Threshold: 0}}

	mutex := libsync.NewLoggedMutex(cfg)
	rwMutex := libsync.NewLoggedRWMutex(cfg)
	waitGroup := libsync.NewLoggedWaitGroup(cfg)

	mutexCounter := 0
	rwCounter := 0

	var wg sync.WaitGroup
	for range 4 {
		wg.Add(1)
		waitGroup.Add(1)

		go func() {
			defer wg.Done()
			defer waitGroup.Done()

			mutex.Lock()
			mutexCounter++
			mutex.Unlock()

			rwMutex.RLock()
			_ = rwMutex.Holders()
			rwMutex.RUnlock()

			rwMutex.Lock()
			rwCounter++
			rwMutex.Unlock()
		}()
	}

	waitGroup.Wait()
	wg.Wait()

	if mutexCounter != 4 || rwCounter != 4 {
		t.Errorf("counters = %d/%d, want 4/4", mutexCounter, rwCounter)
	}
}

func TestLoggedSyncDebugOff(t *testing.T) {
	t.Parallel()
	exercise(t, false)
}

func TestLoggedSyncDebugOn(t *testing.T) {
	t.Parallel()
	exercise(t, true)
}

func TestLoggedMutexHolders(t *testing.T) {
	t.Parallel()

	cfg := &libsync.Config{Log: libsync.LogConfig{Debug: true, Threshold: 0}}

	mutex := libsync.NewLoggedMutex(cfg)
	mutex.Lock()

	if holders := mutex.Holders(); !strings.Contains(holders, "goid") {
		t.Errorf("debug mode must track the holder, got %q", holders)
	}

	mutex.Unlock()

	if holders := mutex.Holders(); holders != "not held" {
		t.Errorf("unlocked mutex holders = %q, want 'not held'", holders)
	}
}
