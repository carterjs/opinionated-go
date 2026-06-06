package main

import (
	"golang.org/x/tools/go/analysis/multichecker"

	"github.com/carterjs/opinionated-go/analyzer/comments"
	"github.com/carterjs/opinionated-go/analyzer/concurrency"
	"github.com/carterjs/opinionated-go/analyzer/errors"
	"github.com/carterjs/opinionated-go/analyzer/naming"
	"github.com/carterjs/opinionated-go/analyzer/pkgnames"
	"github.com/carterjs/opinionated-go/analyzer/structs"
	"github.com/carterjs/opinionated-go/analyzer/testing"
)

func main() {
	multichecker.Main(
		naming.ReceiverNames,
		naming.ConsistentReceivers,
		naming.ShadowBuiltins,
		naming.InitialismCasing,
		naming.SingleLetterExported,
		naming.SingleLetterVariables,
		naming.ContextAndErrorNaming,
		naming.PackageNaming,
		naming.FileNaming,
		naming.GenericPackageNames,
		naming.JavaStyleGetters,
		naming.StutteringNames,

		comments.ExportedCommentFormat,

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
		structs.TooManyParameters,
		structs.InterfaceTooLarge,

		concurrency.ErrGroupImport,
		concurrency.ExportedFuncAcceptsChannel,
		concurrency.ExportedFuncAcceptsFunc,
		concurrency.ContextNotFirstArg,
		concurrency.ContextAsStructField,
		concurrency.ContextWithNotAssigned,

		pkgnames.InitFunction,

		testing.ContextBackgroundInTest,
		testing.FmtPrintInTest,
		testing.StdoutStderrInTest,
		testing.NoSleepInTests,
		testing.TestNameInLoop,
	)
}
