package cpiopatcher

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/grinderz/go-libs/libzap"
	"github.com/grinderz/go-libs/patcher"
	"go.uber.org/zap"
)

const (
	originalContent = "original content"
	rawContent      = "patched payload"
	// testPerm has group/other write bits, which a typical 022 umask strips
	// on O_CREATE, so the test catches a missing explicit chmod.
	testPerm = 0o666
)

func TestMain(m *testing.M) {
	if err := libzap.SetupFromLogger(zap.NewNop()); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

// writeTestFile creates path with content and exactly perm, bypassing umask.
func writeTestFile(t *testing.T, path string, content []byte, perm os.FileMode) {
	t.Helper()

	if err := os.WriteFile(path, content, perm); err != nil {
		t.Fatal(err)
	}

	if err := os.Chmod(path, perm); err != nil {
		t.Fatal(err)
	}
}

// openTestFile opens path for read-write and closes it on cleanup.
func openTestFile(t *testing.T, path string) *os.File {
	t.Helper()

	file, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = file.Close() })

	return file
}

// newTestFile writes content to a new file in dir and opens it.
func newTestFile(t *testing.T, dir, name string, content []byte) *os.File {
	t.Helper()

	path := filepath.Join(dir, name)
	writeTestFile(t, path, content, filePerm)

	return openTestFile(t, path)
}

func gzipBytes(t *testing.T, data []byte) []byte {
	t.Helper()

	var buf bytes.Buffer

	gzWriter := gzip.NewWriter(&buf)

	if _, err := gzWriter.Write(data); err != nil {
		t.Fatal(err)
	}

	if err := gzWriter.Close(); err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}

func gunzipBytes(t *testing.T, data []byte) []byte {
	t.Helper()

	gzReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatal(err)
	}

	out, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatal(err)
	}

	return out
}

func readTestFile(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func assertNotExists(t *testing.T, path string) {
	t.Helper()

	if _, err := os.Lstat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("%s must not exist, stat err: %v", path, err)
	}
}

// packFixture is an original file plus a raw file holding the patched payload.
type packFixture struct {
	patcher *Patcher
	path    string
	inFile  *os.File
	rawFile *os.File
}

func newPackFixture(t *testing.T) packFixture {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "image.gz")
	writeTestFile(t, path, []byte(originalContent), testPerm)

	return packFixture{
		patcher: New(t.TempDir(), path, nil),
		path:    path,
		inFile:  openTestFile(t, path),
		rawFile: newTestFile(t, dir, "image.raw", []byte(rawContent)),
	}
}

func TestPackReplacesOriginal(t *testing.T) {
	t.Parallel()

	fixture := newPackFixture(t)

	if err := fixture.patcher.pack(fixture.rawFile, fixture.inFile, nil, false); err != nil {
		t.Fatal(err)
	}

	if got := gunzipBytes(t, readTestFile(t, fixture.path)); string(got) != rawContent {
		t.Errorf("packed content = %q, want %q", got, rawContent)
	}

	info, err := os.Stat(fixture.path)
	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != testPerm {
		t.Errorf("mode = %v, want %v", info.Mode().Perm(), os.FileMode(testPerm))
	}

	assertNotExists(t, fixture.path+".tmp")
	assertNotExists(t, fixture.path+".bak")
}

func TestPackBackup(t *testing.T) {
	t.Parallel()

	fixture := newPackFixture(t)

	if err := fixture.patcher.pack(fixture.rawFile, fixture.inFile, nil, true); err != nil {
		t.Fatal(err)
	}

	if got := readTestFile(t, fixture.path+".bak"); string(got) != originalContent {
		t.Errorf("backup content = %q, want %q", got, originalContent)
	}

	if got := gunzipBytes(t, readTestFile(t, fixture.path)); string(got) != rawContent {
		t.Errorf("packed content = %q, want %q", got, rawContent)
	}
}

