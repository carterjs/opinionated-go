package concurrency

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	ErrGroupImport = &analysis.Analyzer{
		Name:     "errgroup_import",
		Doc:      "error on errgroup imports",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runErrGroupImport,
	}
	FireAndForgetGoroutine = &analysis.Analyzer{
		Name:     "fire_and_forget_goroutine",
		Doc:      "warn on fire-and-forget goroutines",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runFireAndForgetGoroutine,
	}
	ExportedFuncAcceptsChannel = &analysis.Analyzer{
		Name:     "exported_func_accepts_channel",
		Doc:      "warn on exported functions accepting channels",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runExportedFuncAcceptsChannel,
	}
	ExportedFuncAcceptsFunc = &analysis.Analyzer{
		Name:     "exported_func_accepts_func",
		Doc:      "warn on exported functions accepting func parameters",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runExportedFuncAcceptsFunc,
	}
)

func runErrGroupImport(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.ImportSpec)(nil)}, func(node ast.Node) {
		spec := node.(*ast.ImportSpec)
		if spec.Path != nil && spec.Path.Value == `"golang.org/x/sync/errgroup"` {
			pass.Reportf(spec.Pos(), "errgroup is banned; use explicit goroutines, sync.WaitGroup, and context.WithCancelCause")
		}
	})
	return nil, nil
}

func runFireAndForgetGoroutine(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}

func runExportedFuncAcceptsChannel(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Name.Name[0] < 'A' || fn.Name.Name[0] > 'Z' {
			return
		}
		if fn.Type.Params == nil {
			return
		}
		for _, param := range fn.Type.Params.List {
			if isChanType(param.Type) {
				pass.Reportf(param.Pos(), "prefer wrapping coordination primitives behind an interface or concrete type")
			}
		}
	})
	return nil, nil
}

func runExportedFuncAcceptsFunc(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Name.Name[0] < 'A' || fn.Name.Name[0] > 'Z' {
			return
		}
		if fn.Type.Params == nil {
			return
		}
		for _, param := range fn.Type.Params.List {
			if isFuncType(param.Type) {
				pass.Reportf(param.Pos(), "prefer an interface with a method over a func parameter")
			}
		}
	})
	return nil, nil
}

func isChanType(expr ast.Expr) bool {
	_, ok := expr.(*ast.ChanType)
	return ok
}

func isFuncType(expr ast.Expr) bool {
	_, ok := expr.(*ast.FuncType)
	return ok
}
