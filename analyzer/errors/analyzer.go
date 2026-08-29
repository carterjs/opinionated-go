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
	// NakedErrorReturn errors on an exported function or method returning an
	// [error] without wrapping. Unexported functions are exempt: their caller
	// is the boundary and adds the context. So are methods on an unexported
	// receiver type and test files: a capitalized method on an internal type
	// (an io.Reader wrapper, a test fake) is not part of the package's public
	// API just because Go requires capitalization to satisfy an interface.
	NakedErrorReturn = &analysis.Analyzer{
		Name:     "naked_error_return",
		Doc:      "error on exported functions and methods, on exported receiver types, returning an error without wrapping",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNakedErrorReturn,
	}

	// StringErrorMatching errors on string matching against error messages.
	StringErrorMatching = &analysis.Analyzer{
		Name:     "string_error_matching",
		Doc:      "error on string matching error messages",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runStringErrorMatching,
	}

	// ErrorNotLast errors when [error] is not the last return value.
	ErrorNotLast = &analysis.Analyzer{
		Name:     "error_not_last",
		Doc:      "error when error is not last return value",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runErrorNotLast,
	}

	// NamedReturnValues errors on named return values.
	NamedReturnValues = &analysis.Analyzer{
		Name:     "named_return_values",
		Doc:      "error on named return values",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNamedReturnValues,
	}

	// PanicInNonMain errors on [builtin.panic] in non-main packages.
	PanicInNonMain = &analysis.Analyzer{
		Name:     "panic_in_non_main",
		Doc:      "error on panic in non-main packages",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runPanicInNonMain,
	}

	// SentinelNotAtPackageLevel errors on sentinel errors not declared at package level.
	SentinelNotAtPackageLevel = &analysis.Analyzer{
		Name:     "sentinel_not_at_package_level",
		Doc:      "error on sentinel errors not at package level",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runSentinelNotAtPackageLevel,
	}
)

// runNakedErrorReturn reports an error handed back to a caller outside the
// package without saying what the package was doing when it failed. Wrapping
// earns its keep at the boundary: an unexported helper's caller is a few lines
// away and adds the context itself, so requiring %w there is ceremony. An
// exported function is the last place that knows what the operation was.
func runNakedErrorReturn(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Body == nil || !isExported(fn.Name.Name) {
			return
		}
		if strings.HasSuffix(pass.Fset.Position(fn.Pos()).Filename, "_test.go") {
			return
		}
		if fn.Recv != nil && len(fn.Recv.List) > 0 {
			if receiver := receiverTypeName(fn.Recv.List[0].Type); receiver != "" && !isExported(receiver) {
				return
			}
		}

		// One finding per function. A function that hands err back in four
		// places has one habit, not four defects.
		var naked int
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			ret, ok := inner.(*ast.ReturnStmt)
			if !ok {
				return true
			}
			for _, result := range ret.Results {
				if ident, ok := result.(*ast.Ident); ok && ident.Name == "err" {
					naked++
				}
			}
			return true
		})
		if naked == 0 {
			return
		}
		if naked == 1 {
			pass.Reportf(fn.Name.Pos(), "%s returns an error without wrapping: use fmt.Errorf with %%w", fn.Name.Name)
			return
		}
		pass.Reportf(fn.Name.Pos(), "%s returns an error without wrapping in %d places: use fmt.Errorf with %%w", fn.Name.Name, naked)
	})
	return nil, nil
}

func isExported(name string) bool {
	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}

// receiverTypeName returns the bare type name a method hangs off.
func receiverTypeName(expr ast.Expr) string {
	switch receiver := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(receiver.X)
	case *ast.Ident:
		return receiver.Name
	case *ast.IndexExpr:
		return receiverTypeName(receiver.X)
	case *ast.IndexListExpr:
		return receiverTypeName(receiver.X)
	}
	return ""
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
