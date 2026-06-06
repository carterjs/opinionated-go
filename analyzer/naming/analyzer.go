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
	ReceiverNames = &analysis.Analyzer{
		Name:     "receiver_names",
		Doc:      "warn on receiver names that are 1-2 characters when type name is longer",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runReceiverNames,
	}

	ConsistentReceivers = &analysis.Analyzer{
		Name:     "consistent_receivers",
		Doc:      "error on inconsistent receiver names across methods on the same type",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runConsistentReceivers,
	}

	InitialismCasing = &analysis.Analyzer{
		Name:     "initialism_casing",
		Doc:      "error on initialisms in wrong case (Id, Url, Http, Api, Json, etc.)",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runInitialismCasing,
	}

	SingleLetterExported = &analysis.Analyzer{
		Name:     "single_letter_exported",
		Doc:      "error on single-letter exported names at package scope",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runSingleLetterExported,
	}

	SingleLetterVariables = &analysis.Analyzer{
		Name:     "single_letter_variables",
		Doc:      "warn on single-letter local variables except in for loops or narrow scopes",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runSingleLetterVariables,
	}

	ContextAndErrorNaming = &analysis.Analyzer{
		Name:     "context_error_naming",
		Doc:      "error when context.Context not named ctx or error not named err",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runContextAndErrorNaming,
	}

	PackageNaming = &analysis.Analyzer{
		Name:     "package_naming",
		Doc:      "error on package names with underscores or mismatched directory names",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runPackageNaming,
	}

	FileNaming = &analysis.Analyzer{
		Name:     "file_naming",
		Doc:      "warn on file names with underscores (except _test.go and platform variants)",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runFileNaming,
	}

	GenericPackageNames = &analysis.Analyzer{
		Name:     "generic_package_names",
		Doc:      "error on generic package names (util, common, helpers, shared, etc.)",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runGenericPackageNames,
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

	// Check if this is a test file
	isTestFile := false
	if len(pass.Files) > 0 {
		filename := pass.Fset.File(pass.Files[0].Pos()).Name()
		isTestFile = strings.HasSuffix(filename, "_test.go")
	}

	// Track for loop variables and switch cases (allowed to be single letter)
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

	// Check function/method parameters and local variables
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.FuncLit)(nil)}, func(node ast.Node) {
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

		// Check parameters - allow idiomatic single-letter params
		if params != nil {
			for _, param := range params.List {
				paramType := typeString(param.Type)
				for _, name := range param.Names {
					if len(name.Name) == 1 {
						// Allow 't' for testing.T in test files
						if isTestFile && name.Name == "t" && (paramType == "testing.T" || paramType == "*testing.T") {
							continue
						}
						// Allow 'n' for ast.Node
						if name.Name == "n" && paramType == "ast.Node" {
							continue
						}
						pass.Reportf(name.Pos(), "parameter %q is too short; use a descriptive name", name.Name)
					}
				}
			}
		}

		// Check local variable declarations
		if body != nil {
			checkLocalVariables(pass, body, allowedContexts)
		}
	})

	return nil, nil
}

func checkLocalVariables(pass *analysis.Pass, body *ast.BlockStmt, allowedContexts map[string]bool) {
	ast.Inspect(body, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			if stmt.Tok == token.ASSIGN || stmt.Tok == token.DEFINE {
				for i, lhs := range stmt.Lhs {
					if ident, ok := lhs.(*ast.Ident); ok && len(ident.Name) == 1 && !allowedContexts[ident.Name] {
						// Don't flag blank identifier
						if ident.Name == "_" {
							continue
						}
						// Don't flag type assertions (e.g., x := value.(type))
						if i < len(stmt.Rhs) {
							if _, ok := stmt.Rhs[i].(*ast.TypeAssertExpr); ok {
								continue
							}
						}
						pass.Reportf(ident.Pos(), "variable %q is too short; use a descriptive name", ident.Name)
					}
				}
			}
		}
		return true
	})
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
					if paramType == "context.Context" && name.Name != "ctx" {
						pass.Reportf(name.Pos(), "context.Context parameter should be named ctx, not %q", name.Name)
					}
					if paramType == "error" && name.Name != "err" {
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
	filename := pass.Fset.File(pass.Files[0].Pos()).Name()
	base := strings.Split(filename, "/")
	if len(base) > 0 {
		filename = base[len(base)-1]
	}

	if strings.Contains(filename, "_") && !isValidTestOrPlatformFile(filename) && len(pass.Files) > 0 {
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
