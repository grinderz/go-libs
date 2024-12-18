// Based on https://github.com/syncthing/syncthing/blob/v1.28.1/lib/sync/sync.go MPL 2.0

package libsync

import (
	"runtime"
	"strconv"
	"strings"
)

const holderSkipCaller = 2

type IMutex interface {
	Lock()
	Unlock()
}

type IRWMutex interface {
	IMutex
	RLock()
	RUnlock()
}

type IWaitGroup interface {
	Add(delta int)
	Done()
	Wait()
}

func goID() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]

	id, err := strconv.Atoi(idField)
	if err != nil {
		return -1
	}

	return id
}
