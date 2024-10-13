package cpiopatcher

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/grinderz/go-libs/libio"
	"github.com/grinderz/go-libs/libzap/zerr"
	"github.com/grinderz/go-libs/patcher"
	"github.com/grinderz/go-libs/patcher/cpiopatcher/libcpio"
)

const (
	bufferSize         = 8192
	filePerm           = 0644
	maxDecompressBytes = 524_288_000
)

type Patcher struct {
	tempDir            string
	path               string
	fileName           string
	cpioZeroFooterSize int64
	result             chan<- patcher.Result
	logger             *zap.Logger
}

func New(temp, path string, result chan<- patcher.Result, logger *zap.Logger) *Patcher {
	return &Patcher{
		tempDir:  temp,
		path:     path,
		fileName: filepath.Base(path),
		result:   result,
		logger:   logger,
	}
}

func (p *Patcher) Patch(patterns []*patcher.Pattern, backup bool) {
	var inFile, cpioFile, rawFile *os.File

	inFile, err := os.OpenFile(p.path, os.O_RDWR, filePerm)
	if err != nil {
		p.result <- patcher.NewError(p.path, err)
		return
	}

	defer inFile.Close()

	fileType, err := libcpio.HeaderTypeFromReader(inFile)
	if err != nil {
		p.result <- patcher.NewError(
			p.path,
			err,
		)

		return
	}

	if fileType == libcpio.HeaderTypeCPIO {
		p.logger.Info(
			p.path+": cut cpio header",
			zap.String("path", p.path),
			zap.String("file_type", fileType.String()),
		)

		cpioFilePath := filepath.Join(p.tempDir, p.fileName+".cpio")

		cpioFile, err := os.Create(cpioFilePath)
		if err != nil {
			p.result <- patcher.NewError(
				p.path,
				zerr.Wrap(
					fmt.Errorf("create cpio file: %w", err),
					zap.String("cpio_path", cpioFilePath),
					zap.String("file_type", fileType.String()),
				),
			)

			return
		}

		defer cpioFile.Close()

		if fileType, p.cpioZeroFooterSize, err = libcpio.CutHeader(inFile, cpioFile, bufferSize); err != nil {
			p.result <- patcher.NewError(
				p.path,
				zerr.Wrap(
					err,
					zap.String("cpio_path", cpioFilePath),
				),
			)

			return
		}
	}

	rawFilePath := filepath.Join(p.tempDir, p.fileName+".raw")
	if rawFile, err = os.Create(rawFilePath); err != nil {
		p.result <- patcher.NewError(
			p.path,
			zerr.Wrap(
				err,
				zap.String("file_type", fileType.String()),
				zap.String("raw_path", rawFilePath),
			),
		)

		return
	}

	defer rawFile.Close()

	if err := p.unpack(rawFile, inFile, fileType); err != nil {
		p.result <- patcher.NewError(
			p.path,
			zerr.Wrap(
				err,
				zap.String("file_type", fileType.String()),
				zap.String("raw_path", rawFilePath),
			),
		)

		return
	}

	replaced, err := p.patch(rawFile, patterns)
	if err != nil {
		p.result <- patcher.NewError(
			p.path,
			zerr.Wrap(
				err,
				zap.String("file_type", fileType.String()),
				zap.String("raw_path", rawFilePath),
			),
		)

		return
	}

	if replaced == 0 {
		p.result <- patcher.NewResult(p.path, 0)
		return
	}

	if err := p.pack(rawFile, inFile, cpioFile, backup); err != nil {
		p.result <- patcher.NewError(
			p.path,
			zerr.Wrap(
				err,
				zap.String("file_type", fileType.String()),
				zap.String("raw_path", rawFilePath),
			),
		)

		return
	}

	p.result <- patcher.NewResult(p.path, replaced)
}

func (p *Patcher) backup(inFile *os.File) error {
	p.logger.Info(
		p.path+": backup",
		zap.String("path", p.path),
	)

	if _, err := inFile.Seek(0, 0); err != nil {
		return fmt.Errorf("file seek: %w", err)
	}

	return libio.CloneReader(inFile, p.path+".bak")
}

