# Layered Architecture

Three layers, dependencies flowing one direction only.

```
Presentation  →  Service  →  Data
```

Each layer knows only the layer beneath it, and only through an abstraction the
upper layer owns. This is dependency inversion: the layer that *needs* a
capability defines the interface; the layer that *provides* it satisfies that
interface.


## The three layers

- **Service** owns the domain — types, business logic, the interfaces it depends
  on, and the errors it returns. No I/O of its own.
- **Data** talks to the outside world (databases, external APIs, files) and holds
  no business logic. Each client satisfies service-layer interfaces.
- **Presentation** composes service calls and formats output for a transport
  (HTTP, CLI). No business logic. Imports service packages only, never data.


## Dependency inversion

The service defines the interface. The data layer implements it. `main` wires
the concrete type in. The service package never imports the data package.

```go
// package task (service) — defines what it needs
type Store interface {
  Get(ctx context.Context, id string) (*Task, error)
  Save(ctx context.Context, task *Task) error
}

func NewService(store Store) *Service {
  return &Service{store: store}
}

// package main — wires the concrete implementation in
service := task.NewService(sqlite.NewTaskAdapter(client))
```


## Generic client, per-interface adapter

The data layer separates transport from contract. One generic client owns the
connection; a thin adapter wraps it to satisfy one service interface.

```go
// package sqlite (data) — generic client, then an adapter that satisfies task.Store
type Client struct{ db *sql.DB }

type TaskAdapter struct{ client *Client }

func (adapter *TaskAdapter) Get(ctx context.Context, id string) (*task.Task, error) {
  // uses adapter.client, returns *task.Task and task.Err*
}
```

A `task.Service` is backed by a `task.Store`, satisfied by a `sqlite.TaskAdapter`
wrapping a `sqlite.Client`. A second contract on the same database is a second
adapter (`sqlite.UserAdapter`) over the same client — one transport, many
contracts.

Data-layer types never appear in service or presentation signatures. The adapter
returns `*task.Task` and `task.Err*`, never `sql.Rows` or `sql.ErrNoRows`.
Swapping `sqlite` for `postgres` changes only the wiring in `main`.


## Where things live

| Concern | Layer |
|---|---|
| Domain types, business logic | Service |
| Interfaces (contracts) | Service |
| Sentinel & typed errors | Service |
| Database & API clients | Data |
| Interface adapters | Data |
| Transport handlers | Presentation |
| Error → status mapping | Presentation |

See `references/error-handling.md` for how errors cross these boundaries.


## Presentation: HTTP

The standard library wins. Route with `net/http` (`http.ServeMux` and method
patterns since Go 1.22); do not reach for a third-party router or framework.

Mapping service errors to status codes and response bodies belongs here and
nowhere else. Handlers inspect service errors with `errors.Is` / `errors.As` and
choose the status. The service and data layers never see a status code.

```go
// package api (presentation) — the mapping lives here
func (handler *Handler) getTask(w http.ResponseWriter, r *http.Request) {
  t, err := handler.tasks.Get(r.Context(), r.PathValue("id"))
  if err != nil {
    switch {
    case errors.Is(err, task.ErrNotFound):
      http.Error(w, "not found", http.StatusNotFound)
    default:
      http.Error(w, "internal error", http.StatusInternalServerError)
    }
    return
  }
  writeJSON(w, http.StatusOK, t)
}
```

When there are many errors to map, a shared transport-neutral package may define
error codes or classes that service errors carry, and the handler maps class →
status. The class stays free of HTTP; the status and body mapping stays in the
API layer.
