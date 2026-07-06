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

Mapping errors to the transport lives here and nowhere else. A handler turns a
service error into a code, the code into a status, and writes one consistent
body. The service and data layers never see a status code. Codes and classes
come from the shared `errcode` package — see `references/error-handling.md`.

```go
// package api (presentation)
func writeError(w http.ResponseWriter, err error) {
  code := errcode.FromError(err)
  writeJSON(w, statusForCode(code), errorResponse{
    Error: errorDetail{Code: code, Message: code.Description()},
  })
}
```

**Map by class, override by code.** Most codes take their status from their
class — every `invalid` is a 400, every `not_found` a 404. Switch on a specific
code only when it needs a status its class would not give.

```go
// Do this - class is the default, a code is the rare exception
func statusForCode(code errcode.Code) int {
  switch code {
  case errcode.TaskLocked:
    return http.StatusLocked // specific override
  default:
    return statusForClass(code.Class())
  }
}

// Not this - one case per code; the table grows without bound
func statusForCode(code errcode.Code) int {
  switch code {
  case errcode.TaskNotFound:
    return http.StatusNotFound
  case errcode.UserNotFound:
    return http.StatusNotFound
  // ...every code restated
  }
}
```

One error body shape, everywhere:

```json
{ "error": { "code": "task_not_found", "message": "the requested task does not exist" } }
```


### OpenAPI is the spec

Describe the API with OpenAPI. Because every code carries a class and a
description, the spec follows from the codes: each endpoint enumerates, per
status, the codes it can return, and the description is the code's own.

- **Enumerate.** List the codes an endpoint actually returns under each status —
  a 404 lists its one or two not-found codes; a 400 lists its validation codes.
- **Override or extend per endpoint** when it has codes the shared set lacks.
- A blanket 500 is documented once as a global fallback, not repeated per
  endpoint. Never document "may return any error."

```go
// codes GET /tasks/{id} can return, grouped into the OpenAPI responses by status
var getTaskErrors = []errcode.Code{
  errcode.TaskNotFound, // 404
}
```
