# Error Handling

Errors crossing a package boundary must be wrapped with error chains, never returned naked. Sentinel errors live at package scope. Never match on error messages—use the type system.


## Always wrap with `%w` at the boundary

Use `fmt.Errorf` with the `%w` verb to wrap errors and preserve the chain.

Wrapping earns its keep where the error leaves the package: an **exported**
function is the last place that knows what the operation was, so it says so. An
unexported helper's caller is a few lines away and adds that context itself —
requiring `%w` there is ceremony, and the analyzer does not ask for it.

```go
// Do this
if err := doSomething(); err != nil {
  return fmt.Errorf("processing: %w", err)
}

// Not this
if err := doSomething(); err != nil {
  return err
}
```


## Sentinel errors at package level only

Define sentinel errors as package-level `var` statements. Never create them inline.

```go
// Do this
var ErrNotFound = errors.New("not found")

func Get(key string) (string, error) {
  if !exists(key) {
    return "", ErrNotFound
  }
}

// Not this
func Get(key string) (string, error) {
  if !exists(key) {
    return "", errors.New("not found")
  }
}
```


## Typed error structs

When callers need to inspect error details beyond a sentinel check, use a typed error struct.

```go
type ValidationError struct {
  Field   string
  Message string
}

func (e ValidationError) Error() string {
  return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Caller can then use errors.As
if err := validate(req); err != nil {
  var valErr ValidationError
  if errors.As(err, &valErr) {
    log.Printf("validation failed on %s", valErr.Field)
  }
}
```


## Use `errors.Is` and `errors.As` only

Never string-match on error messages.

```go
// Do this
if errors.Is(err, ErrNotFound) {
  return nil
}

var valErr ValidationError
if errors.As(err, &valErr) {
  log.Printf("validation failed: %s", valErr.Field)
}

// Not this
if strings.Contains(err.Error(), "not found") {
  return nil
}
```


## Errors always last return value

The error must be the final return value in any function signature.

```go
// Do this
func Process(ctx context.Context, req *Request) (string, error)

// Not this
func Process(ctx context.Context, req *Request) (error, string)
```


## Error strings: lowercase, no punctuation

Error messages should be lowercase and without trailing punctuation.

```go
// Do this
return fmt.Errorf("reading config: %w", err)

// Not this
return fmt.Errorf("Reading config: %w.", err)
```


## Indent error flow

Return early on error. Keep the happy path at the left margin.

```go
// Do this
if err := validate(req); err != nil {
  return err
}

if err := save(req); err != nil {
  return err
}

return process(req)

// Not this
if err := validate(req); err == nil {
  if err := save(req); err == nil {
    return process(req)
  }
}
```


## Translate errors at layer boundaries

The service layer owns the domain errors. Sentinels and typed errors are defined
there, next to the types they belong to.

The data-layer adapter translates downstream errors into those service errors —
that is its job. Never propagate a downstream error blindly; it leaks a detail
the caller should not know.

```go
// Do this - adapter maps the downstream error to a service error
func (adapter *TaskAdapter) Get(ctx context.Context, id string) (*task.Task, error) {
  t, err := adapter.client.query(ctx, id)
  if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
      return nil, task.ErrNotFound
    }
    return nil, fmt.Errorf("getting task: %w", err)
  }
  return t, nil
}

// Not this - the downstream error type leaks upward
func (adapter *TaskAdapter) Get(ctx context.Context, id string) (*task.Task, error) {
  return adapter.client.query(ctx, id) // sql.ErrNoRows leaks into the service
}
```

The rule holds when the downstream is another service's client: an adapter
wrapping a `payment.Client` decides what `payment.ErrDeclined` means in its own
domain, rather than passing it up unchanged.

Distinguish a dependency's *failure* from its *answer*. When a downstream is
unreachable, times out, or returns something unusable, translate it to an error
that marks an upstream failure — not a generic internal error — so the API can
report a 502 rather than a 500. A 500 means *we* are stuck; a 502 means a
dependency is.

A downstream 429 is a decision, not a reflex. Depending on the client, either
retry with backoff and honor `Retry-After`, or translate it to a rate-limited
error and let it surface as a 429 upstream. Retrying blindly can deepen the
overload; propagating blindly wastes a legitimate retry. Choose per client.

The presentation layer maps service errors to the transport. Turning
`task.ErrNotFound` into a 404 belongs to the API layer alone — see the HTTP
section in `references/architecture.md`. Neither the service nor the data layer
knows about status codes.


## Error codes: one shared vocabulary

When many errors must reach a transport, a single `errcode` package gives them
stable, serializable names. It owns two types and the mapping — and nothing
about any transport.

- **`Class`** is a coarse, transport-neutral category (`invalid`, `not_found`,
  `conflict`, …). Many codes share one class.
- **`Code`** is a stable string identifier for one condition, carrying its class
  and a description. Codes are values, not bare enum constants, so each one can
  describe itself.
- **`FromError`** is the single place that knows every service's errors. It maps
  sentinels and typed errors to codes with `errors.Is` / `errors.As`.

```go
package errcode

type Class string

const (
  ClassInvalid     Class = "invalid"
  ClassNotFound    Class = "not_found"
  ClassConflict    Class = "conflict"
  ClassRateLimited Class = "rate_limited"
  ClassUpstream    Class = "upstream" // a downstream dependency failed, not us
  ClassInternal    Class = "internal"
)

type Code string

// Each code carries a class and a description.
var (
  Unknown      = define("unknown", ClassInternal, "an unexpected error occurred")
  TaskNotFound = define("task_not_found", ClassNotFound, "the requested task does not exist")
  TaskExists   = define("task_exists", ClassConflict, "a task already exists with that identifier")
)

func FromError(err error) Code {
  switch {
  case errors.Is(err, task.ErrNotFound):
    return TaskNotFound
  case errors.Is(err, task.ErrExists):
    return TaskExists
  }
  return Unknown
}
```

Service packages still own their errors — `task.ErrNotFound` lives in `task`.
`errcode` only names and classifies them, so it imports the services, never the
reverse. Keep it transport-neutral: no status codes, no response shapes. A class
is a category, not an HTTP status.

```go
// Not this - HTTP has leaked into the shared vocabulary
func (code Code) HTTPStatus() int { // status mapping belongs to the API layer
  return 404
}
```


## No `panic` in library code

Library code must return errors. `panic` is only acceptable in `main` or test setup helpers.

```go
// Do this (library)
func Parse(data []byte) (*Config, error) {
  if len(data) == 0 {
    return nil, errors.New("empty data")
  }
}

// Do this (main)
func main() {
  cfg, err := loadConfig()
  if err != nil {
    panic(fmt.Sprintf("failed to load config: %v", err))
  }
}

// Not this (library)
func Parse(data []byte) *Config {
  if len(data) == 0 {
    panic("empty data")
  }
}
```
