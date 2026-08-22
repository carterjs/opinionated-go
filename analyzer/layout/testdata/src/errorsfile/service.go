package errorsfile

import "errors"

// ErrExpired is stranded outside errors.go.
var ErrExpired = errors.New("expired") // want "package declares 4 sentinel errors; move ErrExpired, ErrRevoked into errors.go"

// ErrRevoked is stranded alongside it.
var ErrRevoked = errors.New("revoked")

// Lookup returns a record.
func Lookup(name string) error {
	if name == "" {
		return ErrNotFound
	}
	return nil
}
