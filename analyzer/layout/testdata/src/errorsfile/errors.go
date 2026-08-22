package errorsfile

import "errors"

// ErrNotFound is returned when a record is missing.
var ErrNotFound = errors.New("not found")

// ErrConflict is returned when a write races another.
var ErrConflict = errors.New("conflict")
