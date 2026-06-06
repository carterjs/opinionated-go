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

The bigger the interface, the weaker the abstraction. Prefer many small, focused interfaces.

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
