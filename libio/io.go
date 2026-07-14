package libio

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/grinderz/go-libs/libzap"
	"github.com/grinderz/go-libs/libzap/zerr"
	"github.com/xi2/xz"
	"go.uber.org/zap"
)

type unpackMaxDecompressLimitReachedError struct {
	writtenBytes       int64
	maxDecompressBytes int64
}

func (e *unpackMaxDecompressLimitReachedError) Error() string {
	return fmt.Sprintf(
		"unpack max decompress limit reached: written[%d] limit[%d]",
		e.writtenBytes,
		e.maxDecompressBytes,
	)
}

func newUnpackMaxDecompressLimitReachedError(writtenBytes, maxDecompressBytes int64) error {
	return zerr.Wrap(
		&unpackMaxDecompressLimitReachedError{
			writtenBytes:       writtenBytes,
			maxDecompressBytes: maxDecompressBytes,
		},
		zap.Int64("written_bytes", writtenBytes),
		zap.Int64("max_decompress_bytes", maxDecompressBytes),
	)
}

func CloneReader(reader io.Reader, dst string) error {
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}

	defer func() {
		if err := dstFile.Close(); err != nil {
			zerr.Wrap(err).WithField(
				zap.String("dst", dst),
			).LogError(libzap.Logger(), "dst file close failed")
		}
	}()

	if _, err := io.Copy(dstFile, reader); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	if err = dstFile.Sync(); err != nil {
		return fmt.Errorf("sync dst: %w", err)
	}

	return nil
}

// copyLimited copies src into dst refusing streams larger than
// maxDecompressBytes. It reads limit+1 bytes, so io.EOF within that window
// means the stream fits (a stream of exactly the limit is accepted).
func copyLimited(dst io.Writer, src io.Reader, maxDecompressBytes int64) error {
	written, err := io.CopyN(dst, src, maxDecompressBytes+1)
	if err != nil && !errors.Is(err, io.EOF) {
		return fmt.Errorf("copy: %w", err)
	}

	if written > maxDecompressBytes {
		return newUnpackMaxDecompressLimitReachedError(written, maxDecompressBytes)
	}

	return nil
}

func UnpackXZ(dst io.Writer, reader io.Reader, maxDecompressBytes int64) error {
	xzReader, err := xz.NewReader(reader, 0)
	if err != nil {
		return fmt.Errorf("new reader: %w", err)
	}

	return copyLimited(dst, xzReader, maxDecompressBytes)
}

func UnpackGZ(dst io.Writer, reader io.Reader, maxDecompressBytes int64) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("reader: %w", err)
	}

	defer func() {
		if err := gzReader.Close(); err != nil {
			zerr.Wrap(err).LogError(libzap.Logger(), "gz reader close failed")
		}
	}()

	return copyLimited(dst, gzReader, maxDecompressBytes)
}

func PackGZ(dst io.Writer, reader io.Reader) error {
	gzWriter := gzip.NewWriter(dst)

	if _, err := io.Copy(gzWriter, reader); err != nil {
		_ = gzWriter.Close()
		return fmt.Errorf("copy: %w", err)
	}

	// Close flushes the remaining data and the gzip footer — a swallowed
	// error here means silently truncated output.
	if err := gzWriter.Close(); err != nil {
		return fmt.Errorf("gz writer close: %w", err)
	}

	return nil
}
