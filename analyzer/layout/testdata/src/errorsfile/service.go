package errorsfile

import "errors"

// ErrExpired is stranded outside errors.go.
var ErrExpired = errors.New("expired") // want "package declares 3 sentinel errors; move ErrExpired into errors.go"

// Lookup returns a record.
func Lookup(name string) error {
	if name == "" {
		return ErrNotFound
	}
	return nil
}
