// Package layout enforces where declarations sit within a file and within a
// package. A reader who opens a file should meet its subject first: the
// constants it is measured in, then the type it is about, then the constructor
// that makes one.
package layout

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var (
	// ConstantsAtTop errors on constant declarations that appear below a function.
	ConstantsAtTop = &analysis.Analyzer{
		Name:     "constants_at_top",
		Doc:      "error on const declarations placed below function declarations",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runConstantsAtTop,
	}

	// TypeBeforeConstructor errors when a constructor is not declared immediately after the type it builds, and when a package's namesake file does not open with a type.
	TypeBeforeConstructor = &analysis.Analyzer{
		Name:     "type_before_constructor",
		Doc:      "error when New<Type> is not the declaration immediately following <Type>, or when <package>.go opens with a function instead of a type",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runTypeBeforeConstructor,
	}

	// ErrorsInErrorsFile errors when a package declares more than one sentinel error outside errors.go.
	ErrorsInErrorsFile = &analysis.Analyzer{
		Name:     "errors_in_errors_file",
		Doc:      "error when a package with more than one sentinel error declares any of them outside errors.go",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runErrorsInErrorsFile,
	}
)

// runConstantsAtTop reports constants declared below the functions that use
// them. Constants are the vocabulary of a file; a reader meets them before the
// code that spends them, not after.
func runConstantsAtTop(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if isTestFile(pass, file) {
			continue
		}
		var seenFunc bool
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				seenFunc = true
			case *ast.GenDecl:
				if d.Tok == token.CONST && seenFunc {
					pass.Reportf(d.Pos(), "constants belong at the top of the file, above the functions that use them")
				}
			}
		}
	}
	return nil, nil
}

// runTypeBeforeConstructor reports a constructor separated from its type, and
// a namesake file that opens with something other than its subject.
func runTypeBeforeConstructor(pass *analysis.Pass) (interface{}, error) {
	types := make(map[string]declSite)
	constructors := make(map[string]declSite)

	for _, file := range pass.Files {
		if isTestFile(pass, file) {
			continue
		}
		filename := pass.Fset.Position(file.Pos()).Filename
		for index, decl := range file.Decls {
			site := declSite{filename: filename, index: index, pos: decl.Pos()}
			switch d := decl.(type) {
			case *ast.GenDecl:
				if d.Tok != token.TYPE || len(d.Specs) != 1 {
					continue
				}
				spec, ok := d.Specs[0].(*ast.TypeSpec)
				if !ok {
					continue
				}
				types[spec.Name.Name] = site
			case *ast.FuncDecl:
				if d.Recv != nil || !strings.HasPrefix(d.Name.Name, "New") {
					continue
				}
				constructors[strings.TrimPrefix(d.Name.Name, "New")] = site
			}
		}
		checkNamesakeFileOpensWithType(pass, file, filename)
	}

	for typeName, constructor := range constructors {
		declaration, ok := types[typeName]
		if !ok {
			continue
		}
		if declaration.filename != constructor.filename {
			pass.Reportf(constructor.pos, "constructor New%s belongs beside type %s in %s", typeName, typeName, filepath.Base(declaration.filename))
			continue
		}
		if constructor.index != declaration.index+1 {
			pass.Reportf(constructor.pos, "constructor New%s must be declared immediately after type %s", typeName, typeName)
		}
	}

	return nil, nil
}

// checkNamesakeFileOpensWithType reports a file named after its package that
// declares types but reaches a function first. Opening store.go should show
// the reader Store, not a helper that happens to sort alphabetically.
func checkNamesakeFileOpensWithType(pass *analysis.Pass, file *ast.File, filename string) {
	base := strings.TrimSuffix(filepath.Base(filename), ".go")
	if base != pass.Pkg.Name() {
		return
	}
	if !declaresType(file) {
		return
	}
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Tok == token.TYPE {
				return // the file opens with its subject
			}
		case *ast.FuncDecl:
			pass.Reportf(d.Pos(), "%s.go must open with its type declarations; move %s below them", pass.Pkg.Name(), d.Name.Name)
			return
		}
	}
}

func declaresType(file *ast.File) bool {
	for _, decl := range file.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.TYPE {
			return true
		}
	}
	return false
}

// runErrorsInErrorsFile reports sentinel errors scattered across a package. One
// error can live beside the code that returns it; once there are several, a
// reader looking for the package's failure modes needs a single place to look.
func runErrorsInErrorsFile(pass *analysis.Pass) (interface{}, error) {
	var strays []declSite
	var total int

	for _, file := range pass.Files {
		if isTestFile(pass, file) {
			continue
		}
		filename := pass.Fset.Position(file.Pos()).Filename
		inErrorsFile := filepath.Base(filename) == "errors.go"
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				value, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for index, name := range value.Names {
					if index >= len(value.Values) || !isErrorConstructor(value.Values[index]) {
						continue
					}
					total++
					if !inErrorsFile {
						strays = append(strays, declSite{filename: filename, pos: name.Pos(), name: name.Name})
					}
				}
			}
		}
	}

	if total < 2 {
		return nil, nil
	}
	for _, stray := range strays {
		pass.Reportf(stray.pos, "package declares %d sentinel errors; move %s into errors.go, grouped by the feature that returns it", total, stray.name)
	}

	return nil, nil
}

// isErrorConstructor reports whether the expression builds a sentinel error.
func isErrorConstructor(expr ast.Expr) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}
	return (pkg.Name == "errors" && sel.Sel.Name == "New") || (pkg.Name == "fmt" && sel.Sel.Name == "Errorf")
}

// declSite is where a declaration was found: which file, how far down its
// declaration list, and the position to report against.
type declSite struct {
	filename string
	index    int
	pos      token.Pos
	name     string
}

func isTestFile(pass *analysis.Pass, file *ast.File) bool {
	return strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go")
}
