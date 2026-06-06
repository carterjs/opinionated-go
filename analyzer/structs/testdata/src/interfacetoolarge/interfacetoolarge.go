package interfacetoolarge

// Good: 1 method
type Reader interface {
	Read() error
}

// Good: 2 methods
type ReadCloser interface {
	Read() error
	Close() error
}

// Good: 3 methods
type ReadWriteCloser interface {
	Read() error
	Write() error
	Close() error
}

// Bad: 4 methods
type TooLarge interface { // want "has 4 methods"
	Read() error
	Write() error
	Close() error
	Seek() error
}

// Bad: 5 methods
type WayTooLarge interface { // want "has 5 methods"
	Read() error
	Write() error
	Close() error
	Seek() error
	Stat() error
}
