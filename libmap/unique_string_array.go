package libmap

import (
	"maps"
	"slices"
	"strings"
)

type UniqueStringArray map[string]struct{}

func (a UniqueStringArray) Set(s string) error {
	a[s] = struct{}{}
	return nil
}

// String returns the keys sorted, so the output is deterministic.
func (a UniqueStringArray) String() string {
	return strings.Join(slices.Sorted(maps.Keys(a)), "|")
}
