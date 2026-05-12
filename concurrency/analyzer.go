package concurrency

import (
	"golang.org/x/tools/go/analysis"
)

var (
	ErrGroupImport = &analysis.Analyzer{
		Name: "errgroup_import",
		Doc:  "error on errgroup imports",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	FireAndForgetGoroutine = &analysis.Analyzer{
		Name: "fire_and_forget_goroutine",
		Doc:  "warn on fire-and-forget goroutines",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	ExportedFuncAcceptsChannel = &analysis.Analyzer{
		Name: "exported_func_accepts_channel",
		Doc:  "warn on exported functions accepting channels",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	ExportedFuncAcceptsFunc = &analysis.Analyzer{
		Name: "exported_func_accepts_func",
		Doc:  "warn on exported functions accepting func parameters",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
)
