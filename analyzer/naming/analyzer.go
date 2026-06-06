package naming

import (
	"go/ast"
	"go/token"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	// ReceiverNames warns on receiver names that are too short for longer type names.
	ReceiverNames = &analysis.Analyzer{
		Name:     "receiver_names",
		Doc:      "warn on receiver names that are 1-2 characters when type name is longer",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runReceiverNames,
	}

	// ConsistentReceivers errors on inconsistent receiver names across methods on the same type.
	ConsistentReceivers = &analysis.Analyzer{
		Name:     "consistent_receivers",
		Doc:      "error on inconsistent receiver names across methods on the same type",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runConsistentReceivers,
	}

	// ShadowBuiltins errors on declaring functions that shadow Go built-ins.
	ShadowBuiltins = &analysis.Analyzer{
		Name:     "shadow_builtins",
		Doc:      "error on declaring functions that shadow Go built-ins",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runShadowBuiltins,
	}

	// InitialismCasing errors on initialisms with incorrect casing.
	InitialismCasing = &analysis.Analyzer{
		Name:     "initialism_casing",
		Doc:      "error on initialisms in wrong case (Id, Url, Http, Api, Json, etc.)",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runInitialismCasing,
	}

	// SingleLetterExported errors on single-letter exported names at package scope.
	SingleLetterExported = &analysis.Analyzer{
		Name:     "single_letter_exported",
		Doc:      "error on single-letter exported names at package scope",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runSingleLetterExported,
	}

	// SingleLetterVariables warns on single-letter local variables except in idiomatic contexts.
	SingleLetterVariables = &analysis.Analyzer{
		Name:     "single_letter_variables",
		Doc:      "warn on single-letter local variables except in for loops or narrow scopes",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runSingleLetterVariables,
	}

	// ContextAndErrorNaming errors when [context.Context] is not named ctx or [error] is not named err.
	ContextAndErrorNaming = &analysis.Analyzer{
		Name:     "context_error_naming",
		Doc:      "error when context.Context not named ctx or error not named err",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runContextAndErrorNaming,
	}

	// PackageNaming errors on package names with underscores or mismatched directory names.
	PackageNaming = &analysis.Analyzer{
		Name:     "package_naming",
		Doc:      "error on package names with underscores or mismatched directory names",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runPackageNaming,
	}

	// FileNaming warns on file names with underscores (except _test.go and platform variants).
	FileNaming = &analysis.Analyzer{
		Name:     "file_naming",
		Doc:      "warn on file names with underscores (except _test.go and platform variants)",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runFileNaming,
	}

	// GenericPackageNames errors on overly generic package names.
	GenericPackageNames = &analysis.Analyzer{
		Name:     "generic_package_names",
		Doc:      "error on generic package names (util, common, helpers, shared, etc.)",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runGenericPackageNames,
	}

	// JavaStyleGetters errors on Get<Field>() methods, preferring <Field>() instead.
	JavaStyleGetters = &analysis.Analyzer{
		Name:     "java_style_getters",
		Doc:      "error on Get<Field>() methods (use <Field>() instead)",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runJavaStyleGetters,
	}

	// VerbosePrefixMethods warns on To<Type>() conversion methods.
	VerbosePrefixMethods = &analysis.Analyzer{
		Name:     "verbose_prefix_methods",
		Doc:      "warn on To<Type>() conversion methods without parameters",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runVerbosePrefixMethods,
	}

	// StutteringNames warns on exported names that repeat their package name.
	StutteringNames = &analysis.Analyzer{
		Name:     "stuttering_names",
		Doc:      "warn on exported names that repeat their package name",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runStutteringNames,
	}
)

func runReceiverNames(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Recv == nil || len(fn.Recv.List) == 0 {
			return
		}
		recv := fn.Recv.List[0]
		recvName := recv.Names[0].Name
		recvType := getTypeName(recv.Type)

		if len(recvName) <= 2 && len(recvType) > 2 {
			pass.Reportf(recv.Names[0].Pos(), "receiver name %q is too short; use a descriptive word", recvName)
		}
	})
	return nil, nil
}

