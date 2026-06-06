package contextnotfirstarg

import "context"

// Good examples - context is first or not used
func GoodFunction(ctx context.Context, x int) {}

func NoContext(x int) {}

func Good(ctx context.Context) {}

// Bad examples - context is not first
func BadFunction(x int, ctx context.Context) {} // want "context.Context should be the first parameter"

func BadWithMultiple(x int, ctx context.Context, y string) {} // want "context.Context should be the first parameter"

// Function literals
var _ = func(x int, ctx context.Context) {} // want "context.Context should be the first parameter"

var _ = func(ctx context.Context, x int) {}
