package contextwithnotassigned

import (
	"context"
	"time"
)

func Example() {
	ctx := context.Background()

	// Good - return value is assigned
	ctx2, cancel := context.WithCancel(ctx)
	defer cancel()
	_ = ctx2

	ctx3, cancel2 := context.WithTimeout(ctx, 0)
	defer cancel2()
	_ = ctx3

	ctx4, cancel3 := context.WithDeadline(ctx, time.Time{})
	defer cancel3()
	_ = ctx4

	ctx5, cancel4 := context.WithCancelCause(ctx)
	defer cancel4(nil)
	_ = ctx5

	// Bad - return value not assigned
	context.WithCancel(ctx) // want `context.WithCancel return value must be assigned`

	context.WithTimeout(ctx, 0) // want `context.WithTimeout return value must be assigned`

	context.WithDeadline(ctx, time.Time{}) // want `context.WithDeadline return value must be assigned`

	context.WithCancelCause(ctx) // want `context.WithCancelCause return value must be assigned`

	// Good - other functions
	v := context.WithValue(ctx, "key", "value")
	_ = v
}
