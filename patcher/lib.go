package patcher

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/grinderz/go-libs/libzap/zerr"
	"go.uber.org/zap"
)

var ErrPatternLengthMismatch = errors.New("pattern search/replace length mismatch")

type Pattern struct {
	Description string
	Count       int
	Search      []byte
	Replace     []byte
}

// Validate checks that Replace is exactly as long as Search: ReplaceBytes
// overwrites len(Replace) bytes at each match offset, so a longer Replace
// would corrupt bytes adjacent to the match and a shorter one would leave a
// tail of the old pattern in place.
func (p *Pattern) Validate() error {
	if len(p.Search) != len(p.Replace) {
		return zerr.Wrap(
			ErrPatternLengthMismatch,
			zap.String("pattern_description", p.Description),
			zap.Int("search_len", len(p.Search)),
			zap.Int("replace_len", len(p.Replace)),
		)
	}

	return nil
}

type Result struct {
	Path         string
	BytesPatched int
	Err          error
}

func NewResult(path string, bytesPatched int) Result {
	return Result{path, bytesPatched, nil}
}

func NewError(path string, err error) Result {
	return Result{path, 0, err}
}

func ReplaceBytes(file *os.File, offsets []int64, replace []byte) (int, error) {
	var totalReplaced int

	for _, offset := range offsets {
		replaced, err := file.WriteAt(replace, offset)
		if err != nil {
			return 0, zerr.Wrap(
				fmt.Errorf("patching file: %w", err),
				zap.Int64("offset", offset),
			)
		}

		totalReplaced += replaced
	}

	if err := file.Sync(); err != nil {
		return 0, fmt.Errorf("patched file sync: %w", err)
	}

	return totalReplaced, nil
}

func SearchBytes(reader io.Reader, find []byte, buffSize int, resultCap int) ([]int64, error) {
	result := make([]int64, 0, resultCap)

	findLen := len(find)
	if findLen == 0 {
		return result, nil
	}

	// The window keeps the last findLen-1 bytes of the previous read so
	// matches spanning a read boundary are found.
	buff := make([]byte, buffSize+findLen-1)

	var (
		offset int64 // file position of buff[0]
		carry  int
	)

	for {
		readCounter, err := reader.Read(buff[carry : carry+buffSize])
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("read buffer: %w", err)
		}

		data := buff[:carry+readCounter]

		start := 0
		for {
			ind := bytes.Index(data[start:], find)
			if ind < 0 {
				break
			}

			result = append(result, offset+int64(start+ind))
			start += ind + findLen
		}

		if err == io.EOF {
			break
		}

		// Carry at most findLen-1 unconsumed bytes: bytes before start
		// belong to already reported matches and must not match again.
		keep := min(findLen-1, len(data)-start)
		copy(buff, data[len(data)-keep:])
		offset += int64(len(data) - keep)
		carry = keep
	}

	return result, nil
}
