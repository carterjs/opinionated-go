# Code Examples

Detailed code examples demonstrating each opinionated-go convention. Referenced by `SKILL.md`.

## Naming

### Full words only

```go
// Do this
type Document struct {}
type Request struct {}

// Not this
type Doc struct {}
type Req struct {}
```

### Initialisms uppercase

```go
// Do this
func GetUserID() string {}
func FetchURL() string {}

// Not this
func GetUserId() string {}
func FetchUrl() string {}
```

### Variable length scales with scope

```go
// Do this
for i := 0; i < 10; i++ { }
partitionKey := "user:123"
errorCount := 0

// Not this
pk := "user:123"
ec := 0
```

### ctx and err naming

```go
// Do this
func Process(ctx context.Context) error {}
if err := doSomething(); err != nil {}

// Not this
func Process(c context.Context) error {}
if e := doSomething(); e != nil {}
```

### Receiver names

```go
// Do this
func (store *UserStore) Get(ctx context.Context, id string) (*User, error) {}

// Not this
func (s *UserStore) Get(ctx context.Context, id string) (*User, error) {}
```

### Package names

```
// Do this
package user
directory: user/

// Not this
package user_service
package UserService
```

### File names

```
// Do this
store.go
client.go
store_test.go

// Not this
user_helpers.go
utils.go
```

---

## Comments

### Exported identifiers

```go
// Do this
// Document represents a stored document.
type Document struct {}

// Not this
// A struct for documents
type Document struct {}
```

### Inline comments

```go
// Do this
// Retry with exponential backoff; sync.WaitGroup alone doesn't guarantee order
time.Sleep(time.Duration(math.Pow(2, float64(attempt))) * time.Second)

// Not this
// increment the counter
count++
```

---

## Error Handling

### Always wrap with %w

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

### Sentinel errors at package level

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

### Use errors.Is / errors.As

```go
// Do this
if errors.Is(err, ErrNotFound) {}

// Not this
if strings.Contains(err.Error(), "not found") {}
```

### Errors last return value

```go
// Do this
func Process(ctx context.Context) (string, error)

// Not this
func Process(ctx context.Context) (error, string)
```

### Error strings lowercase, no punctuation

```go
// Do this
return fmt.Errorf("reading config: %w", err)

// Not this
return fmt.Errorf("Reading config: %w.", err)
```

### Indent error flow

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

### No panic in library code

```go
// Do this (library)
func Parse(data []byte) (*Config, error) {
  if len(data) == 0 {
    return nil, errors.New("empty data")
  }
}

// Not this (library)
func Parse(data []byte) *Config {
  if len(data) == 0 {
    panic("empty data")
  }
}
```

---

## Function & Method Design

### Maximum 4 parameters

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

### context.Context first

```go
// Do this
func Fetch(ctx context.Context, id string) (*User, error)

// Not this
func Fetch(id string, ctx context.Context) (*User, error)
```

### *slog.Logger second

```go
// Do this
func Process(ctx context.Context, log *slog.Logger, req *Request) error

// Not this
func Process(ctx context.Context, req *Request, log *slog.Logger) error
```

### No boolean parameters

```go
// Do this
func Save(ctx context.Context, req *Request) error
func SaveDraft(ctx context.Context, req *Request) error

// Not this
func Save(ctx context.Context, req *Request, draft bool) error
```

### No named return values

```go
// Do this
func Parse(data []byte) (*Config, error)

// Not this
func Parse(data []byte) (cfg *Config, err error)
```

### No func parameters

```go
// Do this
type Handler interface {
  Handle(ctx context.Context, req *Request) error
}
func Process(ctx context.Context, handler Handler) error

// Not this
func Process(ctx context.Context, fn func(*Request) error) error
```

### Return concrete types

```go
// Do this
type Store interface { Get(ctx context.Context, key string) (string, error) }
func NewStore() *MemoryStore

// Not this
func NewStore() Store
```

