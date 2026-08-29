package testing

import (
	"go/ast"
	"go/token"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var (
	// ContextBackgroundInTest errors on [context.Background] in test files (requires Go 1.24+ for [testing.T.Context]).
	ContextBackgroundInTest = &analysis.Analyzer{
		Name:     "context_background_in_test",
		Doc:      "error on context.Background in _test.go files",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runContextBackgroundInTest,
	}
	// FmtPrintInTest errors on [fmt.Print] in tests.
	FmtPrintInTest = &analysis.Analyzer{
		Name:     "fmt_print_in_test",
		Doc:      "error on fmt.Print in tests",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runFmtPrintInTest,
	}
	// StdoutStderrInTest warns on [os.Stdout]/[os.Stderr] in tests (requires Go 1.21+ for [testing.T.Output]).
	StdoutStderrInTest = &analysis.Analyzer{
		Name:     "stdout_stderr_in_test",
		Doc:      "warn on os.Stdout/Stderr in tests",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runStdoutStderrInTest,
	}
	// NoSleepInTests errors on [time.Sleep] in tests.
	NoSleepInTests = &analysis.Analyzer{
		Name:     "no_sleep_in_tests",
		Doc:      "error on time.Sleep in tests (use synctest or test utils instead)",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runNoSleepInTests,
	}
	// TestNameInLoop errors on using loop variable .name field in conditionals.
	TestNameInLoop = &analysis.Analyzer{
		Name:     "test_name_in_loop",
		Doc:      "error on using loop variable .name field in conditionals",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runTestNameInLoop,
	}

	// TestFunctionNaming errors on a test whose name does not name an exported function or method.
	TestFunctionNaming = &analysis.Analyzer{
		Name:     "test_function_naming",
		Doc:      "error on a test function that is not Test<Function> or Test<Type>_<Method> for an exported subject in the package",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runTestFunctionNaming,
	}

	// TestCaseNaming errors on table case names that do not state a behavior, and on duplicates within a table.
	TestCaseNaming = &analysis.Analyzer{
		Name:     "test_case_naming",
		Doc:      "error on table case names that name a fixture or an outcome class rather than a behavior, and on duplicate names within one table",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runTestCaseNaming,
	}

	// TableDrivenTest warns on a test function that is not table-driven.
	TableDrivenTest = &analysis.Analyzer{
		Name:     "table_driven_test",
		Doc:      "warn on a test function with no table; a new behavior is a new row, not a new function",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runTableDrivenTest,
	}

	// TestParallel warns on a test or subtest that does not call t.Parallel first.
	TestParallel = &analysis.Analyzer{
		Name:     "test_parallel",
		Doc:      "warn on a test function or subtest that does not call t.Parallel() as its first statement",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run:      runTestParallel,
	}
)

// testingReceivers are the settled names for a testing value, which a subtest hangs off.
var testingReceivers = map[string]bool{
	"t": true,
	"b": true,
	"f": true,
}

// bannedOpeners start a case name with the test's voice rather than the subject's.
var bannedOpeners = map[string]bool{
	"should": true,
	"it":     true,
	"will":   true,
	"would":  true,
	"test":   true,
	"tests":  true,
	"the":    true,
	"when":   true,
}

// bannedNames are outcome classes: they say which half of the table a case sits in, not what the subject does.
var bannedNames = map[string]bool{
	"success":      true,
	"successful":   true,
	"failure":      true,
	"failing":      true,
	"error":        true,
	"error case":   true,
	"errors":       true,
	"happy path":   true,
	"sad path":     true,
	"valid":        true,
	"invalid":      true,
	"ok":           true,
	"positive":     true,
	"negative":     true,
	"base case":    true,
	"default case": true,
}

func runContextBackgroundInTest(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		// t.Context() only exists inside a test, so the advice only applies to
		// test files. Production code is covered by
		// [concurrency.ContextBackgroundOutsideMain].
		if fn.Body == nil || !strings.HasSuffix(pass.Fset.File(fn.Pos()).Name(), "_test.go") {
			return
		}

		count, first := countRootContexts(fn.Body)
		switch {
		case count == 0:
		case count == 1:
			pass.Reportf(first, "use t.Context() instead of context.Background or context.TODO")
		default:
			pass.Reportf(first, "use t.Context() instead of context.Background or context.TODO (%d times in %s)", count, fn.Name.Name)
		}
	})
	return nil, nil
}

