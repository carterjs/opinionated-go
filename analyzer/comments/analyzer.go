package comments

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

var (
	ExportedCommentFormat = &analysis.Analyzer{
		Name:     "exported_comment_format",
		Doc:      "error on comments for exported symbols not starting with the symbol name",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runExportedCommentFormat,
	}
)

func runExportedCommentFormat(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		for _, decl := range file.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if isExported(d.Name.Name) {
					checkCommentFormat(pass, d.Name.Name, d.Doc)
				}
			case *ast.GenDecl:
				if d.Doc != nil {
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							if isExported(s.Name.Name) {
								checkCommentFormat(pass, s.Name.Name, s.Doc)
							}
						case *ast.ValueSpec:
							for _, name := range s.Names {
								if isExported(name.Name) {
									checkCommentFormat(pass, name.Name, s.Doc)
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

func checkCommentFormat(pass *analysis.Pass, symbolName string, doc *ast.CommentGroup) {
	if doc == nil {
		return
	}

	text := doc.Text()
	if text == "" {
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

func isExported(name string) bool {
	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}
