package structs

import (
	"go/ast"
	"go/token"
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
	// MagicNumbers errors on numeric literals other than 0 and 1 outside a constant declaration.
	MagicNumbers = &analysis.Analyzer{
		Name:     "magic_numbers",
		Doc:      "error on numeric literals other than 0 and 1 outside constant declarations",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runMagicNumbers,
	}
)

// runExportedFieldsWithMethods reports a struct that exposes fields directly.
// The finding is about the type, not about each field that happens to be
// capitalised, so it is reported once against the declaration and names the
// fields it means.
func runExportedFieldsWithMethods(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	declarations := structDeclarations(inspect)

	withMethods := typesWithMethods(inspect)

	inspect.Preorder([]ast.Node{(*ast.StructType)(nil)}, func(node ast.Node) {
		st := node.(*ast.StructType)
		if st.Fields == nil || len(st.Fields.List) == 0 {
			return
		}

		// The rule is about types that have behaviour. A plain record — a
		// request payload, a config, a row — is exported fields and nothing
		// else, and has no methods to route access through.
		spec, ok := declarations[st]
		if !ok || !withMethods[spec.Name.Name] {
			return
		}

		var exported []string
		for _, field := range st.Fields.List {
			for _, name := range field.Names {
				if isExported(name.Name) {
					exported = append(exported, name.Name)
				}
			}
		}
		if len(exported) == 0 {
			return
		}

		pass.Reportf(spec.Name.Pos(), "struct %q has methods and should not have exported fields (%s); control access through methods", spec.Name.Name, strings.Join(exported, ", "))
	})
	return nil, nil
}

// typesWithMethods names the types this package declares methods on.
func typesWithMethods(inspect *inspector.Inspector) map[string]bool {
	withMethods := make(map[string]bool)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Recv == nil || len(fn.Recv.List) == 0 {
			return
		}
		if name := receiverTypeName(fn.Recv.List[0].Type); name != "" {
			withMethods[name] = true
		}
	})
	return withMethods
}

// receiverTypeName unwraps a receiver's pointer and type parameters to the
// bare type name the method is declared on.
func receiverTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(t.X)
	case *ast.IndexExpr:
		return receiverTypeName(t.X)
	case *ast.IndexListExpr:
		return receiverTypeName(t.X)
	case *ast.Ident:
		return t.Name
	}
	return ""
}

// structDeclarations maps each named struct type back to the declaration that
// names it, so a finding can be reported against the type rather than against
// whichever field happened to be noticed first.
func structDeclarations(inspect *inspector.Inspector) map[*ast.StructType]*ast.TypeSpec {
	declarations := make(map[*ast.StructType]*ast.TypeSpec)
	inspect.Preorder([]ast.Node{(*ast.TypeSpec)(nil)}, func(node ast.Node) {
		spec := node.(*ast.TypeSpec)
		if st, ok := spec.Type.(*ast.StructType); ok {
			declarations[st] = spec
		}
	})
	return declarations
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

// allowedLiterals are the numeric literals that carry their meaning on their
// face: an empty count, a single step, an identity.
var allowedLiterals = map[string]bool{
	"0":   true,
	"1":   true,
	"0.0": true,
	"1.0": true,
}

// runMagicNumbers reports numeric literals that should have been named. A
// number written where it is used tells the reader what the machine does but
// not what the author meant; a constant carries the meaning to every use.
// Constant declarations are where a literal belongs, and test tables are where
// concrete values are the point, so both are exempt.
func runMagicNumbers(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go") {
			continue
		}
		for _, decl := range file.Decls {
			if isConstDecl(decl) {
				continue
			}
			// The same number repeated within one declaration is one constant
			// waiting to be named, so it is reported once at its first use.
			reported := make(map[string]bool)
			ast.Inspect(decl, func(node ast.Node) bool {
				lit, ok := node.(*ast.BasicLit)
				if !ok {
					return true
				}
				if lit.Kind != token.INT && lit.Kind != token.FLOAT {
					return true
				}
				if allowedLiterals[lit.Value] || reported[lit.Value] {
					return true
				}
				reported[lit.Value] = true
				pass.Reportf(lit.Pos(), "magic number %s; give it a named constant", lit.Value)
				return true
			})
		}
	}
	return nil, nil
}

func isConstDecl(decl ast.Decl) bool {
	gen, ok := decl.(*ast.GenDecl)
	return ok && gen.Tok == token.CONST
}
