package errors

import (
	"golang.org/x/tools/go/analysis"
)

var (
	NakedErrorReturn = &analysis.Analyzer{
		Name: "naked_error_return",
		Doc:  "error on naked error returns without wrapping",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	InlineErrorsNew = &analysis.Analyzer{
		Name: "inline_errors_new",
		Doc:  "error on inline errors.New calls",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	StringErrorMatching = &analysis.Analyzer{
		Name: "string_error_matching",
		Doc:  "error on string matching error messages",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	ErrorNotLast = &analysis.Analyzer{
		Name: "error_not_last",
		Doc:  "error when error is not last return value",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	NamedReturnValues = &analysis.Analyzer{
		Name: "named_return_values",
		Doc:  "error on named return values",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	PanicInNonMain = &analysis.Analyzer{
		Name: "panic_in_non_main",
		Doc:  "error on panic in non-main packages",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	SentinelNotAtPackageLevel = &analysis.Analyzer{
		Name: "sentinel_not_at_package_level",
		Doc:  "error on sentinel errors not at package level",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
)
