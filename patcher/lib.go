package patcher

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"go.uber.org/zap"

	"github.com/grinderz/go-libs/libzap/zerr"
)

type Pattern struct {
	Description string
	Count       int
	Search      []byte
	Replace     []byte
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

func SearchBytes(f io.Reader, find []byte, buffSize int, resultCap int) ([]int64, error) {
	result := make([]int64, 0, resultCap)

	buff := make([]byte, buffSize)
	reader := bufio.NewReader(f)
	findLen := len(find)

	var (
		totalRead   int64
		matchIndex  int
		readCounter int
		err         error
	)

	for {
		if readCounter, err = reader.Read(buff); err != nil && err != io.EOF {
			return nil, fmt.Errorf("read buffer: %w", err)
		}

		for ind, b := range buff {
			if b != find[matchIndex] {
				matchIndex = 0
				continue
			}

			matchIndex++
			if matchIndex == findLen {
				result = append(result, totalRead-int64(matchIndex)+int64(ind)+1)
				matchIndex = 0
			}
		}

		totalRead += int64(readCounter)

		if err == io.EOF {
			break
		}
	}

	return result, nil
}
