package errors_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/carterjs/opinionated-go/analyzer/errors"
)

func TestErrorNotLast(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), errors.ErrorNotLast, "errornotlast")
}
