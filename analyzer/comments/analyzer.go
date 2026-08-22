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

	// DetachedComment errors on file-scope comments that do not bump up against the symbol they document.
	DetachedComment = &analysis.Analyzer{
		Name:     "detached_comment",
		Doc:      "error on file-scope comments separated from the declaration below them, or documenting nothing at all",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runDetachedComment,
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

	// A blank "//" line between paragraphs is fine. Repeated slashes or dashes
	// used as a visual divider are not.
	for _, line := range strings.Split(doc.Text(), "\n") {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) < 3 {
			continue
		}
		if (strings.HasPrefix(trimmed, "///") || strings.HasPrefix(trimmed, "---")) &&
			(strings.Count(trimmed, "/") >= 3 || strings.Count(trimmed, "-") >= 3) {
			pass.Reportf(doc.Pos(), "comments should not use repeated slashes or dashes as dividers")
			return
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

// runDetachedComment reports file-scope comments that float free of a
// declaration. The parser only records a comment group as a node's Doc when no
// blank line separates the two, so any remaining group that sits on its own
// line outside a function body is either separated from what it describes or
// describes nothing.
func runDetachedComment(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go") {
			continue
		}

		docs := documentationGroups(file)
		bodies := functionBodies(file)
		occupied := codeLines(pass.Fset, file)

		for _, group := range file.Comments {
			if docs[group] || isDirectiveGroup(group) {
				continue
			}
			// A header above the package clause is a license or build banner,
			// not documentation for a symbol.
			if group.End() < file.Package {
				continue
			}
			if withinAny(bodies, group.Pos()) {
				continue
			}
			if occupied[pass.Fset.Position(group.Pos()).Line] {
				continue // trailing comment on a line of code
			}
			pass.Reportf(group.Pos(), "comment must bump up against the symbol it documents; remove the blank line or delete the comment")
		}
	}
	return nil, nil
}

// documentationGroups collects every comment group the parser attached to a
// declaration as its documentation.
func documentationGroups(file *ast.File) map[*ast.CommentGroup]bool {
	docs := map[*ast.CommentGroup]bool{file.Doc: true}
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.GenDecl:
			docs[n.Doc] = true
		case *ast.FuncDecl:
			docs[n.Doc] = true
		case *ast.TypeSpec:
			docs[n.Doc] = true
			docs[n.Comment] = true
		case *ast.ValueSpec:
			docs[n.Doc] = true
			docs[n.Comment] = true
		case *ast.ImportSpec:
			docs[n.Doc] = true
			docs[n.Comment] = true
		case *ast.Field:
			docs[n.Doc] = true
			docs[n.Comment] = true
		}
		return true
	})
	delete(docs, nil)
	return docs
}

// functionBodies returns the position ranges of every function body in the
// file. Comments inside a body are the author's running commentary, not
// documentation for a symbol.
func functionBodies(file *ast.File) [][2]token.Pos {
	var bodies [][2]token.Pos
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.FuncDecl:
			if n.Body != nil {
				bodies = append(bodies, [2]token.Pos{n.Body.Pos(), n.Body.End()})
			}
		case *ast.FuncLit:
			bodies = append(bodies, [2]token.Pos{n.Body.Pos(), n.Body.End()})
		}
		return true
	})
	return bodies
}

// codeLines marks every line on which some syntax other than a comment begins
// or ends, so that a comment sharing one of those lines is recognised as
// trailing rather than detached.
func codeLines(fset *token.FileSet, file *ast.File) map[int]bool {
	lines := make(map[int]bool)
	ast.Inspect(file, func(node ast.Node) bool {
		if node == nil {
			return false
		}
		switch node.(type) {
		case *ast.File, *ast.Comment, *ast.CommentGroup:
			return true
		}
		lines[fset.Position(node.Pos()).Line] = true
		lines[fset.Position(node.End()).Line] = true
		return true
	})
	return lines
}

// isDirectiveGroup reports whether the group is a compiler or tool directive
// rather than prose.
func isDirectiveGroup(group *ast.CommentGroup) bool {
	for _, comment := range group.List {
		text := comment.Text
		if strings.HasPrefix(text, "//go:") || strings.HasPrefix(text, "//nolint") || strings.HasPrefix(text, "// +build") {
			return true
		}
	}
	return false
}

func withinAny(ranges [][2]token.Pos, pos token.Pos) bool {
	for _, span := range ranges {
		if pos >= span[0] && pos < span[1] {
			return true
		}
	}
	return false
}

func isExported(name string) bool {
	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}
