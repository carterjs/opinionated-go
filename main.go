package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/carterjs/opinionated-go/concurrency"
	"github.com/carterjs/opinionated-go/errors"
	"github.com/carterjs/opinionated-go/naming"
	"github.com/carterjs/opinionated-go/pkgnames"
	"github.com/carterjs/opinionated-go/structs"
	"github.com/carterjs/opinionated-go/testing"
)

func main() {
	multichecker.Main(
		naming.ReceiverNames,
		naming.InitialismCasing,
		naming.CommonAbbreviations,
		naming.ContextAndErrorNaming,
		naming.PackageNaming,
		naming.FileNaming,
		naming.GenericPackageNames,

		errors.NakedErrorReturn,
		errors.InlineErrorsNew,
		errors.StringErrorMatching,
		errors.ErrorNotLast,
		errors.NamedReturnValues,
		errors.PanicInNonMain,
		errors.SentinelNotAtPackageLevel,

		structs.ExportedFieldsWithMethods,
		structs.BooleanParameters,
		structs.NoConstructorWithUnexportedFields,
		structs.GetenvOutsideMain,
		structs.GlobalSlogFunctions,
		structs.AnyInExportedAPI,
		structs.FunctionTooLong,

		concurrency.ErrGroupImport,
		concurrency.FireAndForgetGoroutine,
		concurrency.ExportedFuncAcceptsChannel,
		concurrency.ExportedFuncAcceptsFunc,

		pkgnames.UnusedInterface,
		pkgnames.InitFunction,

		testing.TestNotTableDriven,
		testing.LoopVariableNaming,
		testing.LoopVariableReassigned,
		testing.TestNamingViolation,
		testing.CoversUnexported,
		testing.OrphanedTestFile,
		testing.ContextBackgroundInTest,
		testing.FmtPrintInTest,
		testing.StdoutStderrInTest,
		testing.SubtestNameFormat,
	)
}
