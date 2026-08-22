---
name: opinionated-go
description: Enforces a coherent, prescriptive philosophy for writing Go
---

# opinionated-go

A single subject, stated consistently across every layer. These rules are prescriptive — apply them as written.

## Naming

- **Full words only.** `Document` not `Doc`, `Request` not `Req`, `Response` not `Resp`, `Configuration` not `Cfg`, `Message` not `Msg`, `Error` not `Err` (as a name — `err` as a variable is correct).
- **Initialisms always uppercase.** `ID`, `URL`, `HTTP`, `API`, `JSON` — never `Id`, `Url`, `Http`.
- **Variable length scales with scope.** Single letters only in the tightest loops where meaning is unambiguous. Variables used across more than a few lines must be fully descriptive: `partitionKey` not `pk`, `errorCount` not `ec`. Exceptions:
  - A name read exactly once is fine however far below its declaration that single reference sits — there is nothing to hold in your head.
  - `n` is always acceptable. It is the settled Go name for a count or a byte tally, in loops and anywhere else.
  - `t`, `b`, `m`, `f` for `*testing.T`, `*testing.B`, `*testing.M`, `*testing.F`. These hold anywhere, including a helper declared outside a `_test.go` file.
  - `w` for `http.ResponseWriter`, `r` for `*http.Request`.
- **Never shadow a built-in with a function.** `func len(...)`, `func copy(...)`, `func max(...)` take the built-in's name out of the file. A **method** may use these names freely: it is only ever reached through a receiver, so it shadows nothing.
- **`ctx` always `ctx`.** Any `context.Context` parameter is always named `ctx`. Never `c`, `context`, or anything else.
- **`err` always `err`.** Any `error` variable is always named `err`. Never `e`, `erro`, or anything else.
- **Receiver names.** A short, readable word derived from the type name — `store`, `mock`, `adapter`, `service`. Never a single letter. Never `s`, `m`, `a`. Consistent across all methods on the type.
- **Getter methods.** Use the Go idiom: `Name()` not `GetName()`. Setter methods are `SetField()`. Never `Get<Field>()` prefix.
- **Package names.** Lowercase, single word, no underscores. Must match the directory name. Never `util`, `common`, `helpers`, `shared`, or similar generic names.
- **File names.** No underscores except `_test.go` and `_<platform>_test.go` patterns. Name files after their primary concept (`store.go`, `schema.go`), never their role (`helpers.go`, `utils.go`).
- **No stuttering names.** Exported names must not repeat the package name. In package `server`, use `Handler` not `ServerHandler`. In package `user`, use `Store` not `UserStore`.

## Comments

- All exported identifiers must have a godoc comment beginning with the identifier's name.
- Unexported identifiers: comment only when the purpose is not clear from the name and context alone.
- Inline comments: only when the *why* is non-obvious — a hidden constraint, subtle invariant, or known workaround.
- **Comments bump up against their symbol.** A blank line between a comment and the declaration below it makes it documentation for nothing. The same goes for a comment with no declaration under it at all — delete it or attach it.
- **Empty `//` lines are fine as paragraph breaks** inside a multi-line comment; that is how godoc marks a new paragraph.
- **Never use repeated characters as a divider.** No `// ─── Matching ───`, no `// -----`, no `// ///`. Grouping comments are not a substitute for splitting the file.

## Error Handling

See `references/error-handling.md` for detailed error handling patterns and examples.

## Functions, Methods & Design

See `references/functions.md` for detailed guidance on function design, methods, interfaces, and constructors.

## Structs & Types

- **No exported fields on structs that have methods.** If a type has behavior, control access through methods.
- **Constructors required** when a struct has unexported fields or requires custom zero-value initialization. Name them `New<Type>`.
- **Zero value must be valid and usable** without a constructor for simple value types.
- **Config structs use `<= 0` checks for defaults.** Domain packages own their defaults; callers that don't need tuning pass `Config{}`.
- **Typed constants over raw string/int constants.** `type Status string` with typed constants beats `const StatusActive = "active"`.
- **No magic numbers.** All numeric literals beyond `0` and `1` must be named constants — `return 10 << 20` is a `maxUploadBytes` waiting to be named. Constant declarations are where a literal belongs; test files are exempt, since concrete values are the point of a table.

## Concurrency

