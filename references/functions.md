# Functions, Methods & Design

Function and method design determines clarity and reusability. Methods belong with their types. Interfaces belong to consumers. Constructors own initialization logic.

## Maximum 4 parameters

Functions with more than 4 parameters should use a config struct or functional options pattern.

```go
// Do this
type Config struct {
  Timeout time.Duration
  Retries int
  Debug   bool
}
func Process(ctx context.Context, req *Request, cfg Config) error

// Not this
func Process(ctx context.Context, req *Request, timeout time.Duration, retries int, debug bool) error
```

### When to group

Collapse parameters into a struct when the signature crosses four, when two of
the same type sit next to each other — callers will swap them — or when the same
set travels together through several calls.

| Type | Holds | Lives |
|---|---|---|
| `Config` | construction-time settings for a `New*` constructor | beside the type it configures |
| `Options` | optional tuning of one call; the zero value is the default | above the function that takes it |
| `Request` | the required inputs of one operation | above the function that takes it |

- **Named from the package's vocabulary, never its name.** In package `user`, `CreateRequest` — not `UserCreateRequest`.
- **Declared in the package that owns the function, immediately above it.** A parameter type is part of one signature, not a shared vocabulary; it moves to its own file only when several functions take it.
- **`Config` and `Options` must be usable as `Config{}`.** A `Request` whose zero value is meaningless is validated by the function, not by the caller.

An unexported function may simply take the parameters. The grouping exists to
keep an exported signature readable at the call site.


## `context.Context` always first

If a function accepts a context, it must be the first parameter after the receiver.

```go
// Do this
func (store *Store) Get(ctx context.Context, id string) (*User, error)
func Fetch(ctx context.Context, url string) ([]byte, error)

// Not this
func (store *Store) Get(id string, ctx context.Context) (*User, error)
func Fetch(url string, ctx context.Context) ([]byte, error)
```


## `*slog.Logger` always second

If a function accepts a logger, it comes immediately after context, named `log` or `logger`.

```go
// Do this
func Process(ctx context.Context, log *slog.Logger, req *Request) error

// Not this
func Process(ctx context.Context, req *Request, log *slog.Logger) error
```


## No boolean parameters

Boolean parameters mean the function does two things. Split it into two functions or use a typed option instead.

```go
// Do this
func Save(ctx context.Context, req *Request) error
func SaveDraft(ctx context.Context, req *Request) error

// Also acceptable - typed option
type SaveOption uint8
const (
  SaveFinal SaveOption = iota
  SaveDraft
)
func Save(ctx context.Context, req *Request, opt SaveOption) error

// Not this
func Save(ctx context.Context, req *Request, draft bool) error
```


## No named return values

Named return values obscure control flow and invite bare `return` statements.

```go
// Do this
func Parse(data []byte) (*Config, error)

// Not this
func Parse(data []byte) (cfg *Config, err error)
```


## No `func` parameters

Never pass a callback function as a parameter. Define an interface instead.

```go
// Do this
type Handler interface {
  Handle(ctx context.Context, req *Request) error
}
func Process(ctx context.Context, handler Handler) error

// Exception: stdlib callbacks are acceptable
sort.Slice(items, func(i, j int) bool { return items[i] < items[j] })

// Not this
func Process(ctx context.Context, fn func(*Request) error) error
```


## Return concrete types

Always return concrete types from functions and constructors. The only exception is `error`.

```go
// Do this
type Store interface {
  Get(ctx context.Context, key string) (string, error)
}
func NewStore() *MemoryStore { return &MemoryStore{} }

// Not this
func NewStore() Store { return &MemoryStore{} }
```


## Absence has one spelling, chosen by the type

Every signature that can come back empty-handed must say so, and the type decides how — not the author, not the call site, not the mood of the package.

- **Records return a pointer.** Structs, and anything else with identity, signal absence with `nil`.
- **Values return `(T, bool)`.** Primitives, and struct types with value semantics like `time.Time`, use the comma-ok idiom — the same shape as `v, ok := m[key]`. The bool is last and is read as `ok` at the call site.
- **Collections have no absent state.** A nil slice and an empty slice are the same thing to every caller that ranges over one. Never `([]T, bool)`, never `*[]T`, never `(map[K]V, bool)`.

```go
// Do this
func (store *Store) User(ctx context.Context, id string) (*User, error)      // record: nil is no user
func (cache *Cache) TTL(key string) (time.Duration, bool)                    // value: ok is presence
func (store *Store) Users(ctx context.Context) ([]User, error)               // collection: empty is empty

// Not this
func (store *Store) User(ctx context.Context, id string) (User, bool, error) // record spelled as a value
func (cache *Cache) TTL(key string) *time.Duration                           // value spelled as a record
func (store *Store) Users(ctx context.Context) ([]User, bool, error)         // collections are never absent
```


### Never both spellings

`(*T, bool)` asks the caller to check twice and leaves `(nil, true)` undefined. The type has already picked the spelling — use it once.

