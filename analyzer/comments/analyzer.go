package comments

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var (
	// ExportedCommentFormat checks that exported symbols have properly formatted doc comments, package comments start with "Package", and only modern build/embed directives are used.
	ExportedCommentFormat = &analysis.Analyzer{
		Name:     "exported_comment_format",
		Doc:      "error on comments for exported symbols not starting with the symbol name, package comments not starting with Package, and legacy build/embed directive syntax",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runExportedCommentFormat,
	}
)

func runExportedCommentFormat(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// Skip test files - they don't need doc comments on test functions
		isTestFile := strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go")

		// Check package comment
		if file.Doc != nil {
			checkPackageComment(pass, file.Name.Name, file.Doc)
		}

		// Check for legacy build/embed directives
		checkLegacyDirectives(pass, file.Comments)

		// Check exported symbol comments (skip in test files)
		if !isTestFile {
			for _, decl := range file.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if isExported(d.Name.Name) {
						checkCommentFormat(pass, d.Name.Name, d.Doc, d.Name.Pos())
					}
				case *ast.GenDecl:
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							if isExported(s.Name.Name) {
								doc := s.Doc
								if doc == nil {
									doc = d.Doc
								}
								checkCommentFormat(pass, s.Name.Name, doc, s.Name.Pos())
							}
						case *ast.ValueSpec:
							for _, name := range s.Names {
								if isExported(name.Name) {
									doc := s.Doc
									if doc == nil {
										doc = d.Doc
									}
									checkCommentFormat(pass, name.Name, doc, name.Pos())
								}
							}
						}
					}
				}
			}
		}
	}
	return nil, nil
}

func checkCommentFormat(pass *analysis.Pass, symbolName string, doc *ast.CommentGroup, pos token.Pos) {
	if doc == nil {
		pass.Reportf(pos, "exported symbol %q must have a doc comment", symbolName)
		return
	}

	text := doc.Text()
	if text == "" {
		pass.Reportf(doc.Pos(), "exported symbol %q must have a doc comment", symbolName)
		return
	}

	if !strings.HasPrefix(text, symbolName) {
		pass.Reportf(doc.Pos(), "exported symbol %q must have a comment starting with %q", symbolName, symbolName)
		return
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if i > 0 && i < len(lines)-1 {
				pass.Reportf(doc.Pos(), "comments should not have empty lines used as dividers")
				return
			}
		}
		if len(trimmed) >= 3 {
			if (strings.HasPrefix(trimmed, "///") || strings.HasPrefix(trimmed, "---")) &&
				(strings.Count(trimmed, "/") >= 3 || strings.Count(trimmed, "-") >= 3) {
				pass.Reportf(doc.Pos(), "comments should not use repeated slashes or dashes as dividers")
				return
			}
		}
	}
}

func checkPackageComment(pass *analysis.Pass, pkgName string, doc *ast.CommentGroup) {
	if doc == nil {
		return
	}

	text := doc.Text()
	if text == "" {
		return
	}

	expected := "Package " + pkgName
	if !strings.HasPrefix(text, expected) {
		pass.Reportf(doc.Pos(), "package comment must start with %q", expected)
	}
}

func checkLegacyDirectives(pass *analysis.Pass, comments []*ast.CommentGroup) {
	for _, group := range comments {
		if group == nil {
			continue
		}
		for _, comment := range group.List {
			text := comment.Text
			// Check for legacy +build directives
			if strings.HasPrefix(text, "// +build ") {
				pass.Reportf(comment.Pos(), "use //go:build instead of legacy // +build syntax")
			}
			// Check for legacy go:embed with space (should be //go:embed)
			if strings.HasPrefix(text, "// go:embed ") {
				pass.Reportf(comment.Pos(), "use //go:embed (no space after //) instead of // go:embed")
			}
		}
	}
}

func isExported(name string) bool {
	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}
