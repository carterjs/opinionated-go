package testing_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	analysistesting "github.com/carterjs/opinionated-go/analyzer/testing"
)

func TestContextBackgroundInTest(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), analysistesting.ContextBackgroundInTest, "ctxbg")
}

func TestFmtPrintInTest(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), analysistesting.FmtPrintInTest, "fmtprint")
}

func TestStdoutStderrInTest(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), analysistesting.StdoutStderrInTest, "stdout")
}
