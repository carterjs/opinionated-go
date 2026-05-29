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
