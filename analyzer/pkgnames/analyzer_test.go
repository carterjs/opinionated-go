package pkgnames_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/carterjs/opinionated-go/analyzer/pkgnames"
)

func TestInitFunction(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), pkgnames.InitFunction, "initfunc")
}
