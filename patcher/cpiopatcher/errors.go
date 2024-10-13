package cpiopatcher

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/grinderz/go-libs/libzap/zerr"
)

type invalidOffsetsLengthError struct {
	path          string
	patternIndex  int
	patternsCount int
	offsetsLength int
}

func (e *invalidOffsetsLengthError) Error() string {
	return fmt.Sprintf(
		"%s: pattern %d invalid offsets length offsets_len[%d] != pattern_count[%d]",
		e.path,
		e.patternIndex,
		e.offsetsLength,
		e.patternsCount,
	)
}

func newInvalidOffsetsLengthError(path string, patternIndex, patternsCount, offsetsLength int) error {
	return zerr.Wrap(&invalidOffsetsLengthError{
		path:          path,
		patternIndex:  patternIndex,
		patternsCount: patternsCount,
		offsetsLength: offsetsLength,
	},
		zap.String("path", path),
		zap.Int("pattern_index", patternIndex),
		zap.Int("patterns_count", patternsCount),
		zap.Int("offsets_length", offsetsLength),
	)
}

type patternNotFoundError struct {
	path         string
	patternIndex int
}

func (e *patternNotFoundError) Error() string {
	return fmt.Sprintf(
		"%s: pattern %d not found",
		e.path,
		e.patternIndex,
	)
}

func newPatternNotFoundError(path string, patternIndex int) error {
	return zerr.Wrap(&patternNotFoundError{
		path:         path,
		patternIndex: patternIndex,
	},
		zap.String("path", path),
		zap.Int("pattern_index", patternIndex),
	)
}
