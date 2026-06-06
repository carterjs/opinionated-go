package ctxbg_test

import (
	"context"
	"testing"
)

func TestSomething(t *testing.T) {
	ctx := context.Background() // want "use t.Context()"
	_ = ctx

	ctx2 := context.TODO() // want "use t.Context()"
	_ = ctx2

	ctx3 := t.Context()
	_ = ctx3
}
