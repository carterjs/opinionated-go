package concurrency

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	// ExportedFuncAcceptsChannel warns on exported functions accepting channels.
	ExportedFuncAcceptsChannel = &analysis.Analyzer{
		Name:     "exported_func_accepts_channel",
		Doc:      "warn on exported functions accepting channels",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runExportedFuncAcceptsChannel,
	}
	// ExportedFuncAcceptsFunc warns on exported functions accepting func parameters.
	ExportedFuncAcceptsFunc = &analysis.Analyzer{
		Name:     "exported_func_accepts_func",
		Doc:      "warn on exported functions accepting func parameters",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runExportedFuncAcceptsFunc,
	}
	// ContextNotFirstArg errors on [context.Context] parameter not being the first argument.
	ContextNotFirstArg = &analysis.Analyzer{
		Name:     "context_not_first_arg",
		Doc:      "error on context parameter not being first argument",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runContextNotFirstArg,
	}
	// ContextAsStructField errors on [context.Context] as struct field.
	ContextAsStructField = &analysis.Analyzer{
		Name:     "context_as_struct_field",
		Doc:      "error on context.Context as struct field",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runContextAsStructField,
	}
	// ContextWithNotAssigned errors on [context.With*] calls not assigning the return value.
	ContextWithNotAssigned = &analysis.Analyzer{
		Name:     "context_with_not_assigned",
		Doc:      "error on context.With* calls not assigning return value",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runContextWithNotAssigned,
	}
	// ContextBackgroundOutsideMain warns on [context.Background] and [context.TODO] outside package main. A root context is something main creates once and passes down.
	ContextBackgroundOutsideMain = &analysis.Analyzer{
		Name:     "context_background_outside_main",
		Doc:      "warn on context.Background or context.TODO outside package main",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runContextBackgroundOutsideMain,
	}
)

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

func runContextNotFirstArg(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil)}, func(node ast.Node) {
		var params *ast.FieldList
		switch n := node.(type) {
		case *ast.FuncDecl:
			params = n.Type.Params
		case *ast.FuncLit:
			params = n.Type.Params
		}

		if params == nil || len(params.List) == 0 {
			return
		}

		// Check if context is present but not first
		for i, param := range params.List {
			if typeString(param.Type) == "context.Context" && i > 0 {
				pass.Reportf(param.Pos(), "context.Context should be the first parameter")
			}
		}
	})
	return nil, nil
}

func runContextAsStructField(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.StructType)(nil)}, func(node ast.Node) {
		st := node.(*ast.StructType)
		for _, field := range st.Fields.List {
			if typeString(field.Type) == "context.Context" {
				pass.Reportf(field.Pos(), "context.Context should not be a struct field; pass it as a parameter")
			}
		}
	})
	return nil, nil
}

func runContextWithNotAssigned(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.ExprStmt)(nil)}, func(node ast.Node) {
		stmt := node.(*ast.ExprStmt)
		if call, ok := stmt.X.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "context" {
					funcName := sel.Sel.Name
					if strings.HasPrefix(funcName, "With") {
						pass.Reportf(call.Pos(), "context.%s return value must be assigned", funcName)
					}
				}
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

func typeString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return ident.Name + "." + t.Sel.Name
		}
	case *ast.StarExpr:
		return "*" + typeString(t.X)
	}
	return ""
}

// runContextBackgroundOutsideMain reports a context rooted in the package that
// uses it. main owns the root context and hands it down; a domain package that
// makes its own has quietly opted out of the caller's cancellation and
// deadline. Test files are exempt — [testing.ContextBackgroundInTest] covers
// those, where the answer is t.Context() rather than a parameter.
func runContextBackgroundOutsideMain(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() == "main" {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Body == nil || strings.HasSuffix(pass.Fset.File(fn.Pos()).Name(), "_test.go") {
			return
		}

		// One finding per function: taking a context parameter is a single
		// change, however many roots the body currently creates.
		var count int
		var first token.Pos
		var verb string
		ast.Inspect(fn.Body, func(inner ast.Node) bool {
			call, ok := inner.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if !ok || ident.Name != "context" {
				return true
			}
			if sel.Sel.Name != "Background" && sel.Sel.Name != "TODO" {
				return true
			}
			if count == 0 {
				first, verb = call.Pos(), sel.Sel.Name
			}
			count++
			return true
		})

		switch {
		case count == 0:
		case count == 1:
			pass.Reportf(first, "prefer accepting a context.Context from the caller; context.%s belongs in main", verb)
		default:
			pass.Reportf(first, "prefer accepting a context.Context from the caller; %s roots its own context %d times", fn.Name.Name, count)
		}
	})

	return nil, nil
}