### Function length 60 lines maximum

```go
// Do this - separate concerns
func Validate(req *Request) error { /* 15 lines */ }
func Save(ctx context.Context, req *Request) error { /* 20 lines */ }
func Process(ctx context.Context, req *Request) error { /* 12 lines */ }

// Not this
func HandleRequest(ctx context.Context, req *Request) error { /* 80 lines */ }
```

---

## Interfaces

### Interfaces belong to consumer

```go
// Do this - in service package
type UserRepository interface {
  GetByID(ctx context.Context, id string) (*User, error)
}

// Not this - in data package
// data/repository.go
type Repository interface {
  GetByID(ctx context.Context, id string) (*User, error)
}
```

### Never define unused interface

```go
// Do this
// Only define Writer if you use it in this package
type Writer interface { Write([]byte) (int, error) }
func ProcessOutput(w Writer) {}

// Not this
// Don't define interfaces you might need someday
type Reader interface { Read([]byte) (int, error) }  // if unused
```

### Keep interfaces small

```go
// Do this
type Reader interface {
  Read([]byte) (int, error)
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

### No `any` in public APIs

```go
// Do this
func Process(ctx context.Context, data *json.RawMessage) error

// Not this
func Process(ctx context.Context, data any) error
```

### No channels in exported signatures

```go
// Do this
type Worker interface {
  Work(ctx context.Context) error
}

// Not this
func StartWorker(ctx context.Context, results chan string, wg *sync.WaitGroup)
```

---

## Global State & Configuration

### os.Getenv only in main

```go
// Do this - in main
timeout := time.Duration(mustAtoi(os.Getenv("TIMEOUT_SECS"))) * time.Second

// Not this - in service package
func (svc *Service) GetTimeout() time.Duration {
  return time.Duration(mustAtoi(os.Getenv("TIMEOUT_SECS"))) * time.Second
}
```

### No global slog functions

```go
// Do this
func Process(ctx context.Context, log *slog.Logger) error {
  log.InfoContext(ctx, "processing")
}

// Not this
func Process(ctx context.Context) error {
  slog.Info("processing")
}
```

### No init() functions

```go
// Do this
func NewService(cfg Config) *Service {
  return &Service{config: cfg}
}

// Not this
var globalService *Service
func init() {
  globalService = &Service{}
}
```

### No errgroup

```go
// Do this
var wg sync.WaitGroup
ctx, cancel := context.WithCancelCause(context.Background())
defer cancel(nil)

wg.Add(1)
go func() {
  defer wg.Done()
  if err := doWork(ctx); err != nil {
    cancel(err)
  }
}()
wg.Wait()

// Not this
g, ctx := errgroup.WithContext(context.Background())
g.Go(func() error { return doWork(ctx) })
```

### Dependency injection via constructor

```go
// Do this
func NewService(db *Database, log *slog.Logger) *Service {
  return &Service{db: db, log: log}
}

// Not this
var globalDB *Database
func NewService() *Service {
  return &Service{
    process: func() error {
      return globalDB.Query()  // captured global state
    },
  }
}
```

---

## Structs & Types

### No exported fields with methods

```go
// Do this
type User struct {
  id   string
  name string
}
func (u *User) ID() string { return u.id }
func (u *User) SetName(n string) { u.name = n }

// Not this
type User struct {
  ID   string
  Name string
}
func (u *User) Validate() error {}
```

### Constructors for unexported fields

```go
// Do this
type Config struct {
  timeout time.Duration
}
func NewConfig(timeout time.Duration) *Config {
  return &Config{timeout: timeout}
}

// Not this
type Config struct {
  timeout time.Duration
}
// no constructor provided
```

### Zero value must be valid

```go
// Do this
type Duration struct {
  seconds int64
}
var d Duration  // zero value is valid