// countRootContexts reports how many times a body roots its own context, and
// where it first does so.
func countRootContexts(body *ast.BlockStmt) (int, token.Pos) {
	var count int
	var first token.Pos
	ast.Inspect(body, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok || ident.Name != "context" {
			return true
		}
		if sel.Sel.Name != "Background" && sel.Sel.Name != "TODO" {
			return true
		}
		if count == 0 {
			first = call.Pos()
		}
		count++
		return true
	})
	return count, first
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
		// Only check in test files
		filename := pass.Fset.File(node.Pos()).Name()
		if !strings.HasSuffix(filename, "_test.go") {
			return
		}

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

// runTestFunctionNaming reports a test whose name resolves to no exported
// subject. A name with nowhere to land is a case that wanted a row.
func runTestFunctionNaming(pass *analysis.Pass) (interface{}, error) {
	// An external test package sees no source files, so it has no subjects to check against.
	subjects := packageSubjects(pass)
	if subjects == nil {
		return nil, nil
	}

	forEachTestFunction(pass, func(fn *ast.FuncDecl) {
		subject := strings.TrimPrefix(fn.Name.Name, "Test")
		if subject == "" {
			return
		}

		parts := strings.Split(subject, "_")
		switch len(parts) {
		case 1:
			if !subjects.functions[subject] {
				pass.Reportf(fn.Pos(), "%s names no exported function in this package; use Test<Function> or Test<Type>_<Method>, and add a case to the existing table instead of a new test function", fn.Name.Name)
			}
		case 2:
			if !subjects.methods[subject] {
				pass.Reportf(fn.Pos(), "%s names no exported method in this package; use Test<Function> or Test<Type>_<Method>, and add a case to the existing table instead of a new test function", fn.Name.Name)
			}
		default:
			pass.Reportf(fn.Pos(), "%s carries more than one underscore; a test splits by subject, never by outcome", fn.Name.Name)
		}
	})
	return nil, nil
}

// runTestCaseNaming reports case names that name a fixture or an outcome class
// rather than a behavior, and names repeated within one table. The name is the
// sentence a failure prints, so it has to say what broke on its own.
func runTestCaseNaming(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.CompositeLit)(nil)}, func(node ast.Node) {
		table := node.(*ast.CompositeLit)
		if !isTestFile(pass, table.Pos()) {
			return
		}
		if _, ok := table.Type.(*ast.ArrayType); !ok {
			return
		}

		seen := make(map[string]bool)
		for _, element := range table.Elts {
			row, ok := element.(*ast.CompositeLit)
			if !ok {
				continue
			}
			name := caseName(row)
			if name == nil {
				continue
			}

			checkCaseName(pass, name.position, name.text)
			if seen[name.text] {
				pass.Reportf(name.position, "case name %q already used in this table; two cases wanting one name are one case, or the name is not specific enough", name.text)
			}
			seen[name.text] = true
		}
	})
	return nil, nil
}

// runTableDrivenTest reports a test function with no table to add a case to.
func runTableDrivenTest(pass *analysis.Pass) (interface{}, error) {
	forEachTestFunction(pass, func(fn *ast.FuncDecl) {
		if delegatesToHarness(fn.Body) {
			return
		}

		ranges := false
		ast.Inspect(fn.Body, func(node ast.Node) bool {
			if _, ok := node.(*ast.RangeStmt); ok {
				ranges = true
				return false
			}
			return true
		})
		if !ranges {
			pass.Reportf(fn.Pos(), "%s is not table-driven; a table is what lets the next behavior be a row rather than another test function", fn.Name.Name)
		}
	})
	return nil, nil
}