func runConsistentReceivers(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Map from receiver type name to (first receiver name, first method name)
	firstReceiverName := make(map[string]string)
	firstMethod := make(map[string]string)

	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Recv == nil || len(fn.Recv.List) == 0 {
			return
		}

		recv := fn.Recv.List[0]
		recvName := recv.Names[0].Name
		recvType := getTypeName(recv.Type)

		if recvType == "" {
			return
		}

		// First time seeing this type
		if firstReceiverName[recvType] == "" {
			firstReceiverName[recvType] = recvName
			firstMethod[recvType] = fn.Name.Name
			return
		}

		// Check if receiver name matches the first one seen
		if recvName != firstReceiverName[recvType] {
			pass.Reportf(recv.Names[0].Pos(),
				"receiver name %q inconsistent with %s (which uses %q)",
				recvName, firstMethod[recvType], firstReceiverName[recvType])
		}
	})

	return nil, nil
}

func runInitialismCasing(pass *analysis.Pass) (interface{}, error) {
	wrongInitialisms := map[string]string{
		"Id":   "ID",
		"Url":  "URL",
		"Http": "HTTP",
		"Api":  "API",
		"Json": "JSON",
		"Xml":  "XML",
		"Sql":  "SQL",
		"Css":  "CSS",
		"Html": "HTML",
		"Rpc":  "RPC",
		"Uid":  "UID",
		"Uuid": "UUID",
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.GenDecl)(nil), (*ast.FuncDecl)(nil), (*ast.TypeSpec)(nil)}, func(node ast.Node) {
		switch n := node.(type) {
		case *ast.GenDecl:
			if n.Tok == token.TYPE || n.Tok == token.CONST || n.Tok == token.VAR {
				for _, spec := range n.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok && isExported(ts.Name.Name) {
						for wrong, correct := range wrongInitialisms {
							if strings.Contains(ts.Name.Name, wrong) {
								pass.Reportf(ts.Pos(), "initialism %q should be %q", wrong, correct)
							}
						}
					} else if vs, ok := spec.(*ast.ValueSpec); ok {
						for _, name := range vs.Names {
							if isExported(name.Name) {
								for wrong, correct := range wrongInitialisms {
									if strings.Contains(name.Name, wrong) {
										pass.Reportf(name.Pos(), "initialism %q should be %q", wrong, correct)
									}
								}
							}
						}
					}
				}
			}
		case *ast.FuncDecl:
			if isExported(n.Name.Name) {
				for wrong, correct := range wrongInitialisms {
					if strings.Contains(n.Name.Name, wrong) {
						pass.Reportf(n.Pos(), "initialism %q should be %q", wrong, correct)
					}
				}
			}
		}
	})
	return nil, nil
}

func runSingleLetterExported(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Check type declarations
	inspect.Preorder([]ast.Node{(*ast.TypeSpec)(nil)}, func(node ast.Node) {
		ts := node.(*ast.TypeSpec)
		if len(ts.Name.Name) == 1 && isExported(ts.Name.Name) {
			pass.Reportf(ts.Name.Pos(), "exported name %q is too short; use a descriptive word", ts.Name.Name)
		}
	})

	// Check function declarations
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if len(fn.Name.Name) == 1 && isExported(fn.Name.Name) {
			pass.Reportf(fn.Name.Pos(), "exported name %q is too short; use a descriptive word", fn.Name.Name)
		}
	})

	// Check var/const declarations at package level
	inspect.Preorder([]ast.Node{(*ast.GenDecl)(nil)}, func(node ast.Node) {
		gd := node.(*ast.GenDecl)
		if gd.Tok == token.VAR || gd.Tok == token.CONST {
			for _, spec := range gd.Specs {
				if vs, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range vs.Names {
						if len(name.Name) == 1 && isExported(name.Name) {
							pass.Reportf(name.Pos(), "exported name %q is too short; use a descriptive word", name.Name)
						}
					}
				}
			}
		}
	})

	return nil, nil
}

func runSingleLetterVariables(pass *analysis.Pass) (interface{}, error) {
	// Skip build cache and generated files
	if shouldSkipPass(pass) {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	allowedContexts := trackForLoopVars(inspect)

	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil)}, func(node ast.Node) {
		// Determine if this specific function is in a test file
		isTestFile := false
		filename := pass.Fset.File(node.Pos()).Name()
		if strings.HasSuffix(filename, "_test.go") {
			isTestFile = true
		}
		checkFunctionVariables(pass, node, isTestFile, allowedContexts)
	})

	return nil, nil
}

func isTestFilePath(pass *analysis.Pass) bool {
	if len(pass.Files) == 0 {
		return false
	}
	filename := pass.Fset.File(pass.Files[0].Pos()).Name()
	return strings.HasSuffix(filename, "_test.go")
}

