package structs

import (
	"golang.org/x/tools/go/analysis"
)

var (
	ExportedFieldsWithMethods = &analysis.Analyzer{
		Name: "exported_fields_with_methods",
		Doc:  "error on exported fields in structs with methods",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	BooleanParameters = &analysis.Analyzer{
		Name: "boolean_parameters",
		Doc:  "error on boolean parameters",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	NoConstructorWithUnexportedFields = &analysis.Analyzer{
		Name: "no_constructor_with_unexported_fields",
		Doc:  "warn when struct has unexported fields but no constructor",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	GetenvOutsideMain = &analysis.Analyzer{
		Name: "getenv_outside_main",
		Doc:  "error on os.Getenv outside main",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	GlobalSlogFunctions = &analysis.Analyzer{
		Name: "global_slog_functions",
		Doc:  "error on global slog function calls",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	AnyInExportedAPI = &analysis.Analyzer{
		Name: "any_in_exported_api",
		Doc:  "warn on any/interface{} in exported APIs",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	FunctionTooLong = &analysis.Analyzer{
		Name: "function_too_long",
		Doc:  "warn on functions longer than 60 lines",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
)
