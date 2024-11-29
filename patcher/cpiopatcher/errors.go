package cpiopatcher

import (
	"fmt"

	"go.uber.org/zap"

	"github.com/grinderz/go-libs/libzap/zerr"
)

type invalidOffsetsLengthError struct {
	path               string
	patternDescription string
	patternIndex       int
	patternsCount      int
	offsetsLength      int
}

func (e *invalidOffsetsLengthError) Error() string {
	return fmt.Sprintf(
		"%s: pattern %d (%s) invalid offsets length offsets_len[%d] != pattern_count[%d]",
		e.path,
		e.patternIndex,
		e.patternDescription,
		e.offsetsLength,
		e.patternsCount,
	)
}

func newInvalidOffsetsLengthError(
	path, patternDescription string,
	patternIndex, patternsCount, offsetsLength int,
) error {
	return zerr.Wrap(
		&invalidOffsetsLengthError{
			path:               path,
			patternDescription: patternDescription,
			patternIndex:       patternIndex,
			patternsCount:      patternsCount,
			offsetsLength:      offsetsLength,
		},
		zap.String("path", path),
		zap.String("pattern_description", patternDescription),
		zap.Int("pattern_index", patternIndex),
		zap.Int("patterns_count", patternsCount),
		zap.Int("offsets_length", offsetsLength),
	)
}

type patternNotFoundError struct {
	path               string
	patternDescription string
	patternIndex       int
}

func (e *patternNotFoundError) Error() string {
	return fmt.Sprintf(
		"%s: pattern %d (%s) not found",
		e.path,
		e.patternIndex,
		e.patternDescription,
	)
}

func newPatternNotFoundError(path, patternDescription string, patternIndex int) error {
	return zerr.Wrap(
		&patternNotFoundError{
			path:               path,
			patternDescription: patternDescription,
			patternIndex:       patternIndex,
		},
		zap.String("path", path),
		zap.String("pattern_description", patternDescription),
		zap.Int("pattern_index", patternIndex),
	)
}