func trackForLoopVars(inspect *inspector.Inspector) map[string]bool {
	allowedContexts := make(map[string]bool)
	inspect.Preorder([]ast.Node{(*ast.ForStmt)(nil)}, func(node ast.Node) {
		fs := node.(*ast.ForStmt)
		if fs.Init != nil {
			if as, ok := fs.Init.(*ast.AssignStmt); ok {
				for _, lhs := range as.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok && len(ident.Name) == 1 {
						allowedContexts[ident.Name] = true
					}
				}
			}
		}
	})
	return allowedContexts
}

func checkFunctionVariables(pass *analysis.Pass, node ast.Node, isTestFile bool, allowedContexts map[string]bool) {
	var params *ast.FieldList
	var body *ast.BlockStmt

	switch n := node.(type) {
	case *ast.FuncDecl:
		params = n.Type.Params
		body = n.Body
	case *ast.FuncLit:
		params = n.Type.Params
		body = n.Body
	}

	checkFunctionParams(pass, params, isTestFile, body)

	if body != nil {
		checkLocalVariables(pass, body, allowedContexts)
	}
}

func checkFunctionParams(pass *analysis.Pass, params *ast.FieldList, isTestFile bool, body *ast.BlockStmt) {
	if params == nil || body == nil {
		return
	}

	// Build map of parameter usage spans
	paramUsage := make(map[string]token.Pos)
	ast.Inspect(body, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			paramUsage[ident.Name] = ident.Pos()
		}
		return true
	})

	for _, param := range params.List {
		paramType := typeString(param.Type)
		for _, name := range param.Names {
			if len(name.Name) == 1 && name.Name != "_" && !isIdiomatic(name.Name, paramType, isTestFile) {
				// Check if parameter is used within 5 lines
				if lastUsePos, used := paramUsage[name.Name]; used {
					declLine := pass.Fset.Position(name.Pos()).Line
					lastUseLine := pass.Fset.Position(lastUsePos).Line
					span := lastUseLine - declLine
					if span <= 5 {
						continue // Allow if used within 5 lines
					}
				}
				pass.Reportf(name.Pos(), "parameter %q is too short; use a descriptive name", name.Name)
			}
		}
	}
}

func isIdiomatic(name, paramType string, isTestFile bool) bool {
	if isTestFile {
		// Standard testing parameters
		if name == "t" && (paramType == "testing.T" || paramType == "*testing.T" || paramType == "testing.TB" || paramType == "*testing.TB") {
			return true
		}
		if name == "b" && (paramType == "testing.B" || paramType == "*testing.B") {
			return true
		}
		if name == "m" && (paramType == "testing.M" || paramType == "*testing.M") {
			return true
		}
		if name == "f" && (paramType == "testing.F" || paramType == "*testing.F") {
			return true
		}
	}
	if name == "n" && paramType == "ast.Node" {
		return true
	}
	if name == "w" && (paramType == "http.ResponseWriter" || paramType == "ResponseWriter") {
		return true
	}
	if name == "r" && (paramType == "*http.Request" || paramType == "*Request") {
		return true
	}
	// Common single-letter parameter names in loops and comparison functions
	commonParams := map[string]bool{
		"i": true, // loop index
		"j": true, // loop index
		"k": true, // loop index
		"a": true, // comparison function
		"b": true, // comparison function
		"c": true, // comparison function
	}
	if commonParams[name] {
		return true
	}
	return false
}

