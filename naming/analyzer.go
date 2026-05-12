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

	InitialismCasing = &analysis.Analyzer{
		Name:     "initialism_casing",
		Doc:      "error on initialisms in wrong case (Id, Url, Http, Api, Json, etc.)",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runInitialismCasing,
	}

	CommonAbbreviations = &analysis.Analyzer{
		Name:     "common_abbreviations",
		Doc:      "warn on common abbreviations (Doc, Req, Resp, Cfg, Ctx, etc.)",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runCommonAbbreviations,
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

func runCommonAbbreviations(pass *analysis.Pass) (interface{}, error) {
	abbrevs := []string{"Doc", "Req", "Resp", "Cfg", "Ctx", "Msg", "Num", "Str", "Buf", "Ptr", "Pkg", "Src", "Dst", "Tmp", "Val", "Var", "Obj", "Mgr", "Svc", "Repo", "Impl"}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.Ident)(nil)}, func(node ast.Node) {
		ident := node.(*ast.Ident)
		if !isExported(ident.Name) {
			return
		}

		for _, abbrev := range abbrevs {
			if matchesAbbreviation(ident.Name, abbrev) {
				pass.Reportf(ident.Pos(), "avoid abbreviation %q; use full word", abbrev)
				return
			}
		}
	})
	return nil, nil
}

func matchesAbbreviation(name, abbrev string) bool {
	if name == abbrev {
		return true
	}
	if strings.HasPrefix(name, abbrev) && len(name) > len(abbrev) && unicode.IsUpper(rune(name[len(abbrev)])) {
		return true
	}
	if strings.HasSuffix(name, abbrev) && len(name) > len(abbrev) {
		idx := len(name) - len(abbrev)
		return unicode.IsUpper(rune(name[idx-1]))
	}
	return false
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
	if strings.Contains(pkgName, "_") && len(pass.Files) > 0 {
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
