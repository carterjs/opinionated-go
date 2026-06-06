package errors_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/carterjs/opinionated-go/analyzer/errors"
)

func TestNakedErrorReturn(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), errors.NakedErrorReturn, "nakederrorreturn")
}

func TestStringErrorMatching(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), errors.StringErrorMatching, "stringerrormatching")
}

func TestSentinelNotAtPackageLevel(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), errors.SentinelNotAtPackageLevel, "sentinelnotatpackagelevel")
}

func TestNamedReturnValues(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), errors.NamedReturnValues, "namedreturnvalues")
}

func TestErrorNotLast(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), errors.ErrorNotLast, "errornotlast")
}

func TestPanicInNonMain(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), errors.PanicInNonMain, "panicinnonmain")
}
