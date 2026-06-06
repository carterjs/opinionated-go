package refactoring

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	// InlineableFunction suggests inlining small unexported functions with few callers.
	InlineableFunction = &analysis.Analyzer{
		Name:     "inlineable_function",
		Doc:      "suggest inlining small unexported functions with few callers",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runInlineableFunction,
	}
)

func runInlineableFunction(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Map function names to their usage count and line count
	funcInfo := make(map[string]*funcData)

	// First pass: collect function definitions
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		// Only check unexported functions
		if fn.Name.Name[0] < 'a' || fn.Name.Name[0] > 'z' {
			return
		}

		if fn.Body != nil {
			lineCount := countFuncLines(fn)
			funcInfo[fn.Name.Name] = &funcData{
				name:      fn.Name.Name,
				lineCount: lineCount,
				funcDecl:  fn,
				uses:      0,
			}
		}
	})

	// Second pass: count usages in call expressions
	inspect.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call := node.(*ast.CallExpr)
		if ident, ok := call.Fun.(*ast.Ident); ok {
			if info, ok := funcInfo[ident.Name]; ok {
				info.uses++
			}
		}
	})

	// Check inlineability
	for _, info := range funcInfo {
		if shouldInline(info) {
			pass.Reportf(info.funcDecl.Name.Pos(), "function %q should be inlined (%d lines, %d uses)",
				info.name, info.lineCount, info.uses)
		}
	}

	return nil, nil
}

func shouldInline(info *funcData) bool {
	if info.uses == 0 {
		return false
	}
	// Flag single-statement functions used 1-2 times
	if info.lineCount == 1 && info.uses <= 2 {
		return true
	}
	// Flag 2-statement functions used only once
	if info.lineCount <= 2 && info.uses == 1 {
		return true
	}
	return false
}

func countFuncLines(fn *ast.FuncDecl) int {
	if fn.Body == nil {
		return 0
	}
	// Estimate lines based on statement count with conservative scaling
	// Statement count doesn't map directly to lines due to nesting
	return max(1, countStmts(fn.Body.List)*2)
}

func countStmts(stmts []ast.Stmt) int {
	count := 0
	for _, stmt := range stmts {
		count += countStmt(stmt)
	}
	return count
}

func countStmt(stmt ast.Stmt) int {
	switch s := stmt.(type) {
	case *ast.BlockStmt:
		return 1 + countStmts(s.List)
	case *ast.IfStmt:
		count := 1
		if s.Body != nil {
			count += countStmt(s.Body)
		}
		if s.Else != nil {
			count += countStmt(s.Else)
		}
		return count
	case *ast.ForStmt, *ast.RangeStmt, *ast.SwitchStmt:
		return 3
	default:
		return 1
	}
}

type funcData struct {
	name      string
	lineCount int
	funcDecl  *ast.FuncDecl
	uses      int
}
