package comments

import (
	"go/ast"
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
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

	// DocCommentTooLong warns on a doc comment that has outgrown its declaration.
	DocCommentTooLong = &analysis.Analyzer{
		Name:     "doc_comment_too_long",
		Doc:      "warn on a doc comment longer than 40 words; past that it is package documentation, or the declaration is doing too much",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runDocCommentTooLong,
	}

	// InlineCommentTooLong warns on an inline comment that runs past one line.
	InlineCommentTooLong = &analysis.Analyzer{
		Name:     "inline_comment_too_long",
		Doc:      "warn on a multi-line inline comment; a why that needs more than a line is a doc comment or a named helper",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runInlineCommentTooLong,
	}
)

// maxDocWords is the point past which a doc comment has stopped documenting one declaration.
const maxDocWords = 40

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

// runDetachedComment reports file-scope comments that float free of the code
// they describe. A comment earns its place by sitting directly above something;
// once a blank line comes between them, or nothing follows at all, the reader
// has to guess what it refers to.
func runDetachedComment(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go") {
			continue
		}

		bodies := functionBodies(file)
		occupied := codeLines(pass.Fset, file)

		for _, group := range file.Comments {
			if isDirectiveGroup(group) {
				continue
			}
			// A header above the package clause is a license or build banner,
			// not documentation for a symbol.
			if group.End() < file.Package {
				continue
			}
			// Comments inside a function body are the author's running
			// commentary, judged by a different rule.
			if withinAny(bodies, group.Pos()) {
				continue
			}
			if occupied[pass.Fset.Position(group.Pos()).Line] {
				continue // trailing comment on a line of code
			}
			if occupied[pass.Fset.Position(group.End()).Line+1] {
				continue // bumps up against the line below it
			}
			pass.Reportf(group.Pos(), "comment must bump up against the symbol it documents; remove the blank line or delete the comment")
		}
	}
	return nil, nil
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

// runDocCommentTooLong reports a doc comment that has outgrown its
// declaration. Words are counted rather than sentences because no sentence
// count can tell "e.g." or a URL from a full stop.
func runDocCommentTooLong(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil), (*ast.GenDecl)(nil)}, func(node ast.Node) {
		var doc *ast.CommentGroup
		var name string
		switch decl := node.(type) {
		case *ast.FuncDecl:
			doc, name = decl.Doc, decl.Name.Name
		case *ast.GenDecl:
			doc, name = decl.Doc, genDeclName(decl)
		}
		if doc == nil {
			return
		}

		if words := countDocWords(doc); words > maxDocWords {
			pass.Reportf(node.Pos(), "doc comment on %s runs %d words; keep a declaration to %d and move the rest to the package comment", name, words, maxDocWords)
		}
	})
	return nil, nil
}

// runInlineCommentTooLong reports an inline comment group spanning more than one line.
func runInlineCommentTooLong(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		documented := documentedPositions(file)
		for _, group := range file.Comments {
			if documented[group.Pos()] || len(group.List) < 2 {
				continue
			}
			if isDirective(group.List[0].Text) {
				continue
			}
			lines := pass.Fset.Position(group.End()).Line - pass.Fset.Position(group.Pos()).Line + 1
			pass.Reportf(group.Pos(), "inline comment runs %d lines; a why that needs more than one line is a doc comment on the function or a named helper", lines)
		}
	}
	return nil, nil
}

// countDocWords counts the words in a doc comment, ignoring directives and paragraph breaks.
func countDocWords(doc *ast.CommentGroup) int {
	words := 0
	for _, comment := range doc.List {
		if isDirective(comment.Text) {
			continue
		}
		text := strings.TrimPrefix(comment.Text, "//")
		text = strings.TrimPrefix(text, "/*")
		text = strings.TrimSuffix(text, "*/")
		words += len(strings.Fields(text))
	}
	return words
}

// documentedPositions collects the comment groups already attached to a declaration.
func documentedPositions(file *ast.File) map[token.Pos]bool {
	documented := make(map[token.Pos]bool)
	if file.Doc != nil {
		documented[file.Doc.Pos()] = true
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch decl := node.(type) {
		case *ast.FuncDecl:
			if decl.Doc != nil {
				documented[decl.Doc.Pos()] = true
			}
		case *ast.GenDecl:
			if decl.Doc != nil {
				documented[decl.Doc.Pos()] = true
			}
		case *ast.TypeSpec:
			if decl.Doc != nil {
				documented[decl.Doc.Pos()] = true
			}
		case *ast.ValueSpec:
			if decl.Doc != nil {
				documented[decl.Doc.Pos()] = true
			}
		case *ast.Field:
			if decl.Doc != nil {
				documented[decl.Doc.Pos()] = true
			}
		}
		return true
	})
	return documented
}

// isDirective reports whether text is a compiler directive rather than prose.
func isDirective(text string) bool {
	body := strings.TrimPrefix(text, "//")
	if body == text || body == "" || strings.HasPrefix(body, " ") {
		return false
	}
	name, _, found := strings.Cut(body, ":")
	return found && name != "" && !strings.ContainsAny(name, " \t")
}

// genDeclName names a declaration for a diagnostic.
func genDeclName(decl *ast.GenDecl) string {
	if len(decl.Specs) == 0 {
		return decl.Tok.String()
	}
	switch spec := decl.Specs[0].(type) {
	case *ast.TypeSpec:
		return spec.Name.Name
	case *ast.ValueSpec:
		if len(spec.Names) > 0 {
			return spec.Names[0].Name
		}
	}
	return decl.Tok.String()
}
