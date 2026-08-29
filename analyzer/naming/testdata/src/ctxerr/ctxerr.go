package ctxerr

import "context"

func Get(c context.Context) string { // want "context.Context parameter should be named ctx"
	return ""
}

func Parse(ctx context.Context) {}

func Handle() {
	var e error // want "error variable should be named err"
	_ = e
}

func Process() {
	var err error
	_ = err
}

// FirstError is exempt: renaming firstErr to err would shadow the
// per-iteration err below it, and stop the accumulator from being written to.
func FirstError(items []int) error {
	var firstErr error
	for _, item := range items {
		if err := validate(item); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func validate(int) error { return nil }
