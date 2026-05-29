package errors

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	NakedErrorReturn = &analysis.Analyzer{
		Name:     "naked_error_return",
		Doc:      "error on naked error returns without wrapping",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNakedErrorReturn,
	}

	InlineErrorsNew = &analysis.Analyzer{
		Name:     "inline_errors_new",
		Doc:      "error on inline errors.New calls",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runInlineErrorsNew,
	}

	StringErrorMatching = &analysis.Analyzer{
		Name:     "string_error_matching",
		Doc:      "error on string matching error messages",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runStringErrorMatching,
	}

	ErrorNotLast = &analysis.Analyzer{
		Name:     "error_not_last",
		Doc:      "error when error is not last return value",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runErrorNotLast,
	}

	NamedReturnValues = &analysis.Analyzer{
		Name:     "named_return_values",
		Doc:      "error on named return values",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNamedReturnValues,
	}

	PanicInNonMain = &analysis.Analyzer{
		Name:     "panic_in_non_main",
		Doc:      "error on panic in non-main packages",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runPanicInNonMain,
	}

	SentinelNotAtPackageLevel = &analysis.Analyzer{
		Name:     "sentinel_not_at_package_level",
		Doc:      "error on sentinel errors not at package level",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runSentinelNotAtPackageLevel,
	}
)

func runNakedErrorReturn(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.ReturnStmt)(nil)}, func(node ast.Node) {
		ret := node.(*ast.ReturnStmt)
		if len(ret.Results) == 0 {
			return
		}

		for _, result := range ret.Results {
			if ident, ok := result.(*ast.Ident); ok && ident.Name == "err" {
				pass.Reportf(ident.Pos(), "return error without wrapping: use fmt.Errorf with %%w")
			}
		}
	})
	return nil, nil
}

func runInlineErrorsNew(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.ReturnStmt)(nil), (*ast.AssignStmt)(nil)}, func(node ast.Node) {
		switch n := node.(type) {
		case *ast.ReturnStmt:
			for _, result := range n.Results {
				if isErrorsNewCall(result) {
					pass.Reportf(result.Pos(), "errors.New should be assigned to a package-level var, not returned inline")
				}
			}
		case *ast.AssignStmt:
			for _, rhs := range n.Rhs {
				if isErrorsNewCall(rhs) && n.Tok == token.ASSIGN {
					pass.Reportf(rhs.Pos(), "errors.New should be assigned to a package-level var, not inline")
				}
			}
		}
	})
	return nil, nil
}

func runStringErrorMatching(pass *analysis.Pass) (interface{}, error) {
	stringFuncs := map[string]bool{
		"Contains":  true,
		"HasPrefix": true,
		"HasSuffix": true,
		"EqualFold": true,
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call := node.(*ast.CallExpr)
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "strings" && stringFuncs[sel.Sel.Name] {
				for _, arg := range call.Args {
					if callExpr, ok := arg.(*ast.CallExpr); ok {
						if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
							if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "err" && sel.Sel.Name == "Error" {
								pass.Reportf(call.Pos(), "string matching on error message; use errors.Is or errors.As")
							}
						}
					}
				}
			}
		}
	})
	return nil, nil
}

func runErrorNotLast(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil)}, func(node ast.Node) {
		var results *ast.FieldList
		switch n := node.(type) {
		case *ast.FuncDecl:
			results = n.Type.Results
		case *ast.FuncLit:
			results = n.Type.Results
		}

		if results == nil || len(results.List) < 2 {
			return
		}

		for i, field := range results.List {
			if isErrorType(field.Type) && i < len(results.List)-1 {
				pass.Reportf(field.Pos(), "error should be the last return value")
			}
		}
	})
	return nil, nil
}

func runNamedReturnValues(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil)}, func(node ast.Node) {
		var results *ast.FieldList
		switch n := node.(type) {
		case *ast.FuncDecl:
			results = n.Type.Results
		case *ast.FuncLit:
			results = n.Type.Results
		}

		if results == nil {
			return
		}

		for _, field := range results.List {
			if len(field.Names) > 0 && field.Names[0].Name != "" {
				pass.Reportf(field.Pos(), "named return values are banned; return values must be unnamed")
			}
		}
	})
	return nil, nil
}

func runPanicInNonMain(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() == "main" {
		return nil, nil
	}

	isTest := false
	if len(pass.Files) > 0 {
		fset := pass.Fset
		filename := fset.File(pass.Files[0].Pos()).Name()
		isTest = strings.HasSuffix(filename, "_test.go")
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call := node.(*ast.CallExpr)
		if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" && !isTest {
			pass.Reportf(call.Pos(), "panic is not allowed in library code; return an error instead")
		}
	})
	return nil, nil
}

func runSentinelNotAtPackageLevel(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.AssignStmt)(nil)}, func(node ast.Node) {
		assign := node.(*ast.AssignStmt)
		if assign.Tok != token.ASSIGN {
			return
		}

		for _, rhs := range assign.Rhs {
			if isErrorsNewCall(rhs) {
				pass.Reportf(assign.Pos(), "sentinel errors must be declared at package level with var, not assigned in functions")
			}
		}
	})
	return nil, nil
}

// Helpers

func isErrorsNewCall(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "errors" && sel.Sel.Name == "New" {
			return true
		}
	}
	return false
}

func isErrorType(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok && ident.Name == "error" {
		return true
	}
	return false
}