func checkLocalVariables(pass *analysis.Pass, body *ast.BlockStmt, allowedContexts map[string]bool) {
	// Track single-letter variable declarations and their usage spans
	varDecls := make(map[string]token.Pos)  // var name -> declaration position
	varLastUse := make(map[string]token.Pos) // var name -> last usage position

	// First pass: find all single-letter variable declarations
	ast.Inspect(body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			if stmt.Tok == token.ASSIGN || stmt.Tok == token.DEFINE {
				for i, lhs := range stmt.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok && len(ident.Name) == 1 && !allowedContexts[ident.Name] {
						if ident.Name != "_" {
							// Don't record type assertions
							if i < len(stmt.Rhs) {
								if _, ok := stmt.Rhs[i].(*ast.TypeAssertExpr); ok {
									continue
								}
							}
							varDecls[ident.Name] = ident.Pos()
						}
					}
				}
			}
		}
		return true
	})

	// Second pass: find last usage of each variable
	ast.Inspect(body, func(n ast.Node) bool {
		if ident, ok := n.(*ast.Ident); ok {
			if _, isDeclared := varDecls[ident.Name]; isDeclared {
				varLastUse[ident.Name] = ident.Pos()
			}
		}
		return true
	})

	// Check if any variable spans more than 5 lines
	for varName, declPos := range varDecls {
		lastUsePos, hasUsage := varLastUse[varName]
		if !hasUsage {
			continue // Variable not used after declaration
		}

		declLine := pass.Fset.Position(declPos).Line
		lastUseLine := pass.Fset.Position(lastUsePos).Line
		span := lastUseLine - declLine

		if span > 5 {
			pass.Reportf(declPos, "variable %q is too short and spans %d lines; use a descriptive name", varName, span)
		}
	}
}

func runContextAndErrorNaming(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil)}, func(node ast.Node) {
		var params *ast.FieldList
		switch n := node.(type) {
		case *ast.FuncDecl:
			params = n.Type.Params
		case *ast.FuncLit:
			params = n.Type.Params
		}

		if params != nil {
			for _, param := range params.List {
				paramType := typeString(param.Type)
				for _, name := range param.Names {
					if paramType == "context.Context" && name.Name != "ctx" && name.Name != "_" {
						pass.Reportf(name.Pos(), "context.Context parameter should be named ctx, not %q", name.Name)
					}
					if paramType == "error" && name.Name != "err" && name.Name != "_" {
						pass.Reportf(name.Pos(), "error parameter should be named err, not %q", name.Name)
					}
				}
			}
		}
	})

	inspect.Preorder([]ast.Node{(*ast.ValueSpec)(nil)}, func(node ast.Node) {
		spec := node.(*ast.ValueSpec)
		if spec.Type != nil && typeString(spec.Type) == "error" {
			for _, name := range spec.Names {
				if name.Name != "err" && name.Name != "_" {
					pass.Reportf(name.Pos(), "error variable should be named err, not %q", name.Name)
				}
			}
		}
	})

	return nil, nil
}

func runPackageNaming(pass *analysis.Pass) (interface{}, error) {
	pkgName := pass.Pkg.Name()
	// Allow _test suffix for test packages
	if strings.Contains(pkgName, "_") && !strings.HasSuffix(pkgName, "_test") && len(pass.Files) > 0 {
		pass.Reportf(pass.Files[0].Pos(), "package name %q contains underscore; use a single word", pkgName)
	}
	return nil, nil
}

func runFileNaming(pass *analysis.Pass) (interface{}, error) {
	if len(pass.Files) == 0 {
		return nil, nil
	}
	filename := pass.Fset.File(pass.Files[0].Pos()).Name()
	base := strings.Split(filename, "/")
	if len(base) > 0 {
		filename = base[len(base)-1]
	}

	if strings.Contains(filename, "_") && !isValidTestOrPlatformFile(filename) {
		pass.Reportf(pass.Files[0].Pos(), "file name %q should not contain underscores", filename)
	}
	return nil, nil
}

func runGenericPackageNames(pass *analysis.Pass) (interface{}, error) {
	genericNames := map[string]bool{
		"util":    true,
		"utils":   true,
		"common":  true,
		"shared":  true,
		"helpers": true,
		"helper":  true,
		"misc":    true,
		"base":    true,
	}

	pkgName := pass.Pkg.Name()
	if genericNames[pkgName] && len(pass.Files) > 0 {
		pass.Reportf(pass.Files[0].Pos(), "package name %q is too generic; use a descriptive name", pkgName)
	}
	return nil, nil
}

// Helpers

func isExported(name string) bool {
	return len(name) > 0 && unicode.IsUpper(rune(name[0]))
}

func getTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return getTypeName(t.X)
	default:
		return ""
	}
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

func isValidTestOrPlatformFile(filename string) bool {
	if strings.HasSuffix(filename, "_test.go") {
		return true
	}

	platforms := []string{"_linux", "_darwin", "_windows", "_amd64", "_arm64"}
	for _, platform := range platforms {
		if strings.Contains(filename, platform) {
			return true
		}
	}
	return false
}

func shouldSkipPass(pass *analysis.Pass) bool {
	if len(pass.Files) == 0 {
		return false
	}
	filename := pass.Fset.File(pass.Files[0].Pos()).Name()
	// Skip build cache and generated files
	return strings.Contains(filename, "/go-build/")
}

