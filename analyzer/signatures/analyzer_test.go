package signatures_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/carterjs/opinionated-go/analyzer/signatures"
)

func TestAbsenceSpelling(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), signatures.AbsenceSpelling, "absence")
}

func TestRecordValueWithBool(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), signatures.RecordValueWithBool, "recordvalue")
}

func TestAdjacentSameTypeParameters(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), signatures.AdjacentSameTypeParameters, "adjacentparams")
}
