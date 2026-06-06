package contextasstructfield

import "context"

// Good examples
type GoodServer struct {
	Name string
}

type GoodClient struct {
	ID int
}

// Bad examples
type BadServer struct {
	ctx context.Context // want "context.Context should not be a struct field"
}

type BadClient struct {
	Name string
	ctx  context.Context // want "context.Context should not be a struct field"
	Port int
}

type BadMultiple struct {
	ctx1 context.Context // want "context.Context should not be a struct field"
	ctx2 context.Context // want "context.Context should not be a struct field"
}