- **Synchronous by default.** Never hide goroutines, channels, or async I/O inside library functions. Let the caller decide when to add concurrency.
- **No fire-and-forget goroutines.** Every goroutine must have a clear owner and defined lifetime, managed with `sync.WaitGroup` or a done channel.
- **`context.WithCancelCause` over `context.WithCancel`** when cancellation reason is meaningful to the caller.
- **The root context belongs to `main`.** `context.Background()` and `context.TODO()` are for `main` — everywhere else, accept a `context.Context` from the caller. A package that roots its own context has quietly opted out of the caller's cancellation and deadline. In tests the answer is `t.Context()`, never `context.Background()`.

## Global State & Configuration

- **`os.Getenv` only in `main`** or a `config` package loaded exclusively by `main`. Domain packages — stores, services, adapters — must never read environment variables directly.
- **No global `slog` functions.** Never call `slog.Info`, `slog.Error`, `slog.Debug` etc. at the package level. Inject a `*slog.Logger` via constructor or parameter.
- **No `init()` functions.** Initialization logic belongs in constructors or `main`.
- **Concurrency mechanisms are internal implementation.** Channels, `sync.WaitGroup`, `errgroup`, and other concurrency primitives belong in function implementations, never as function parameters or interface methods. Hide concurrency details from callers.
- **Dependency injection via constructor arguments or receiver fields.** Never closures capturing external state.

## Package & File Organization

- **Dependencies flow strictly downward:** Presentation → Service → Data. Never import upward or across layers.
- **Presentation-layer packages under `internal/`.** HTTP handlers, CLI commands, and other I/O boundaries are not reusable and must not be importable externally.
- **One purpose per package.** If naming a package is difficult, it needs splitting.
- **File organization within a package:**
  - **Constants at the top**, above the functions that spend them. A file's constants are its vocabulary; a reader meets them before the code.
  - **The primary type opens its namesake file.** Open `store.go` in package `store` and the first thing on the screen is the type the package is about — not a helper, not a free function.
  - **A constructor sits immediately after its type.** `NewStore` is the very next declaration after `type Store`, in the same file. Nothing in between.
  - **Errors live in `errors.go`** once a package has more than one, grouped by the feature that returns them. A lone sentinel may stay beside the code that returns it.
  - Exported symbols first, unexported below.
  - Methods on a type belong in the same file as the type definition.
  - Helpers defined immediately after their first use, in order of use.
  - Prefer extending an existing file over creating a new one.
  - A new file is justified only when a self-contained concept has outgrown its current home.
- **Delete dead code.** Never leave unused functions, variables, types, or imports.
- **No `_test.go` file without a corresponding `.go` source file.**

## Layered Architecture

See `references/architecture.md` for detailed patterns and examples.

```
Presentation  →  Service  →  Data
```

- **Service layer** owns all business logic and all interface definitions. No I/O or persistence logic lives here.
- **Data layer** contains database adapters, external API clients, file I/O, and other persistence concerns. Satisfies interfaces defined by the service layer. No business logic lives here.
- **Presentation layer** composes service calls and formats output. Imports only service packages — never data-layer packages directly. No business logic lives here.
- **Dependency inversion.** The consuming layer owns the interface; the providing layer satisfies it. A generic client in the data layer is wrapped by a per-interface adapter (`sqlite.TaskAdapter` satisfies `task.Store`).
- **Errors are translated at each boundary.** The service layer defines its domain errors. The data-layer adapter maps downstream errors into them — never propagate a downstream error blindly. The presentation layer maps service errors to the transport.
- **HTTP uses the standard library.** Route with `net/http`; no third-party frameworks. Mapping service errors to status codes and bodies lives in the presentation layer only — never in the service or data layer.
- **Error codes are a shared, transport-neutral vocabulary.** An `errcode` package names service errors as coded values, each with a `Class` and a description, and owns the error → code mapping. It imports the service packages, never the reverse, and knows nothing about HTTP.
- **The API is specified with OpenAPI.** Each endpoint enumerates the error codes it can return per status; descriptions come from the codes. Map status by class, override by specific code.

## Testing

See `references/testing.md` for complete testing conventions.

## Enforcing These Rules

**Check compliance** using the opinionated-go analyzer:

```bash
go run github.com/carterjs/opinionated-go/analyzer@latest ./...
```

The analyzer precisely reports violations. Each violation requires human review and manual fixing — the rules are too intertwined with intent and context for safe automated correction.

When working on code, apply these rules to every change, and use the analyzer to verify compliance. Refactoring, bug fixes, and new features should all follow this philosophy — there are no exceptions based on circumstance.
