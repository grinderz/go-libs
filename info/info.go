package info

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"sync"

	"go.uber.org/zap/zapcore"
)

const (
	goOS   = runtime.GOOS
	goArch = runtime.GOARCH
)

var (
	version   = ""
	gitCommit = "unknown"              //nolint:gochecknoglobals
	buildDate = "1970-01-01T00:00:00Z" //nolint:gochecknoglobals
	goVersion = runtime.Version()      //nolint:gochecknoglobals
	numCPU    = runtime.NumCPU()       //nolint:gochecknoglobals
)

type Info struct {
	Version    string `json:"version"`
	GitCommit  string `json:"git_commit"`
	BuildDate  string `json:"build_date"`
	GoVersion  string `json:"go_version"`
	GoOS       string `json:"go_os"`
	GoArch     string `json:"go_arch"`
	GoMaxProcs int    `json:"go_max_procs"`
	NumCPU     int    `json:"num_cpu"`
}

var (
	instance Info      //nolint:gochecknoglobals
	once     sync.Once //nolint:gochecknoglobals
)

func GetInstance() Info {
	once.Do(func() {
		instance = newInfo()
	})

	return instance
}

func newInfo() Info {
	if version == "" {
		i, _ := debug.ReadBuildInfo()
		version = i.Main.Version
	}

	return Info{
		version,
		gitCommit,
		buildDate,
		goVersion,
		goOS,
		goArch,
		runtime.GOMAXPROCS(0),
		numCPU,
	}
}

func (i Info) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("version", i.Version)
	enc.AddString("git_commit", i.GitCommit)
	enc.AddString("build_date", i.BuildDate)
	enc.AddString("go_version", i.GoVersion)
	enc.AddString("go_os", i.GoOS)
	enc.AddString("go_arch", i.GoArch)
	enc.AddInt("go_max_procs", i.GoMaxProcs)
	enc.AddInt("num_cpu", i.NumCPU)

	return nil
}

func (i Info) String() string {
	return fmt.Sprintf("%#v", i)
}
