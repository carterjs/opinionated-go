// Package detachedcomment tests that comments bump up against their symbol.
package detachedcomment

import "errors"

// attached documents the constant below it.
const attached = 1

// detached floats above a blank line. // want "comment must bump up against the symbol"

const detached = 2

// Store is a thing.
type Store struct {
	name string

	// detached field comment. // want "comment must bump up against the symbol"

	size int
}

// NewStore builds a Store.
func NewStore() *Store {
	// A comment inside a body is running commentary, not documentation.

	return &Store{}
}

var errNotFound = errors.New("not found") // a trailing comment is not detached

//go:generate echo directives are not documentation

// a comment at the end of the file documents nothing. // want "comment must bump up against the symbol"
