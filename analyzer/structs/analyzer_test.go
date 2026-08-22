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

func TestExportedFieldsWithMethods(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), structs.ExportedFieldsWithMethods, "exportedfields")
}

func TestGetenvOutsideMain(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), structs.GetenvOutsideMain, "getenvoutsideimain")
}

func TestGlobalSlogFunctions(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), structs.GlobalSlogFunctions, "globalslog")
}

func TestAnyInExportedAPI(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), structs.AnyInExportedAPI, "anyinapi")
}

func TestFunctionTooLong(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), structs.FunctionTooLong, "functiontoolong")
}

func TestTooManyParameters(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), structs.TooManyParameters, "toomanyparams")
}

func TestInterfaceTooLarge(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), structs.InterfaceTooLarge, "interfacetoolarge")
}

func TestNoConstructorWithUnexportedFields(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), structs.NoConstructorWithUnexportedFields, "noconstructorunexported")
}

func TestMagicNumbers(t *testing.T) {
	t.Parallel()
	analysistest.Run(t, analysistest.TestData(), structs.MagicNumbers, "magicnumbers")
}
