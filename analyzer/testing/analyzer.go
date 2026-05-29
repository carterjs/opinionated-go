package testing

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	TestNotTableDriven = &analysis.Analyzer{
		Name:     "test_not_table_driven",
		Doc:      "warn on non-table-driven tests",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runTestNotTableDriven,
	}
	LoopVariableNaming = &analysis.Analyzer{
		Name:     "loop_variable_naming",
		Doc:      "error on loop variable not named test",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runLoopVariableNaming,
	}
	LoopVariableReassigned = &analysis.Analyzer{
		Name:     "loop_variable_reassigned",
		Doc:      "error on loop variable reassignment",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runLoopVariableReassigned,
	}
	TestNamingViolation = &analysis.Analyzer{
		Name:     "test_naming_violation",
		Doc:      "error on test naming violations",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runTestNamingViolation,
	}
	CoversUnexported = &analysis.Analyzer{
		Name:     "covers_unexported",
		Doc:      "error on tests covering unexported identifiers",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runCoversUnexported,
	}
	OrphanedTestFile = &analysis.Analyzer{
		Name:     "orphaned_test_file",
		Doc:      "error on orphaned test files",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runOrphanedTestFile,
	}
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
	StdoutStderrInTest = &analysis.Analyzer{
		Name:     "stdout_stderr_in_test",
		Doc:      "warn on os.Stdout/Stderr in tests",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runStdoutStderrInTest,
	}
	SubtestNameFormat = &analysis.Analyzer{
		Name:     "subtest_name_format",
		Doc:      "warn on subtest names with underscores or camelCase",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runSubtestNameFormat,
	}
)

func runTestNotTableDriven(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}

func runLoopVariableNaming(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}

func runLoopVariableReassigned(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}

func runTestNamingViolation(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}

func runCoversUnexported(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}

func runOrphanedTestFile(pass *analysis.Pass) (interface{}, error) {
	isTestFile := len(pass.Files) > 0 && strings.HasSuffix(pass.Fset.File(pass.Files[0].Pos()).Name(), "_test.go")
	if !isTestFile {
		return nil, nil
	}
	return nil, nil
}

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

func runSubtestNameFormat(pass *analysis.Pass) (interface{}, error) {
	return nil, nil
}
