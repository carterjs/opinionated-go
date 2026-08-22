package ctxbackground

import "context"

// Fetch takes its context from the caller.
func Fetch(ctx context.Context) error {
	_ = ctx
	return nil
}

// Detached roots its own context instead of accepting one.
func Detached() error {
	ctx := context.Background() // want "prefer accepting a context.Context from the caller"
	return Fetch(ctx)
}

// Placeholder is the same problem spelled differently.
func Placeholder() error {
	ctx := context.TODO() // want "prefer accepting a context.Context from the caller"
	return Fetch(ctx)
}
