package refactoring_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/carterjs/opinionated-go/analyzer/refactoring"
)

func TestInlineableFunction(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), refactoring.InlineableFunction, "inlineable")
}
