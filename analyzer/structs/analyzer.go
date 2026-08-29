package structs

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// strconvBaseFuncs are strconv conversions whose non-string integer
// arguments are a base or a bit size, not a value the author chose: the 10
// in strconv.FormatInt(n, 10) means decimal, not a magic number to name.
var strconvBaseFuncs = map[string]bool{
	"FormatInt":    true,
	"FormatUint":   true,
	"AppendInt":    true,
	"AppendUint":   true,
	"ParseInt":     true,
	"ParseUint":    true,
	"ParseFloat":   true,
	"ParseComplex": true,
	"AppendFloat":  true,
}

// knownNumericIdiomLiterals collects the positions of integer literals that
// are arguments to a strconv base/bit-size conversion, so runMagicNumbers can
// leave them alone.
func knownNumericIdiomLiterals(pass *analysis.Pass, decl ast.Decl) map[token.Pos]bool {
	exempt := make(map[token.Pos]bool)
	ast.Inspect(decl, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || !strconvBaseFuncs[sel.Sel.Name] {
			return true
		}
		obj := pass.TypesInfo.Uses[sel.Sel]
		if obj == nil || obj.Pkg() == nil || obj.Pkg().Path() != "strconv" {
			return true
		}
		for _, arg := range call.Args {
			if lit, ok := arg.(*ast.BasicLit); ok && lit.Kind == token.INT {
				exempt[lit.Pos()] = true
			}
		}
		return true
	})
	return exempt
}

const (
	// maxParameters is the point past which a signature owes its callers a struct.
	maxParameters = 4
	// maxInterfaceMethods is the point past which an interface has stopped being an abstraction.
	maxInterfaceMethods = 5
	// maxMainStatements is the point past which main has stopped only wiring.
	maxMainStatements = 10
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
	// OsExitOutsideMain errors on [os.Exit] outside main.
	OsExitOutsideMain = &analysis.Analyzer{
		Name:     "os_exit_outside_main",
		Doc:      "error on os.Exit outside func main; every other function returns an error and lets its caller decide",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runOsExitOutsideMain,
	}
	// MainDoesMoreThanWire warns when main holds logic that belongs in Run.
	MainDoesMoreThanWire = &analysis.Analyzer{
		Name:     "main_does_more_than_wire",
		Doc:      "warn when func main does more than wire and exit; logic belongs in a Run function a test can call",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runMainDoesMoreThanWire,
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

// runOsExitOutsideMain reports os.Exit outside main. Exiting is the one
// decision a library never gets to make for its caller; everywhere else the
// answer is an error returned up to the function that owns the process.
func runOsExitOutsideMain(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Body == nil {
			return
		}
		// TestMain owns the test binary's exit code the same way main owns the program's.
		if fn.Recv == nil && (isProgramMain(pass, fn) || fn.Name.Name == "TestMain") {
			return
		}

		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			if selector, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := selector.X.(*ast.Ident); ok && ident.Name == "os" && selector.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "os.Exit only allowed in main; return an error and let the caller decide")
				}
			}
			return true
		})
	})
	return nil, nil
}

// runMainDoesMoreThanWire reports a main that has grown past wiring. Logic
// inside main is logic no test can reach without starting a process, so it
// belongs in a Run function that takes its dependencies and returns an error.
func runMainDoesMoreThanWire(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Body == nil || !isProgramMain(pass, fn) {
			return
		}
		if len(fn.Body.List) > maxMainStatements {
			pass.Reportf(fn.Pos(), "main has %d statements; main wires and exits, so move the logic into Run(ctx context.Context, ...) error", len(fn.Body.List))
		}
	})
	return nil, nil
}

// isProgramMain reports whether fn is the entry point of a program.
func isProgramMain(pass *analysis.Pass, fn *ast.FuncDecl) bool {
	return fn.Recv == nil && fn.Name.Name == "main" && pass.Pkg.Name() == "main"
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
		if fn.Type.Params == nil || !isExported(fn.Name.Name) {
			return
		}

		// `a, b, c string` is one field naming three parameters, and counting fields lets it past the limit.
		count := 0
		for _, field := range fn.Type.Params.List {
			if len(field.Names) == 0 {
				count++
				continue
			}
			count += len(field.Names)
		}
		if count > maxParameters {
			pass.Reportf(fn.Pos(), "function has %d parameters; use config struct for more than %d", count, maxParameters)
		}
	})
	return nil, nil
}

func runInterfaceTooLarge(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	named := interfaceDeclarations(inspect)
	inspect.Preorder([]ast.Node{(*ast.InterfaceType)(nil)}, func(node ast.Node) {
		iface := node.(*ast.InterfaceType)
		if iface.Methods == nil {
			return
		}

		// An unexported interface answers to its own package; an anonymous one has no name to judge.
		spec, ok := named[iface]
		if !ok || !isExported(spec.Name.Name) {
			return
		}

		methodCount := len(iface.Methods.List)
		if methodCount > maxInterfaceMethods {
			pass.Reportf(iface.Pos(), "interface has %d methods; keep interfaces small (%d or fewer)", methodCount, maxInterfaceMethods)
		}
	})
	return nil, nil
}

// interfaceDeclarations maps each named interface type to the spec that names it.
func interfaceDeclarations(inspect *inspector.Inspector) map[*ast.InterfaceType]*ast.TypeSpec {
	declarations := make(map[*ast.InterfaceType]*ast.TypeSpec)
	inspect.Preorder([]ast.Node{(*ast.TypeSpec)(nil)}, func(node ast.Node) {
		spec := node.(*ast.TypeSpec)
		if iface, ok := spec.Type.(*ast.InterfaceType); ok {
			declarations[iface] = spec
		}
	})
	return declarations
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
			exempt := knownNumericIdiomLiterals(pass, decl)
			ast.Inspect(decl, func(node ast.Node) bool {
				lit, ok := node.(*ast.BasicLit)
				if !ok {
					return true
				}
				if lit.Kind != token.INT && lit.Kind != token.FLOAT {
					return true
				}
				if allowedLiterals[lit.Value] || reported[lit.Value] || exempt[lit.Pos()] {
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
