# Testing

Tests are contracts. Readable enough to serve as documentation of intended behavior. No special treatment — the same conventions that apply to production code apply here.

-----

## Structure

- **Table-driven always.** Every test function must use a table: `tests := []struct{ name string; ... }{}` followed by `for _, test := range tests`. Any test function that is not table-driven is a violation — warn.
- **Loop variable always named `test`.** Never `tt`, `tc`, `c`, or any other shorthand. Never reassigned inside the loop body (no `test := test` — Go 1.22+ scopes loop variables per iteration).
- **No branching inside the loop body.** Each case must be fully self-contained. If branching is required, the table is not the right structure.
- **Always parallel.** Every test function calls `t.Parallel()` immediately. Every subtest calls `t.Parallel()` immediately inside `t.Run`.
- **`t.Run` subtest names are human readable.** Spaces only — `"returns error when input is empty"` not `"returns_error_when_input_is_empty"` or `"returnsErrorWhenInputIsEmpty"`. No snake_case, no camelCase.

-----

## Naming

- **`Test<FunctionName>`** for functions: `TestParseDocument`.
- **`Test<TypeName>_<MethodName>`** for methods: `TestStore_Get`.
- No other naming forms are acceptable — error.
- No test function may cover unexported functionality directly — error. Test behavior through the public API.
- **No `*_test.go` file without a corresponding `*.go` source file** — error. Exception: `export_test.go` is banned. Do not expose unexported identifiers to external test packages.

-----

## File Organization

- Test functions appear at the top of the file.
- Mocks and helpers follow below, after all test functions.
- Mock files live in their own file named `mock_<concept>.go` within the package, with symbols sorted alphabetically.

-----

## Output & Context

- **`t.Log` / `t.Logf`** for debug output. Never `fmt.Print*` inside tests.
- **`t.Attr`** for structured key-value metadata associated with the test run — prefer over ad-hoc `t.Log` calls when recording structured information for CI or tooling.
- **`t.Output()`** when a test requires an `io.Writer`. Never pass `os.Stdout` or `os.Stderr` to code under test.
- **`t.Context()`** always. Never `context.Background()` or `context.TODO()` inside a test function.

-----

## Assertions

- Use `github.com/stretchr/testify/assert` and `require`.
- `require` for preconditions and setup — failure stops the test immediately.
- `assert` for the actual assertions — failure is recorded but the test continues.

-----

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

-----

## Unit vs Integration

- **Unit tests** isolate a single component by mocking its direct dependencies. The unit exercises only its own logic.
- **Integration tests** verify behavior across multiple real layers. Mock one level deeper than the boundary under test.
- Integration test files named `<concept>_integration_test.go` — the filename names the boundary being tested, not a single source file.

-----

## General

- Prefer extending an existing table over adding a new test function.
- If a behavior is already exercised as a side-effect of an existing case, do not add a dedicated test for it.
- Add a new `Test*` function only when the subject truly cannot fit an existing table — a different function or method is under test.
- Minimal setup outside of test functions. Heavy setup belongs in table case fields or helpers called within the loop.
- Use `testify/assert` — never `t.Fatal` or `t.Error` directly for assertions.
