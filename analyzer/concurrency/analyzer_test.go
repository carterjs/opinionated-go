package concurrency_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/carterjs/opinionated-go/analyzer/concurrency"
)

func TestExportedFuncAcceptsChannel(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), concurrency.ExportedFuncAcceptsChannel, "exportedchan")
}

func TestExportedFuncAcceptsFunc(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), concurrency.ExportedFuncAcceptsFunc, "exportedfunc")
}

func TestErrGroupImport(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), concurrency.ErrGroupImport, "errgroup")
}

func TestContextNotFirstArg(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), concurrency.ContextNotFirstArg, "contextnotfirstarg")
}

func TestContextAsStructField(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), concurrency.ContextAsStructField, "contextasstructfield")
}

func TestContextWithNotAssigned(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), concurrency.ContextWithNotAssigned, "contextwithnotassigned")
}
