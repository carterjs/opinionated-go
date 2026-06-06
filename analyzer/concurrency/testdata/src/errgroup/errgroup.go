package errgroup

import (
	"golang.org/x/sync/errgroup" // want "errgroup is banned"
)

var _ = errgroup.Group{}