func runJavaStyleGetters(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		// Only check exported methods (not bare functions)
		if !isExported(fn.Name.Name) {
			return
		}
		if fn.Recv == nil || len(fn.Recv.List) == 0 {
			return
		}

		// Check if method name starts with "Get" followed by uppercase
		name := fn.Name.Name
		if len(name) > 3 && name[:3] == "Get" && unicode.IsUpper(rune(name[3])) {
			// Only flag simple single-word getters (Get + one CamelCase word)
			remainder := name[3:]
			if isSimpleWord(remainder) && isSimpleGetter(fn) {
				suggestedName := strings.ToLower(name[3:4]) + name[4:]
				pass.Reportf(fn.Name.Pos(), "method %q should be named %q (Go idiom: method names should not use Get prefix)", name, suggestedName)
			}
		}
	})

	return nil, nil
}

// isSimpleWord checks if a string is a single CamelCase word (no lowercase before uppercase).
func isSimpleWord(word string) bool {
	if len(word) == 0 {
		return false
	}
	// Check if there's another uppercase letter after the first one
	for i := 1; i < len(word); i++ {
		if unicode.IsUpper(rune(word[i])) {
			return false // Multiple words
		}
	}
	return true
}

// isSimpleGetter checks if a method looks like a simple getter.
func isSimpleGetter(fn *ast.FuncDecl) bool {
	// Should take no parameters (only receiver)
	if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 {
		return false
	}
	// Should return exactly 1 value (no error return)
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}
	return true
}

func runVerbosePrefixMethods(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		// Only check exported methods (not bare functions)
		if !isExported(fn.Name.Name) {
			return
		}
		if fn.Recv == nil || len(fn.Recv.List) == 0 {
			return
		}

		// Check if method name starts with "To" followed by uppercase
		name := fn.Name.Name
		if len(name) > 2 && name[:2] == "To" && unicode.IsUpper(rune(name[2])) {
			// Only flag if it takes no parameters (except receiver) and returns one value
			if isConversionMethod(fn) {
				pass.Reportf(fn.Name.Pos(), "method %q uses verbose To prefix; consider a more idiomatic name", name)
			}
		}
	})

	return nil, nil
}

// isConversionMethod checks if a method looks like a simple conversion method.
func isConversionMethod(fn *ast.FuncDecl) bool {
	// Should take no parameters (only receiver)
	if fn.Type.Params != nil && len(fn.Type.Params.List) > 0 {
		return false
	}
	// Should return exactly 1 value (no error return)
	if fn.Type.Results == nil || len(fn.Type.Results.List) != 1 {
		return false
	}
	return true
}

func runShadowBuiltins(pass *analysis.Pass) (interface{}, error) {
	builtins := map[string]bool{
		"append":  true,
		"cap":     true,
		"close":   true,
		"complex": true,
		"copy":    true,
		"delete":  true,
		"imag":    true,
		"len":     true,
		"make":    true,
		"new":     true,
		"panic":   true,
		"print":   true,
		"println": true,
		"real":    true,
		"recover": true,
		"max":     true,
		"min":     true,
		"clear":   true,
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if builtins[fn.Name.Name] {
			pass.Reportf(fn.Name.Pos(), "function %q shadows Go built-in; use a different name", fn.Name.Name)
		}
	})

	return nil, nil
}

func runStutteringNames(pass *analysis.Pass) (interface{}, error) {
	pkgName := pass.Pkg.Name()

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.TypeSpec)(nil)}, func(node ast.Node) {
		var name string
		var pos token.Pos

		switch n := node.(type) {
		case *ast.FuncDecl:
			// Only check exported package-level functions
			if n.Recv != nil || !isExported(n.Name.Name) {
				return
			}
			name = n.Name.Name
			pos = n.Name.Pos()
		case *ast.TypeSpec:
			if !isExported(n.Name.Name) {
				return
			}
			name = n.Name.Name
			pos = n.Name.Pos()
		default:
			return
		}

		// Check if name starts with the package name (case-insensitive comparison)
		if strings.HasPrefix(strings.ToLower(name), strings.ToLower(pkgName)) {
			pass.Reportf(pos, "%q is a stuttering name; avoid repeating the package name %q", name, pkgName)
		}
	})

	return nil, nil
}
