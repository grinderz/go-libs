package libio

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"

	"github.com/xi2/xz"
	"go.uber.org/zap"

	"github.com/grinderz/go-libs/libzap/zerr"
)

func CloneReader(reader io.Reader, dst string) error {
	dstFile, err := os.Create(dst)
	if err != nil {
		return zerr.Wrap(
			fmt.Errorf("clone reader create dst: %w", err),
			zap.String("clone_dst", dst),
		)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, reader); err != nil {
		return zerr.Wrap(
			fmt.Errorf("clone reader copy: %w", err),
			zap.String("clone_dst", dst),
		)
	}

	if err = dstFile.Sync(); err != nil {
		return zerr.Wrap(
			fmt.Errorf("clone reader sync dst: %w", err),
			zap.String("clone_dst", dst),
		)
	}

	return nil
}

func UnpackXZ(dst io.Writer, reader io.Reader) error {
	xzReader, err := xz.NewReader(reader, 0)
	if err != nil {
		return fmt.Errorf("unpack xz reader: %w", err)
	}

	if _, err = io.Copy(dst, xzReader); err != nil {
		return fmt.Errorf("unpack xz copy: %w", err)
	}

	return nil
}

func UnpackGZ(dst io.Writer, reader io.Reader, maxDecompressBytes int64) error {
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		return fmt.Errorf("unpack gz reader: %w", err)
	}

	defer gzReader.Close()

	written, err := io.CopyN(dst, gzReader, maxDecompressBytes)
	if err != nil {
		return fmt.Errorf("unpack gz copy: %w", err)
	}

	if written >= maxDecompressBytes {
		return zerr.Wrap(ErrUnpackMaxDecompressLimitReached, zap.Int64("written_bytes", written))
	}

	return nil
}

func PackGZ(dst io.Writer, reader io.Reader) error {
	gzWriter := gzip.NewWriter(dst)
	defer gzWriter.Close()

	if _, err := io.Copy(gzWriter, reader); err != nil {
		return fmt.Errorf("pack gz copy: %w", err)
	}

	return nil
}
