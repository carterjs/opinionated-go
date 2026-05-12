package testing

import (
	"golang.org/x/tools/go/analysis"
)

var (
	TestNotTableDriven = &analysis.Analyzer{
		Name: "test_not_table_driven",
		Doc:  "warn on non-table-driven tests",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	LoopVariableNaming = &analysis.Analyzer{
		Name: "loop_variable_naming",
		Doc:  "error on loop variable not named test",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	LoopVariableReassigned = &analysis.Analyzer{
		Name: "loop_variable_reassigned",
		Doc:  "error on loop variable reassignment",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	TestNamingViolation = &analysis.Analyzer{
		Name: "test_naming_violation",
		Doc:  "error on test naming violations",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	CoversUnexported = &analysis.Analyzer{
		Name: "covers_unexported",
		Doc:  "error on tests covering unexported identifiers",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	OrphanedTestFile = &analysis.Analyzer{
		Name: "orphaned_test_file",
		Doc:  "error on orphaned test files",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	ContextBackgroundInTest = &analysis.Analyzer{
		Name: "context_background_in_test",
		Doc:  "error on context.Background in tests",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	FmtPrintInTest = &analysis.Analyzer{
		Name: "fmt_print_in_test",
		Doc:  "error on fmt.Print in tests",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	StdoutStderrInTest = &analysis.Analyzer{
		Name: "stdout_stderr_in_test",
		Doc:  "warn on os.Stdout/Stderr in tests",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
	SubtestNameFormat = &analysis.Analyzer{
		Name: "subtest_name_format",
		Doc:  "warn on subtest names with underscores or camelCase",
		Run:  func(pass *analysis.Pass) (interface{}, error) { return nil, nil },
	}
)
