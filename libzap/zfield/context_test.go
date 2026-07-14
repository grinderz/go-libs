package zfield_test

import (
	"context"
	"testing"

	"github.com/grinderz/go-libs/libzap/zfield"
	"go.uber.org/zap"
)

func TestSiblingContextsDoNotShareBacking(t *testing.T) {
	t.Parallel()

	parent := zfield.Context(
		context.Background(),
		zap.String("a", "1"),
		zap.String("b", "2"),
		zap.String("c", "3"),
	)

	child1 := zfield.Context(parent, zap.String("child", "one"))
	_ = zfield.Context(parent, zap.String("child", "two"))

	fields := zfield.GetFields(child1)

	last := fields[len(fields)-1]
	if last.Key != "child" || last.String != "one" {
		t.Errorf("sibling context overwrote fields: %s=%s", last.Key, last.String)
	}

	if got := len(zfield.GetFields(parent)); got != 3 {
		t.Errorf("parent fields mutated: len %d, want 3", got)
	}
}
