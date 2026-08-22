package ctxbg_test

import (
	"context"
	"testing"
)

func TestRootsTwice(t *testing.T) {
	ctx := context.Background() // want "use t.Context\\(\\) instead of context.Background or context.TODO \\(2 times in TestRootsTwice\\)"
	_ = ctx

	ctx2 := context.TODO()
	_ = ctx2
}

func TestRootsOnce(t *testing.T) {
	ctx := context.Background() // want "use t.Context\\(\\) instead of context.Background or context.TODO"
	_ = ctx
}

func TestUsesTestContext(t *testing.T) {
	ctx := t.Context()
	_ = ctx
}