func TestPackWritesCPIOHeader(t *testing.T) {
	t.Parallel()

	const (
		header     = "CPIO-HEADER"
		footerSize = 4
	)

	fixture := newPackFixture(t)
	fixture.patcher.cpioZeroFooterSize = footerSize
	cpioFile := newTestFile(t, t.TempDir(), "image.cpio", []byte(header))

	// A non-zero offset proves pack rewinds the cpio file itself.
	if _, err := cpioFile.Seek(0, io.SeekEnd); err != nil {
		t.Fatal(err)
	}

	if err := fixture.patcher.pack(fixture.rawFile, fixture.inFile, cpioFile, false); err != nil {
		t.Fatal(err)
	}

	packed := readTestFile(t, fixture.path)

	prefix := append([]byte(header), make([]byte, footerSize)...)

	if !bytes.HasPrefix(packed, prefix) {
		t.Fatalf("packed file must start with cpio header and zero footer, got %q", packed[:min(len(packed), len(prefix))])
	}

	if got := gunzipBytes(t, packed[len(prefix):]); string(got) != rawContent {
		t.Errorf("packed content = %q, want %q", got, rawContent)
	}
}

func TestPackFailureKeepsOriginal(t *testing.T) {
	t.Parallel()

	fixture := newPackFixture(t)

	// A closed cpio file fails after the temporary output already exists.
	cpioFile := newTestFile(t, t.TempDir(), "image.cpio", []byte("header"))
	if err := cpioFile.Close(); err != nil {
		t.Fatal(err)
	}

	if err := fixture.patcher.pack(fixture.rawFile, fixture.inFile, cpioFile, false); err == nil {
		t.Fatal("pack with a closed cpio file must fail")
	}

	if got := readTestFile(t, fixture.path); string(got) != originalContent {
		t.Errorf("original content = %q, want %q", got, originalContent)
	}

	assertNotExists(t, fixture.path+".tmp")
}

func TestPackCreateOutFailureKeepsOriginal(t *testing.T) {
	t.Parallel()

	fixture := newPackFixture(t)

	// A directory in place of the temporary output makes its creation fail.
	if err := os.Mkdir(fixture.path+".tmp", 0o755); err != nil {
		t.Fatal(err)
	}

	if err := fixture.patcher.pack(fixture.rawFile, fixture.inFile, nil, false); err == nil {
		t.Fatal("pack must fail when the temporary output cannot be created")
	}

	if got := readTestFile(t, fixture.path); string(got) != originalContent {
		t.Errorf("original content = %q, want %q", got, originalContent)
	}
}

func runPatch(t *testing.T, path string, patterns []*patcher.Pattern) patcher.Result {
	t.Helper()

	result := make(chan patcher.Result, 1)
	New(t.TempDir(), path, result).Patch(patterns, false)

	return <-result
}

func TestPatchGZ(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "image.gz")
	writeTestFile(t, path, gzipBytes(t, []byte("hello OLD world OLD!")), testPerm)

	res := runPatch(t, path, []*patcher.Pattern{
		{Description: "old to new", Count: 2, Search: []byte("OLD"), Replace: []byte("NEW")},
	})
	if res.Err != nil {
		t.Fatal(res.Err)
	}

	if res.BytesPatched != 6 {
		t.Errorf("BytesPatched = %d, want 6", res.BytesPatched)
	}

	if got := gunzipBytes(t, readTestFile(t, path)); string(got) != "hello NEW world NEW!" {
		t.Errorf("patched content = %q", got)
	}

	assertNotExists(t, path+".tmp")
}

func TestPatchInvalidPatternKeepsFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "image.gz")
	original := gzipBytes(t, []byte("hello OLD world"))
	writeTestFile(t, path, original, testPerm)

	res := runPatch(t, path, []*patcher.Pattern{
		{Description: "length mismatch", Count: 1, Search: []byte("OLD"), Replace: []byte("NEWER")},
	})
	if !errors.Is(res.Err, patcher.ErrPatternLengthMismatch) {
		t.Fatalf("expected ErrPatternLengthMismatch, got %v", res.Err)
	}

	if got := readTestFile(t, path); !bytes.Equal(got, original) {
		t.Error("file must stay untouched when a pattern is invalid")
	}
}
