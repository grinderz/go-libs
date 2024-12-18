package libsync

import (
	"fmt"
	"path/filepath"
	"runtime"
	"time"
)

type holder struct {
	at   string
	time time.Time
	goID int
}

func getHolder() holder {
	_, file, line, _ := runtime.Caller(holderSkipCaller)
	file = filepath.Join(filepath.Base(filepath.Dir(file)), filepath.Base(file))

	return holder{
		at:   fmt.Sprintf("%s:%d", file, line),
		goID: goID(),
		time: time.Now(),
	}
}

func (h holder) String() string {
	if h.at == "" {
		return "not held"
	}

	return fmt.Sprintf("at %s goid: %d for %s", h.at, h.goID, time.Since(h.time))
}
