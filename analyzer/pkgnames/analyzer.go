package pkgnames

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	UnusedInterface = &analysis.Analyzer{
		Name:     "unused_interface",
		Doc:      "error on unused exported interfaces",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runUnusedInterface,
	}
	InitFunction = &analysis.Analyzer{
		Name:     "init_function",
		Doc:      "warn on init functions",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runInitFunction,
	}
)

func runUnusedInterface(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}

func runInitFunction(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Name.Name == "init" {
			pass.Reportf(fn.Pos(), "init functions should be avoided; use constructors instead")
		}
	})
	return nil, nil
}
