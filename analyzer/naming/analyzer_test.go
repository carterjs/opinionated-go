package naming_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/carterjs/opinionated-go/analyzer/naming"
)

func TestReceiverNames(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.ReceiverNames, "receiver")
}

func TestConsistentReceivers(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.ConsistentReceivers, "consistentreceivers")
}

func TestInitialismCasing(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.InitialismCasing, "initialism")
}

func TestSingleLetterExported(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.SingleLetterExported, "singleletterexported")
}

func TestSingleLetterVariables(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.SingleLetterVariables, "singlelettervars")
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

func TestOrphanTestFile(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.OrphanTestFile, "orphantestfile")
}

func TestGenericPackageNames(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.GenericPackageNames, "generic")
}

func TestJavaStyleGetters(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.JavaStyleGetters, "javagetters")
}

func TestStutteringNames(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.StutteringNames, "stuttering")
}

func TestShadowBuiltins(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), naming.ShadowBuiltins, "shadowbuiltins")
}
