package patcher_test

import (
	"bytes"
	"errors"
	"slices"
	"testing"

	"github.com/grinderz/go-libs/patcher"
)

func TestSearchBytes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		data     []byte
		find     []byte
		buffSize int
		want     []int64
	}{
		{
			name:     "single match",
			data:     []byte("hello world"),
			find:     []byte("world"),
			buffSize: 64,
			want:     []int64{6},
		},
		{
			name:     "match spans buffer boundary",
			data:     append(bytes.Repeat([]byte{0x01}, 6), []byte("needle")...),
			find:     []byte("needle"),
			buffSize: 8,
			want:     []int64{6},
		},
		{
			name:     "self overlapping prefix",
			data:     []byte("aaab"),
			find:     []byte("aab"),
			buffSize: 64,
			want:     []int64{1},
		},
		{
			name:     "multiple matches",
			data:     []byte("abc--abc--abc"),
			find:     []byte("abc"),
			buffSize: 4,
			want:     []int64{0, 5, 10},
		},
		{
			name:     "no match",
			data:     []byte("abcdef"),
			find:     []byte("xyz"),
			buffSize: 4,
			want:     []int64{},
		},
		{
			name:     "match consumed at boundary not recounted via carry",
			data:     []byte("aaab"),
			find:     []byte("aa"),
			buffSize: 2,
			want:     []int64{0},
		},
		{
			name:     "pattern bytes in stale buffer tail not matched",
			data:     append([]byte("needle"), bytes.Repeat([]byte{0x02}, 3)...),
			find:     []byte("needle"),
			buffSize: 64,
			want:     []int64{0},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got, err := patcher.SearchBytes(
				bytes.NewReader(testCase.data),
				testCase.find,
				testCase.buffSize,
				len(testCase.want),
			)
			if err != nil {
				t.Fatal(err)
			}

			if !slices.Equal(got, testCase.want) {
				t.Fatalf("offsets: got %v want %v", got, testCase.want)
			}
		})
	}
}

func TestPatternValidate(t *testing.T) {
	t.Parallel()

	valid := patcher.Pattern{Description: "ok", Count: 1, Search: []byte("abc"), Replace: []byte("xyz")}
	if err := valid.Validate(); err != nil {
		t.Errorf("equal lengths must validate: %v", err)
	}

	invalid := patcher.Pattern{Description: "bad", Count: 1, Search: []byte("abc"), Replace: []byte("toolong")}
	if err := invalid.Validate(); !errors.Is(err, patcher.ErrPatternLengthMismatch) {
		t.Errorf("length mismatch must fail, got %v", err)
	}
}