func (p *Patcher) unpack(rawFile, inFile *os.File, fileType libcpio.HeaderTypeEnum) error {
	if _, err := inFile.Seek(-libcpio.MaxMagicSize, 1); err != nil {
		return fmt.Errorf("in file seek: %w", err)
	}

	switch fileType {
	case libcpio.HeaderTypeXZ:
		p.logger.Info(
			p.path+": unpack xz",
			zap.String("path", p.path),
			zap.String("file_type", fileType.String()),
		)

		if err := libio.UnpackXZ(rawFile, inFile); err != nil {
			return err
		}
	case libcpio.HeaderTypeGZ:
		p.logger.Info(
			p.path+": unpack gz",
			zap.String("path", p.path),
			zap.String("file_type", fileType.String()),
		)

		if err := libio.UnpackGZ(rawFile, inFile, maxDecompressBytes); err != nil {
			return err
		}
	case libcpio.HeaderTypeCPIO, libcpio.HeaderTypeUnknown:
		return libcpio.NewHeaderTypeValueError(fileType.String())
	}

	return nil
}

func (p *Patcher) patch(rawFile *os.File, patterns []*patcher.Pattern) (int, error) {
	var replaced int

	for patternIndex, pattern := range patterns {
		p.logger.Info(
			fmt.Sprintf("%s: search %d [%s]", p.path, patternIndex, pattern.Description),
			zap.String("path", p.path),
			zap.Int("pattern_index", patternIndex),
			zap.String("pattern", pattern.Description),
		)

		if _, err := rawFile.Seek(0, 0); err != nil {
			return 0, zerr.Wrap(
				fmt.Errorf("raw seek: %w", err),
				zap.Int("pattern_index", patternIndex),
				zap.String("pattern", pattern.Description),
			)
		}

		offsets, err := patcher.SearchBytes(rawFile, pattern.Search, bufferSize, pattern.Count)
		if err != nil {
			return 0, zerr.Wrap(
				err,
				zap.Int("pattern_index", patternIndex),
				zap.String("pattern", pattern.Description),
			)
		}

		if len(offsets) == 0 {
			return 0, zerr.Wrap(
				newPatternNotFoundError(p.path, patternIndex),
				zap.String("pattern", pattern.Description),
			)
		}

		if len(offsets) != pattern.Count {
			return 0, zerr.Wrap(
				newInvalidOffsetsLengthError(p.path, patternIndex, pattern.Count, len(offsets)),
				zap.String("pattern", pattern.Description),
			)
		}

		p.logger.Info(
			fmt.Sprintf("%s: patch %d", p.path, patternIndex),
			zap.String("path", p.path),
			zap.Int("pattern_index", patternIndex),
			zap.String("pattern", pattern.Description),
		)

		rbs, err := patcher.ReplaceBytes(rawFile, offsets, pattern.Replace)
		if err != nil {
			return 0, zerr.Wrap(
				err,
				zap.Int("pattern_index", patternIndex),
				zap.String("pattern", pattern.Description),
			)
		}

		replaced += rbs
	}

	return replaced, nil
}

func (p *Patcher) pack(rawFile, inFile, cpioFile *os.File, backup bool) error {
	if backup {
		if err := p.backup(inFile); err != nil {
			return err
		}
	}

	if _, err := rawFile.Seek(0, 0); err != nil {
		return fmt.Errorf("raw file seek: %w", err)
	}

	if _, err := inFile.Seek(0, 0); err != nil {
		return fmt.Errorf("in file seek: %w", err)
	}

	if err := inFile.Truncate(0); err != nil {
		return fmt.Errorf("in file truncate: %w", err)
	}

	if cpioFile != nil {
		if _, err := cpioFile.Seek(0, 0); err != nil {
			return fmt.Errorf("cpio file seek: %w", err)
		}

		if err := libcpio.WriteHeader(inFile, cpioFile, p.cpioZeroFooterSize); err != nil {
			return err
		}
	}

	p.logger.Info(
		p.path+": pack gz",
		zap.String("path", p.path),
	)

	return libio.PackGZ(inFile, rawFile)
}
