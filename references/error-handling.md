# Error Handling

All errors must be wrapped with error chains, never returned naked. Sentinel errors live at package scope. Never match on error messages—use the type system.


## Always wrap with `%w`

Use `fmt.Errorf` with the `%w` verb to wrap errors and preserve the chain.

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

The presentation layer maps service errors to the transport. Turning
`task.ErrNotFound` into a 404 belongs to the API layer alone — see the HTTP
section in `references/architecture.md`. Neither the service nor the data layer
knows about status codes.


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
