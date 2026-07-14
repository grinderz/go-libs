package libio_test

import (
	"bytes"
	"compress/gzip"
	"errors"
	"testing"

	"github.com/grinderz/go-libs/libio"
)

var errWriteFailed = errors.New("write failed")

// failAfterWriter accepts limit bytes, then fails every write.
type failAfterWriter struct {
	limit int
}

func (w *failAfterWriter) Write(data []byte) (int, error) {
	if len(data) > w.limit {
		return 0, errWriteFailed
	}

	w.limit -= len(data)

	return len(data), nil
}

func packGZ(t *testing.T, payload []byte) []byte {
	t.Helper()

	var packed bytes.Buffer

	gzWriter := gzip.NewWriter(&packed)
	if _, err := gzWriter.Write(payload); err != nil {
		t.Fatal(err)
	}

	if err := gzWriter.Close(); err != nil {
		t.Fatal(err)
	}

	return packed.Bytes()
}

func TestUnpackGZLimits(t *testing.T) {
	t.Parallel()

	const size = 1024

	packed := packGZ(t, bytes.Repeat([]byte("x"), size))

	var out bytes.Buffer
	if err := libio.UnpackGZ(&out, bytes.NewReader(packed), size); err != nil {
		t.Fatalf("exact-limit stream must unpack: %v", err)
	}

	if out.Len() != size {
		t.Fatalf("written %d, want %d", out.Len(), size)
	}

	out.Reset()

	if err := libio.UnpackGZ(&out, bytes.NewReader(packed), size-1); err == nil {
		t.Fatal("over-limit stream must fail")
	}
}

func TestPackGZReportsCloseError(t *testing.T) {
	t.Parallel()

	// Enough for the gzip header; the payload flush happens on Close and
	// must surface the write error instead of swallowing it.
	dst := &failAfterWriter{limit: 16}

	if err := libio.PackGZ(dst, bytes.NewReader([]byte("data"))); err == nil {
		t.Fatal("PackGZ must report the close/flush error")
	}
}
