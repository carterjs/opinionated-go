package structs

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	// ExportedFieldsWithMethods errors on exported fields in structs with methods.
	ExportedFieldsWithMethods = &analysis.Analyzer{
		Name:     "exported_fields_with_methods",
		Doc:      "error on exported fields in structs with methods",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runExportedFieldsWithMethods,
	}
	// BooleanParameters errors on [bool] parameters in exported functions.
	BooleanParameters = &analysis.Analyzer{
		Name:     "boolean_parameters",
		Doc:      "error on boolean parameters in exported functions",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runBooleanParameters,
	}
	// NoConstructorWithUnexportedFields warns when a struct has unexported fields but no constructor.
	NoConstructorWithUnexportedFields = &analysis.Analyzer{
		Name:     "no_constructor_with_unexported_fields",
		Doc:      "warn when struct has unexported fields but no constructor",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNoConstructorWithUnexportedFields,
	}
	// GetenvOutsideMain errors on [os.Getenv] calls outside main.
	GetenvOutsideMain = &analysis.Analyzer{
		Name:     "getenv_outside_main",
		Doc:      "error on os.Getenv outside main",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runGetenvOutsideMain,
	}
	// GlobalSlogFunctions errors on global [log/slog] function calls.
	GlobalSlogFunctions = &analysis.Analyzer{
		Name:     "global_slog_functions",
		Doc:      "error on global slog function calls",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runGlobalSlogFunctions,
	}
	// AnyInExportedAPI warns on any/interface{} in exported APIs.
	AnyInExportedAPI = &analysis.Analyzer{
		Name:     "any_in_exported_api",
		Doc:      "warn on any/interface{} in exported APIs",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runAnyInExportedAPI,
	}
	// FunctionTooLong warns on functions longer than 60 lines.
	FunctionTooLong = &analysis.Analyzer{
		Name:     "function_too_long",
		Doc:      "warn on functions longer than 60 lines",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runFunctionTooLong,
	}
	// TooManyParameters warns on functions with more than 4 parameters.
	TooManyParameters = &analysis.Analyzer{
		Name:     "too_many_parameters",
		Doc:      "warn on functions with more than 4 parameters",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runTooManyParameters,
	}
	// InterfaceTooLarge warns on interfaces with more than 5 methods.
	InterfaceTooLarge = &analysis.Analyzer{
		Name:     "interface_too_large",
		Doc:      "warn on interfaces with more than 5 methods",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runInterfaceTooLarge,
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
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// First pass: collect all struct types with unexported fields
	structsWithUnexported := make(map[string]*ast.TypeSpec)
	inspect.Preorder([]ast.Node{(*ast.TypeSpec)(nil)}, func(node ast.Node) {
		spec := node.(*ast.TypeSpec)
		if !isExported(spec.Name.Name) {
			return
		}

		st, ok := spec.Type.(*ast.StructType)
		if !ok || st.Fields == nil {
			return
		}

		hasUnexported := false
		for _, field := range st.Fields.List {
			if len(field.Names) > 0 && !isExported(field.Names[0].Name) {
				hasUnexported = true
				break
			}
		}

		if hasUnexported {
			structsWithUnexported[spec.Name.Name] = spec
		}
	})

	// Second pass: find constructors for these structs
	constructors := make(map[string]bool)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Recv != nil {
			return // Skip methods
		}
		if strings.HasPrefix(fn.Name.Name, "New") && len(fn.Name.Name) > 3 {
			constructors[fn.Name.Name[3:]] = true // Remove "New" prefix
		}
	})

	// Check each struct with unexported fields for a constructor
	for typeName, spec := range structsWithUnexported {
		if !constructors[typeName] {
			pass.Reportf(spec.Name.Pos(), "struct %q has unexported fields but no constructor; add New%s()", typeName, typeName)
		}
	}

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
		var startToken, endToken ast.Node
		switch n := node.(type) {
		case *ast.FuncDecl:
			body = n.Body
			startToken = n
			endToken = n
		case *ast.FuncLit:
			body = n.Body
			startToken = n
			endToken = n
		}
		if body == nil {
			return
		}
		startLine := pass.Fset.Position(startToken.Pos()).Line
		endLine := pass.Fset.Position(endToken.End()).Line
		lineCount := endLine - startLine + 1
		if lineCount > 60 {
			pass.Reportf(startToken.Pos(), "function too long (%d lines); maximum 60 lines", lineCount)
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
		return ident.Name == "any"
	}
	// Check for interface{}
	if iface, ok := expr.(*ast.InterfaceType); ok {
		if iface.Methods == nil || len(iface.Methods.List) == 0 {
			return true
		}
	}
	return false
}

func runTooManyParameters(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Type.Params == nil {
			return
		}

		// Count parameters, excluding receiver for methods
		params := fn.Type.Params.List
		if len(params) > 4 {
			pass.Reportf(fn.Pos(), "function has %d parameters; use config struct for more than 4", len(params))
		}
	})
	return nil, nil
}

func runInterfaceTooLarge(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.InterfaceType)(nil)}, func(node ast.Node) {
		iface := node.(*ast.InterfaceType)
		if iface.Methods == nil {
			return
		}

		methodCount := len(iface.Methods.List)
		if methodCount > 5 {
			pass.Reportf(iface.Pos(), "interface has %d methods; keep interfaces small (5 or fewer)", methodCount)
		}
	})
	return nil, nil
}

