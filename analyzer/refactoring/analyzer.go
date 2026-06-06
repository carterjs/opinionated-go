package refactoring

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
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
		return false // Don't inline unused functions
	}
	// One-liner used fewer than 3 times
	if info.lineCount == 1 && info.uses < 3 {
		return true
	}
	// Less than 10 lines and used only once
	if info.lineCount < 10 && info.uses == 1 {
		return true
	}
	return false
}

func countFuncLines(fn *ast.FuncDecl) int {
	if fn.Body == nil {
		return 0
	}
	// Count statements in function body as approximation of lines
	return countStmts(fn.Body.List)
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
		return countStmts(s.List)
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
		return 2 // Approximate
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
