package testing

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	// t.Context() requires Go 1.20+
	ContextBackgroundInTest = &analysis.Analyzer{
		Name:     "context_background_in_test",
		Doc:      "error on context.Background in tests",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runContextBackgroundInTest,
	}
	FmtPrintInTest = &analysis.Analyzer{
		Name:     "fmt_print_in_test",
		Doc:      "error on fmt.Print in tests",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runFmtPrintInTest,
	}
	// t.Output() requires Go 1.21+
	StdoutStderrInTest = &analysis.Analyzer{
		Name:     "stdout_stderr_in_test",
		Doc:      "warn on os.Stdout/Stderr in tests",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runStdoutStderrInTest,
	}
	NoSleepInTests = &analysis.Analyzer{
		Name:     "no_sleep_in_tests",
		Doc:      "error on time.Sleep in tests (use synctest or test utils instead)",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNoSleepInTests,
	}
	TestNameInLoop = &analysis.Analyzer{
		Name:     "test_name_in_loop",
		Doc:      "error on using loop variable .name field in conditionals",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runTestNameInLoop,
	}
)

func runContextBackgroundInTest(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call := node.(*ast.CallExpr)
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "context" && (sel.Sel.Name == "Background" || sel.Sel.Name == "TODO") {
				pass.Reportf(call.Pos(), "use t.Context() instead of context.Background or context.TODO")
			}
		}
	})
	return nil, nil
}

func runFmtPrintInTest(pass *analysis.Pass) (interface{}, error) {
	printFuncs := map[string]bool{
		"Print":   true,
		"Printf":  true,
		"Println": true,
	}
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call := node.(*ast.CallExpr)
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "fmt" && printFuncs[sel.Sel.Name] {
				pass.Reportf(call.Pos(), "use t.Log or t.Logf for test output, not fmt.Print")
			}
		}
	})
	return nil, nil
}

func runStdoutStderrInTest(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.SelectorExpr)(nil)}, func(node ast.Node) {
		sel := node.(*ast.SelectorExpr)
		if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "os" && (sel.Sel.Name == "Stdout" || sel.Sel.Name == "Stderr") {
			pass.Reportf(sel.Pos(), "use t.Output() when an io.Writer is required in tests, not os.Stdout/Stderr")
		}
	})
	return nil, nil
}

func runNoSleepInTests(pass *analysis.Pass) (interface{}, error) {
	// Only check test files
	if len(pass.Files) == 0 {
		return nil, nil
	}
	filename := pass.Fset.File(pass.Files[0].Pos()).Name()
	if !strings.HasSuffix(filename, "_test.go") {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	var inSynctest bool

	inspect.Preorder([]ast.Node{(*ast.CallExpr)(nil)}, func(node ast.Node) {
		call := node.(*ast.CallExpr)

		// Check if we're in a synctest context (synctest.Run call)
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "synctest" && sel.Sel.Name == "Run" {
				inSynctest = true
				defer func() { inSynctest = false }()
			}
		}

		// Check for time.Sleep
		if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "time" && sel.Sel.Name == "Sleep" {
				if !inSynctest {
					pass.Reportf(call.Pos(), "avoid time.Sleep in tests; use synctest or test helpers for deterministic timing")
				}
			}
		}
	})
	return nil, nil
}

func runTestNameInLoop(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	inspect.Preorder([]ast.Node{(*ast.RangeStmt)(nil), (*ast.ForStmt)(nil)}, func(node ast.Node) {
		var loopVar string
		var body *ast.BlockStmt

		switch n := node.(type) {
		case *ast.RangeStmt:
			// Get the loop variable (value part for range loops)
			if ident, ok := n.Value.(*ast.Ident); ok {
				loopVar = ident.Name
			}
			body = n.Body
		case *ast.ForStmt:
			// For standard for loops, we don't check (they're not table-driven tests typically)
			return
		}

		if loopVar == "" || loopVar == "_" {
			return
		}

		// Check loop body for conditionals using loopVar.name
		ast.Inspect(body, func(n ast.Node) bool {
			switch stmt := n.(type) {
			case *ast.IfStmt:
				if usesLoopVarName(stmt.Cond, loopVar) {
					pass.Reportf(stmt.Pos(), "avoid checking %s.name in conditionals; restructure test data or use %s.skip", loopVar, loopVar)
				}
			case *ast.SwitchStmt:
				if usesLoopVarName(stmt.Tag, loopVar) {
					pass.Reportf(stmt.Pos(), "avoid switching on %s.name; restructure test data or use %s.skip", loopVar, loopVar)
				}
			}
			return true
		})
	})

	return nil, nil
}

func usesLoopVarName(expr ast.Expr, loopVar string) bool {
	if expr == nil {
		return false
	}
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok {
			if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == loopVar && sel.Sel.Name == "name" {
				found = true
				return false // Stop searching once found
			}
		}
		return true
	})
	return found
}
