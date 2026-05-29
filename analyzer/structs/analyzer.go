package structs

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	ExportedFieldsWithMethods = &analysis.Analyzer{
		Name:     "exported_fields_with_methods",
		Doc:      "error on exported fields in structs with methods",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runExportedFieldsWithMethods,
	}
	BooleanParameters = &analysis.Analyzer{
		Name:     "boolean_parameters",
		Doc:      "error on boolean parameters in exported functions",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runBooleanParameters,
	}
	NoConstructorWithUnexportedFields = &analysis.Analyzer{
		Name:     "no_constructor_with_unexported_fields",
		Doc:      "warn when struct has unexported fields but no constructor",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNoConstructorWithUnexportedFields,
	}
	GetenvOutsideMain = &analysis.Analyzer{
		Name:     "getenv_outside_main",
		Doc:      "error on os.Getenv outside main",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runGetenvOutsideMain,
	}
	GlobalSlogFunctions = &analysis.Analyzer{
		Name:     "global_slog_functions",
		Doc:      "error on global slog function calls",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runGlobalSlogFunctions,
	}
	AnyInExportedAPI = &analysis.Analyzer{
		Name:     "any_in_exported_api",
		Doc:      "warn on any/interface{} in exported APIs",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runAnyInExportedAPI,
	}
	FunctionTooLong = &analysis.Analyzer{
		Name:     "function_too_long",
		Doc:      "warn on functions longer than 60 lines",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runFunctionTooLong,
	}
)

func runExportedFieldsWithMethods(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.StructType)(nil)}, func(node ast.Node) {
		st := node.(*ast.StructType)
		if st.Fields == nil || len(st.Fields.List) == 0 {
			return
		}
		for _, field := range st.Fields.List {
			if len(field.Names) > 0 && isExported(field.Names[0].Name) {
				pass.Reportf(field.Pos(), "struct with methods should not have exported fields")
			}
		}
	})
	return nil, nil
}

func runBooleanParameters(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if !isExported(fn.Name.Name) {
			return
		}
		if fn.Type.Params == nil {
			return
		}
		for _, param := range fn.Type.Params.List {
			if isBoolType(param.Type) {
				pass.Reportf(param.Pos(), "boolean parameters indicate a function does two things; split the function or use a typed option")
			}
		}
	})
	return nil, nil
}

func runNoConstructorWithUnexportedFields(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}

func runGetenvOutsideMain(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() == "main" {
		return nil, nil
	}
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call := node.(*ast.CallExpr)
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && (sel.Sel.Name == "Getenv" || sel.Sel.Name == "LookupEnv") {
				pass.Reportf(call.Pos(), "os.Getenv only allowed in main or config package")
			}
		}
	})
	return nil, nil
}

func runGlobalSlogFunctions(pass *analysis.Pass) (interface{}, error) {
	slogFuncs := map[string]bool{
		"Info":         true,
		"Error":        true,
		"Warn":         true,
		"Debug":        true,
		"Log":          true,
		"InfoContext":  true,
		"ErrorContext": true,
		"WarnContext":  true,
		"DebugContext": true,
	}
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call := node.(*ast.CallExpr)
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "slog" && slogFuncs[sel.Sel.Name] {
				pass.Reportf(call.Pos(), "inject *slog.Logger via constructor or parameter; do not use global slog functions")
			}
		}
	})
	return nil, nil
}

func runAnyInExportedAPI(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if !isExported(fn.Name.Name) {
			return
		}
		if fn.Type.Params != nil {
			for _, param := range fn.Type.Params.List {
				if isAnyType(param.Type) {
					pass.Reportf(param.Pos(), "avoid any/interface{} in exported APIs; use a concrete type or interface")
				}
			}
		}
	})
	return nil, nil
}

func runFunctionTooLong(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil)}, func(node ast.Node) {
		var body *ast.BlockStmt
		var startPos, endPos int
		switch n := node.(type) {
		case *ast.FuncDecl:
			body = n.Body
			startPos = int(n.Pos())
			endPos = int(n.End())
		case *ast.FuncLit:
			body = n.Body
			startPos = int(n.Pos())
			endPos = int(n.End())
		}
		if body == nil {
			return
		}
		lineCount := countLines(startPos, endPos)
		if lineCount > 60 {
			pass.Reportf(ast.Node(node).(ast.Node).Pos(), "function too long (%d lines); maximum 60 lines", lineCount)
		}
	})
	return nil, nil
}

func isExported(name string) bool {
	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}

func isBoolType(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "bool"
	}
	return false
}

func isAnyType(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok {
		return ident.Name == "any" || ident.Name == "interface{}"
	}
	return false
}

func countLines(start, end int) int {
	return end - start
}