// delegatesToHarness reports whether a test body is nothing but calls into a
// harness that owns the cases itself, as an analysistest run does. There is no
// table to add a row to, so there is no table to require.
func delegatesToHarness(body *ast.BlockStmt) bool {
	for _, statement := range body.List {
		expression, ok := statement.(*ast.ExprStmt)
		if !ok {
			return false
		}
		call, ok := expression.X.(*ast.CallExpr)
		if !ok || isSubtestCall(call) {
			return false
		}
	}
	return len(body.List) > 0
}

// runTestParallel reports a test or subtest that does not open with
// t.Parallel(). A test or subtest that calls t.Setenv or t.Chdir anywhere
// within it — including inside its own subtests — is exempt: both affect
// process-wide state, and the testing package itself panics if either is
// combined with a parallel ancestor, so requiring Parallel there would be
// asking for code that cannot run.
func runTestParallel(pass *analysis.Pass) (interface{}, error) {
	forEachTestFunction(pass, func(fn *ast.FuncDecl) {
		if !usesNonParallelSafeAPI(fn.Body) && !opensWithParallel(fn.Body) {
			pass.Reportf(fn.Pos(), "%s does not call t.Parallel() as its first statement", fn.Name.Name)
		}

		ast.Inspect(fn.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || !isSubtestCall(call) || len(call.Args) < 2 {
				return true
			}
			subtest, ok := call.Args[1].(*ast.FuncLit)
			if !ok {
				return true
			}
			if usesNonParallelSafeAPI(subtest.Body) {
				return true
			}
			if !opensWithParallel(subtest.Body) {
				pass.Reportf(subtest.Pos(), "subtest does not call t.Parallel() as its first statement")
			}
			return true
		})
	})
	return nil, nil
}

// usesNonParallelSafeAPI reports whether body, or anything nested inside it,
// calls t.Setenv or t.Chdir — the testing package's own documented
// exceptions to running in parallel.
func usesNonParallelSafeAPI(body *ast.BlockStmt) bool {
	if body == nil {
		return false
	}
	found := false
	ast.Inspect(body, func(node ast.Node) bool {
		if found {
			return false
		}
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return true
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok || !testingReceivers[ident.Name] {
			return true
		}
		if sel.Sel.Name == "Setenv" || sel.Sel.Name == "Chdir" {
			found = true
			return false
		}
		return true
	})
	return found
}

// checkCaseName reports the ways a case name stops being a sentence about the subject.
func checkCaseName(pass *analysis.Pass, position token.Pos, name string) {
	if name == "" {
		return
	}

	if name[0] >= 'A' && name[0] <= 'Z' {
		pass.Reportf(position, "case name %q starts uppercase; case names are lowercase sentences", name)
	}
	if strings.HasSuffix(name, ".") {
		pass.Reportf(position, "case name %q ends with a period; case names take no trailing punctuation", name)
	}
	if strings.Contains(name, "_") {
		pass.Reportf(position, "case name %q uses underscores; case names are human readable, spaces only", name)
		return
	}
	if !strings.Contains(strings.TrimSpace(name), " ") {
		pass.Reportf(position, "case name %q is a single word, so it names a fixture or an outcome rather than a behavior; say what the subject does", name)
		return
	}

	lowered := strings.ToLower(strings.TrimSpace(name))
	if bannedNames[lowered] {
		pass.Reportf(position, "case name %q names an outcome class; the happy path states its result like every other case", name)
		return
	}
	if opener := strings.Fields(lowered)[0]; bannedOpeners[opener] {
		pass.Reportf(position, "case name %q opens with %q; open with what the subject does, then the condition after \"when\"", name, opener)
	}
}

// subjects are the exported names a test function may claim to cover.
type subjects struct {
	functions map[string]bool
	methods   map[string]bool
}

