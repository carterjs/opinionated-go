package structs_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/carterjs/opinionated-go/analyzer/structs"
)

func TestBooleanParameters(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), structs.BooleanParameters, "boolparam")
}
