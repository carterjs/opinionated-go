package layout_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/carterjs/opinionated-go/analyzer/layout"
)

func TestConstantsAtTop(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), layout.ConstantsAtTop, "constantsattop")
}

func TestTypeBeforeConstructor(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), layout.TypeBeforeConstructor, "typebeforeconstructor", "funcfirst")
}

func TestErrorsInErrorsFile(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), layout.ErrorsInErrorsFile, "errorsfile", "oneerror")
}