// packageSubjects collects the exported functions and Type_Method pairs a test
// may name, or nil when the package has no source files to draw them from.
func packageSubjects(pass *analysis.Pass) *subjects {
	found := &subjects{
		functions: make(map[string]bool),
		methods:   make(map[string]bool),
	}
	sawSource := false

	for _, file := range pass.Files {
		if strings.HasSuffix(pass.Fset.Position(file.Pos()).Filename, "_test.go") {
			continue
		}
		sawSource = true

		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}
			if fn.Recv == nil {
				if isExported(fn.Name.Name) {
					found.functions[fn.Name.Name] = true
				}
				continue
			}
			if receiver := receiverTypeName(fn.Recv.List[0].Type); receiver != "" {
				found.methods[receiver+"_"+fn.Name.Name] = true
			}
		}
	}
	if !sawSource {
		return nil
	}
	return found
}

// forEachTestFunction calls visit for every Test function in the package's test files.
func forEachTestFunction(pass *analysis.Pass, visit func(fn *ast.FuncDecl)) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{(*ast.FuncDecl)(nil)}, func(node ast.Node) {
		fn := node.(*ast.FuncDecl)
		if fn.Body == nil || fn.Recv != nil || !isTestFile(pass, fn.Pos()) {
			return
		}
		// TestMain is the test binary's entry point, not a test of anything.
		if !strings.HasPrefix(fn.Name.Name, "Test") || fn.Name.Name == "TestMain" || !takesTestingT(fn) {
			return
		}
		visit(fn)
	})
}

// tableCase is the name a table row gives itself, and where it says it.
type tableCase struct {
	position token.Pos
	text     string
}

// caseName returns the name a table row gives itself, or nil when it gives none.
func caseName(row *ast.CompositeLit) *tableCase {
	for _, element := range row.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := pair.Key.(*ast.Ident)
		if !ok || key.Name != "name" {
			continue
		}
		literal, ok := pair.Value.(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			continue
		}
		unquoted, err := strconv.Unquote(literal.Value)
		if err != nil {
			continue
		}
		return &tableCase{position: literal.Pos(), text: unquoted}
	}
	return nil
}

// opensWithParallel reports whether body's first statement is a t.Parallel() call.
func opensWithParallel(body *ast.BlockStmt) bool {
	if body == nil || len(body.List) == 0 {
		return false
	}
	statement, ok := body.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := statement.X.(*ast.CallExpr)
	return ok && isMethodCall(call, "Parallel")
}

// isSubtestCall reports whether call opens a subtest. The receiver has to be
// the test's own value: analysistest.Run is a harness entry point, not a subtest.
func isSubtestCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Run" {
		return false
	}
	receiver, ok := selector.X.(*ast.Ident)
	return ok && testingReceivers[receiver.Name]
}

// isMethodCall reports whether call invokes the named method on some receiver.
func isMethodCall(call *ast.CallExpr, method string) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == method
}

// takesTestingT reports whether fn has the signature of a test.
func takesTestingT(fn *ast.FuncDecl) bool {
	if fn.Type.Params == nil || len(fn.Type.Params.List) != 1 {
		return false
	}
	pointer, ok := fn.Type.Params.List[0].Type.(*ast.StarExpr)
	if !ok {
		return false
	}
	selector, ok := pointer.X.(*ast.SelectorExpr)
	return ok && selector.Sel.Name == "T"
}

// receiverTypeName returns the bare type name a method hangs off.
func receiverTypeName(expr ast.Expr) string {
	switch receiver := expr.(type) {
	case *ast.StarExpr:
		return receiverTypeName(receiver.X)
	case *ast.Ident:
		return receiver.Name
	case *ast.IndexExpr:
		return receiverTypeName(receiver.X)
	case *ast.IndexListExpr:
		return receiverTypeName(receiver.X)
	}
	return ""
}

func isTestFile(pass *analysis.Pass, position token.Pos) bool {
	return strings.HasSuffix(pass.Fset.Position(position).Filename, "_test.go")
}

func isExported(name string) bool {
	return len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
}
