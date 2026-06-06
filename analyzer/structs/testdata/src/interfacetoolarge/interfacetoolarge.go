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

// Good: 4 methods
type ReadWriteCloseSeefer interface {
	Read() error
	Write() error
	Close() error
	Seek() error
}

// Good: 5 methods
type ReadWriteCloseSeeker interface {
	Read() error
	Write() error
	Close() error
	Seek() error
	Stat() error
}

// Bad: 6 methods
type TooLarge interface { // want "has 6 methods"
	Read() error
	Write() error
	Close() error
	Seek() error
	Stat() error
	Truncate() error
}

// Bad: 7 methods
type WayTooLarge interface { // want "has 7 methods"
	Read() error
	Write() error
	Close() error
	Seek() error
	Stat() error
	Truncate() error
	Sync() error
}
