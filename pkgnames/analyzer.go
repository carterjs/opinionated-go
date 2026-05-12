package pkgnames

import (
	"golang.org/x/tools/go/analysis"
)

var (
	UnusedInterface = &analysis.Analyzer{
		Name: "unused_interface",
		Doc:  "error on unused exported interfaces",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	InitFunction = &analysis.Analyzer{
		Name: "init_function",
		Doc:  "warn on init functions",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
)