// Not this
type Duration struct {
  seconds int64
}
// zero value is invalid without constructor
```

### Config structs with <= 0 checks

```go
// Do this
type Config struct {
  Timeout time.Duration
}
func (cfg Config) GetTimeout() time.Duration {
  if cfg.Timeout <= 0 {
    return 30 * time.Second
  }
  return cfg.Timeout
}

// Not this
caller := NewService(Config{
  Timeout: 30 * time.Second,
})
```

### Typed constants

```go
// Do this
type Status string
const (
  StatusActive Status = "active"
  StatusPending Status = "pending"
)

// Not this
const (
  StatusActive = "active"
  StatusPending = "pending"
)
```

### No magic numbers

```go
// Do this
const MaxRetries = 3
const ContextTimeoutSeconds = 30

for i := 0; i < MaxRetries; i++ {}
ctx, cancel := context.WithTimeout(ctx, ContextTimeoutSeconds*time.Second)

// Not this
for i := 0; i < 3; i++ {}
ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
```

---

## Concurrency

### Synchronous by default

```go
// Do this - caller decides
func Fetch(ctx context.Context, url string) ([]byte, error)

// caller controls concurrency
results := make(chan []byte)
for url := range urls {
  go func(u string) {
    results <- Fetch(ctx, u)
  }(url)
}

// Not this - hides goroutines
func FetchAll(urls []string) [][]byte {
  results := make(chan []byte)
  for url := range urls {
    go func(u string) {
      results <- fetch(u)
    }(url)
  }
  // ...
}
```

### No fire-and-forget goroutines

```go
// Do this - tracked with WaitGroup
var wg sync.WaitGroup
wg.Add(1)
go func() {
  defer wg.Done()
  doWork()
}()
wg.Wait()

// Not this - forgotten
go doWork()
```

### context.WithCancelCause

```go
// Do this
ctx, cancel := context.WithCancelCause(context.Background())
if err := work(ctx); err != nil {
  cancel(err)
}

// Not this
ctx, cancel := context.WithCancel(context.Background())
if err := work(ctx); err != nil {
  cancel()  // loses error information
}
```

---

## Package & File Organization

### Dependencies flow downward

```
// Do this
presentation/ (imports service/)
  handler.go
service/     (imports data/)
  logic.go
data/        (imports nothing internal)
  store.go

// Not this
data/ imports service/
```

### Presentation under internal/

```
// Do this
internal/http/
  handler.go
internal/cli/
  cmd.go
service/
  logic.go

// Not this
http/
  handler.go  // accidentally importable
```

### One purpose per package

```
// Do this
user/
product/
order/

// Not this
domain/     // vague purpose
common/     // too generic
```

### File organization

```go
// Do this - order in file
type User struct { }      // exported type
func (u *User) Save() {}  // methods on User
func NewUser() *User {}   // constructors

func helper() {}          // helpers below

// Not this
func helper() {}
type User struct { }
```

### Delete dead code

```go
// Do this
func (u *User) Save(ctx context.Context) error { }

// Not this
func (u *User) Save(ctx context.Context) error { }
func (u *User) Save2(ctx context.Context) error { }  // unused variant
```

---

## Layered Architecture

### Service owns interfaces

```go
// Do this - service defines what it needs
type Repository interface {
  GetByID(ctx context.Context, id string) (*User, error)
}
type UserService struct {
  repo Repository
}

// Not this - data package owns the interface
// data/user.go
type UserRepository interface { }
```

### Data satisfies service interfaces

```go
// Do this
// data/user_store.go satisfies service.Repository
type PostgresStore struct { }
func (p *PostgresStore) GetByID(ctx context.Context, id string) (*User, error)

// Not this
// data/ defines what service needs
```

### Presentation imports service only

```go
// Do this - presentation/handler.go
import "myapp/service"

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
  user, err := h.svc.GetUser(r.Context(), id)
}

// Not this
import "myapp/data"

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
  user, err := h.store.GetByID(r.Context(), id)  // direct data access
}
```

---

## Testing

See `references/testing.md` for complete testing conventions and examples.
