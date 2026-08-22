package oneerror

import "errors"

// ErrNotFound is the only sentinel in this package, so it may live here.
var ErrNotFound = errors.New("not found")

// Lookup returns a record.
func Lookup(name string) error {
	if name == "" {
		return ErrNotFound
	}
	return nil
}
