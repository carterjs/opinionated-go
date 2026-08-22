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

// Repeated roots its own context more than once, and is reported once.
func Repeated() error {
	ctx := context.Background() // want "prefer accepting a context.Context from the caller; Repeated roots its own context 2 times"
	if err := Fetch(ctx); err != nil {
		return err
	}
	return Fetch(context.TODO())
}
