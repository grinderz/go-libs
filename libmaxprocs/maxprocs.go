package libmaxprocs

import (
	"context"
	"fmt"
	"math"
	"os"
	"runtime"

	"github.com/grinderz/go-libs/libzap"
	"github.com/grinderz/go-libs/libzap/zerr"
	"go.uber.org/automaxprocs/maxprocs"
)

const (
	namespace   = "libmaxprocs"
	maxProcsKey = "GOMAXPROCS"
)

func Set(ctx context.Context, cfg *Config) {
	ctx = libzap.ToContext(ctx, libzap.Logger().With(libzap.FieldPkg(namespace)))

	switch cfg.Engine {
	case EngineAuto:
		setAuto(ctx, &cfg.Auto)
	case EngineDirect:
		setDirect(ctx, &cfg.Direct)
	case EngineDisabled, EngineUnknown:
		fallthrough
	default:
	}
}

func setAuto(ctx context.Context, cfg *AutoConfig) int {
	logger := libzap.FromContext(ctx)

	roundQuotaFn := func(v float64) int {
		value := int(math.Floor(v))

		if cfg.RuntimeOverhead < 1 {
			return value
		}

		maxProcs := value - cfg.RuntimeOverhead
		if maxProcs > 0 {
			logger.Info(fmt.Sprintf("maxprocs: Runtime overhead value applied GOMAXPROCS=%d", maxProcs))
			return maxProcs
		}

		return value
	}

	undoFun, err := maxprocs.Set(
		maxprocs.Logger(logger.Sugar().Infof),
		maxprocs.RoundQuotaFunc(roundQuotaFn),
	)
	if err != nil {
		zerr.Wrap(err).LogError(logger, "maxprocs: Set failed")
		undoFun()
	}

	return 0
}

func setDirect(ctx context.Context, cfg *DirectConfig) {
	logger := libzap.FromContext(ctx)

	if maxProcs, exists := os.LookupEnv(maxProcsKey); exists {
		logger.Info(fmt.Sprintf("maxprocs: Honoring GOMAXPROCS=%q as set in environment", maxProcs))
		return
	}

	logger.Info(fmt.Sprintf("maxprocs: Updating GOMAXPROCS=%d: using direct value", cfg.Value))

	runtime.GOMAXPROCS(cfg.Value)
}