```go
// Do this
func (store *Store) User(ctx context.Context, id string) (*User, error)

// Not this
func (store *Store) User(ctx context.Context, id string) (*User, bool, error)
```


### Never a sentinel value

`-1`, `""`, `0`, and `time.Time{}` are values, not signals. A caller that has to know `-1` means missing was owed a `bool`.

```go
// Do this
func (index *Index) Offset(key string) (int, bool)

// Not this
func (index *Index) Offset(key string) int // returns -1 when absent
```


### `ok` is not an error

A lookup that can legitimately miss returns `nil` or `ok`, and no error. A miss the caller cannot proceed past is an error — a sentinel from `errors.go`, and no bool. Since `error` is always the last return and `ok` is always the last return, no signature can carry both: `(T, bool, error)` offers two ways to fail and no rule for reading them.

```go
// Do this - a cache miss is ordinary
func (cache *Cache) Token(key string) (string, bool)

// Do this - a missing user is a failure the caller must handle
func (store *Store) User(ctx context.Context, id string) (*User, error) // errors.Is(err, ErrUserNotFound)

// Not this
func (store *Store) User(ctx context.Context, id string) (*User, bool, error)
```


### The returned `ok` is the only bool a signature may carry

Boolean *parameters* stay banned — they mean the function does two things. A returned `ok` switches nothing: it is the presence half of a value, and it always comes last.


## Function length: 60 lines maximum

A function approaching 60 lines is a signal it's doing too much. Split it.

```go
// Do this - separate concerns
func Validate(req *Request) error {
  // 15 lines
}

func Save(ctx context.Context, req *Request) error {
  // 20 lines
}

func Process(ctx context.Context, req *Request) error {
  // 12 lines
}

// Not this
func HandleRequest(ctx context.Context, req *Request) error {
  // 80 lines doing validation, saving, processing
}
```


## Methods on the type

Methods belong in the same file as the type definition they operate on.

```go
// store.go
type Store struct { }
func (store *Store) Get(ctx context.Context, id string) (*User, error)
func (store *Store) Save(ctx context.Context, user *User) error

// Not this - methods in separate file
// store.go
type Store struct { }

// store_methods.go
func (store *Store) Get(ctx context.Context, id string) (*User, error)
```


## Interfaces belong to the consumer

The package that uses an interface defines it, not the package that implements it. The service layer defines what it needs; the data layer satisfies it.

```go
// Do this - in service package
type UserRepository interface {
  GetByID(ctx context.Context, id string) (*User, error)
  Save(ctx context.Context, user *User) error
}

type UserService struct {
  repo UserRepository
}

// Do this - in data package
type PostgresStore struct { }
func (p *PostgresStore) GetByID(ctx context.Context, id string) (*User, error)
func (p *PostgresStore) Save(ctx context.Context, user *User) error

// Not this - data package defines the interface
// data/repository.go
type Repository interface {
  GetByID(ctx context.Context, id string) (*User, error)
}
```


## Never define an unused interface

Interfaces are contracts. Speculative interfaces should not exist.

```go
// Do this - only define if this package uses it
type Reader interface {
  Read([]byte) (int, error)
}

func ProcessInput(r Reader) {}

// Not this - interface exists but is unused in this package
type Reader interface {
  Read([]byte) (int, error)
}

type Writer interface {
  Write([]byte) (int, error)
}
// Neither Reader nor Writer is used here
```


## Keep interfaces small

The bigger the interface, the weaker the abstraction. Prefer many small, focused interfaces. Limit interfaces to 5 or fewer methods.

```go
// Do this
type Reader interface {
  Read([]byte) (int, error)
}

type Writer interface {
  Write([]byte) (int, error)
}

// Not this
type ReadWriter interface {
  Read([]byte) (int, error)
  Write([]byte) (int, error)
  Close() error
  Seek(int64, int) (int64, error)
  Stat() (os.FileInfo, error)
}
```


## No `any` in exported APIs

Use concrete types or well-defined interfaces instead of `any`/`interface{}`.

```go
// Do this
func Process(ctx context.Context, data *json.RawMessage) error

// Not this
func Process(ctx context.Context, data any) error
```


## No channels or sync.WaitGroup in exported signatures

Wrap coordination primitives behind a concrete type or interface.

```go
// Do this
type Worker interface {
  Work(ctx context.Context) error
}

func StartWorker(ctx context.Context, worker Worker) error

// Not this
func StartWorker(ctx context.Context, results chan string, wg *sync.WaitGroup)
```


## Constructors

Use constructors to initialize types with unexported fields or complex initialization.

```go
// Do this
type Config struct {
  timeout time.Duration
}

func NewConfig(timeout time.Duration) *Config {
  if timeout <= 0 {
    timeout = 30 * time.Second
  }
  return &Config{timeout: timeout}
}

// Not this
type Config struct {
  Timeout time.Duration
}
// No constructor; caller must set values
```

Constructor naming: `New<TypeName>` for single return, `New<TypeName>s` for collections.

```go
func NewUser(name string) *User
func NewUsers(names []string) []*User
func NewUserStore() *Store
```
