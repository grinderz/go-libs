package libio

import (
	"compress/gzip"
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

func UnpackXZ(dst io.Writer, reader io.Reader) error {
	xzReader, err := xz.NewReader(reader, 0)
	if err != nil {
		return fmt.Errorf("new reader: %w", err)
	}

	if _, err = io.Copy(dst, xzReader); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	return nil
}

func UnpackGZ(dst io.Writer, reader io.Reader, maxDecompressBytes int64) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("reader: %w", err)
	}

	defer func() {
		if err := gzReader.Close(); err != nil {
			zerr.Wrap(err).WithField(
				zap.String("gz_reader", gzReader.Name),
			).LogError(libzap.Logger(), "gz reader close failed")
		}
	}()

	written, err := io.CopyN(dst, gzReader, maxDecompressBytes)
	if err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	if written >= maxDecompressBytes {
		return newUnpackMaxDecompressLimitReachedError(written, maxDecompressBytes)
	}

	return nil
}

func PackGZ(dst io.Writer, reader io.Reader) error {
	gzWriter := gzip.NewWriter(dst)

	defer func() {
		if err := gzWriter.Close(); err != nil {
			zerr.Wrap(err).WithField(
				zap.String("gz_writer", gzWriter.Name),
			).LogError(libzap.Logger(), "gz writer close failed")
		}
	}()

	if _, err := io.Copy(gzWriter, reader); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	return nil
}
