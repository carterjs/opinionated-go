package pkgnames

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	// InitFunction warns on init functions.
	InitFunction = &analysis.Analyzer{
		Name:     "init_function",
		Doc:      "warn on init functions",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runInitFunction,
	}
)

func runInitFunction(pass *analysis.Pass) (interface{}, error) {
	// Skip generated/cached files
	if len(pass.Files) > 0 {
		filename := pass.Fset.File(pass.Files[0].Pos()).Name()
		if strings.Contains(filename, "/go-build/") || isGenerated(pass.Files[0]) {
			return nil, nil
		}
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Name.Name == "init" {
			pass.Reportf(fn.Pos(), "init functions should be avoided; use constructors instead")
		}
	})
	return nil, nil
}

func isGenerated(file *ast.File) bool {
	if file.Comments == nil {
		return false
	}
	for _, cg := range file.Comments {
		for _, c := range cg.List {
			if strings.Contains(c.Text, "Code generated") && strings.Contains(c.Text, "DO NOT EDIT") {
				return true
			}
		}
	}
	return false
}
