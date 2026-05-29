package naming_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/carterjs/opinionated-go/naming"
)

func TestReceiverNames(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.ReceiverNames, "receiver")
}

func TestInitialismCasing(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.InitialismCasing, "initialism")
}

func TestCommonAbbreviations(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.CommonAbbreviations, "abbrev")
}

func TestContextAndErrorNaming(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.ContextAndErrorNaming, "ctxerr")
}

func TestPackageNaming(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.PackageNaming, "pkgname")
}

func TestFileNaming(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.FileNaming, "filename")
}

func TestGenericPackageNames(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.GenericPackageNames, "generic")
}
