package libcpio

import (
	"fmt"
	"io"
	"os"

	cpio "github.com/grinderz/gocpio"
)

const zeroByte = 0x00

func findZeroFooterSize(inFile *os.File, buffSize int) (int64, error) {
	buff := make([]byte, buffSize)

	var (
		index     int64
		readBytes int
		totalRead int64
		err       error
	)

	for {
		if readBytes, err = inFile.Read(buff); err != nil && err != io.EOF {
			return 0, fmt.Errorf("read file: %w", err)
		}

		totalRead += int64(readBytes)

		for _, b := range buff {
			if b != zeroByte {
				if _, err := inFile.Seek(-totalRead+index, 1); err != nil {
					return 0, fmt.Errorf("file seek: %w", err)
				}

				return index, nil
			}

			index++
		}

		if err == io.EOF {
			return 0, fmt.Errorf("read file EOF: %w", err)
		}
	}
}

func findTrailer(file *os.File) (int64, error) {
	rdr := cpio.NewReader(file)

	var (
		hdr *cpio.Header
		err error
	)

	for {
		hdr, err = rdr.Next()
		if err != nil {
			return 0, fmt.Errorf("cpio reader: %w", err)
		}

		if hdr.Name == "TRAILER!!!" {
			break
		}
	}

	if _, err := file.Seek(0, 0); err != nil {
		return 0, fmt.Errorf("cpio seek: %w", err)
	}

	return rdr.Pos(), nil
}

func cut(dst io.Writer, src *os.File) error {
	if _, err := src.Seek(0, 0); err != nil {
		return fmt.Errorf("src seek: %w", err)
	}

	i, err := findTrailer(src)
	if err != nil {
		return fmt.Errorf("find trailer: %w", err)
	}

	if _, err = io.CopyN(dst, src, i); err != nil {
		return fmt.Errorf("CopyN: %w", err)
	}

	return nil
}

func WriteHeader(dst io.Writer, reader io.Reader, footerSize int64) error {
	if _, err := io.Copy(dst, reader); err != nil {
		return fmt.Errorf("stream copy: %w", err)
	}

	if _, err := dst.Write(make([]byte, footerSize)); err != nil {
		return fmt.Errorf("write to writer: %w", err)
	}

	return nil
}

func CutHeader(inFile, cpioFile *os.File, bufferSize int) (HeaderTypeEnum, int64, error) {
	if err := cut(cpioFile, inFile); err != nil {
		return HeaderTypeUnknown, 0, fmt.Errorf("cut cpio: %w", err)
	}

	cpioZeroFooterSize, err := findZeroFooterSize(inFile, bufferSize)
	if err != nil {
		return HeaderTypeUnknown, 0, fmt.Errorf("find zero footer size: %w", err)
	}

	fileType, err := HeaderTypeFromReader(inFile)
	if err != nil {
		return HeaderTypeUnknown, 0, fmt.Errorf("recognize header type: %w", err)
	}

	return fileType, cpioZeroFooterSize, nil
}
