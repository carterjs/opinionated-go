# Testing

Tests are contracts that serve as documentation. Apply the same conventions as production code.

## Structure

- **One test function per subject.** How many `Test*` functions a package has is decided by its API surface, not by how many behaviors you are covering: one per exported function, one per exported method, nothing else. A new behavior is a new row in the table that already exists.
- **Add cases, not functions.** A second function for a subject that already has one — `TestParseDocument_EdgeCases` beside `TestParseDocument`, `TestStore_Get_Errors` beside `TestStore_Get` — is a violation, not a style choice. Splitting by outcome (`_Success`, `_Failure`, `_Invalid`) is the same violation wearing a suffix. If a case will not fit the existing table, the table is wrong — widen it.
- **Table-driven always.** Every test function must use a table: `tests := []struct{ name string; ... }{}` followed by `for _, test := range tests`. Any test function that is not table-driven is a violation — warn.
- **Loop variable always named `test`.** Never `tt`, `tc`, `c`, or any other shorthand. Never reassigned inside the loop body (no `test := test` — Go 1.22+ scopes loop variables per iteration).
- **No branching inside the loop body.** Each case must be fully self-contained. If branching is required, the table is not the right structure.
- **Always parallel.** Every test function calls `t.Parallel()` immediately. Every subtest calls `t.Parallel()` immediately inside `t.Run`.

## Naming

### Test functions

- **`Test<FunctionName>`** for functions: `TestParseDocument`.
- **`Test<TypeName>_<MethodName>`** for methods: `TestStore_Get`.
- No other naming forms are acceptable — error.
- No test function may cover unexported functionality directly — error. Test behavior through the public API.
- **A `*_test.go` file with no corresponding `*.go` source file is worth a second look** — warn, not error. A `<concept>_integration_test.go` file, or a file that deliberately tests one behavior across several source files rather than mirroring a single one (`document_rename_test.go`, `conformance_users_test.go`), is a legitimate convention; the warning exists to catch a genuinely orphaned leftover, not to force every test file into a 1:1 mapping. `export_test.go` is the one hard exception: it's banned outright, error — do not expose unexported identifiers to external test packages.

### Test cases

A case's `name` is the sentence a failure prints. Written well, the reader knows what broke without opening the table. One grammar per table, one grammar per package.

- **Behavior first, condition second.** `"returns error when input is empty"`. The opening verb phrase says what the subject does — never what the test does.
- **Present tense, third person, no subject.** `"returns"`, `"rejects"`, `"trims trailing slash"`. Never `"should return"`, `"will return"`, `"it returns"`, `"test returns empty"`.
- **Lowercase start, no trailing period. Spaces only** — `"returns error when input is empty"`, not `"returns_error_when_input_is_empty"` or `"returnsErrorWhenInputIsEmpty"`. No snake_case, no camelCase.
- **`when` is the only condition word.** Not `"if"`, `"with"`, `"given"`, or a bare comma.
- **Never name a case after its input.** `"empty string"`, `"nil store"`, and `"case 3"` name a fixture, not a contract; the input already lives in the case's fields. `"returns error when store is nil"` is the name.
- **Never name a case after its outcome class.** No `"success"`, `"failure"`, `"happy path"`, `"error case"`, `"valid"`, `"invalid"`. The happy path states its result like every other case: `"returns the parsed document"`.
- **Names are unique within a table.** Two cases reaching for the same name are one case, or the name is not yet specific enough to tell them apart.

```go
tests := []struct {
    name     string
    input    string
    expected *Document
}{
    {name: "returns the parsed document", ...},
    {name: "returns error when input is empty", ...},
    {name: "trims trailing whitespace when the body is padded", ...},
}
```

## File Organization

- Test functions appear at the top of the file.
- Mocks and helpers follow below, after all test functions.
- Mock files live in their own file named `mock_<concept>.go` within the package, with symbols sorted alphabetically.

## Output & Context

- **`t.Log` / `t.Logf`** for debug output. Never `fmt.Print*` inside tests.
- **`t.Attr`** for structured key-value metadata associated with the test run — prefer over ad-hoc `t.Log` calls when recording structured information for CI or tooling.
- **`t.Output()`** when a test requires an `io.Writer`. Never pass `os.Stdout` or `os.Stderr` to code under test.
- **`t.Context()`** always. Never `context.Background()` or `context.TODO()` inside a `_test.go` file. Outside tests the rule is different: a root context belongs to `main`, and everything else accepts one from its caller.

## Assertions

- Use `github.com/stretchr/testify/assert` and `require`.
- `require` for preconditions and setup — failure stops the test immediately.
- `assert` for the actual assertions — failure is recorded but the test continues.

## Mocks

- Hand-written functional mocks only. No generated mocks (`mockery`, `gomock`, etc.).
- Mirror the interface name with a `Mock` prefix: `MockStore` for a `Store` interface.
- Each method has a corresponding `<Method>Func` field of the matching function type.
- The method implementation calls the func field directly. A nil field panics — this surfaces unexpected calls immediately rather than silently passing.
- Each test case sets only the func fields it expects to be called.
- Mocks live in their own file: `mock_<concept>.go`.

```go
type MockStore struct {
    GetFunc    func(ctx context.Context, key string) (string, error)
    DeleteFunc func(ctx context.Context, key string) error
}

func (mock *MockStore) Get(ctx context.Context, key string) (string, error) {
    return mock.GetFunc(ctx, key)
}

func (mock *MockStore) Delete(ctx context.Context, key string) error {
    return mock.DeleteFunc(ctx, key)
}
```

## Unit vs Integration

- **Unit tests** isolate a single component by mocking its direct dependencies. The unit exercises only its own logic.
- **Integration tests** verify behavior across multiple real layers. Mock one level deeper than the boundary under test.
- Integration test files named `<concept>_integration_test.go` — the filename names the boundary being tested, not a single source file.

## General

- If a behavior is already exercised as a side-effect of an existing case, do not add a case for it either.
- Add a new `Test*` function only when the subject is a different function or method.
- Minimal setup outside of test functions. Heavy setup belongs in table case fields or helpers called within the loop.
- Use `testify/assert` — never `t.Fatal` or `t.Error` directly for assertions.
