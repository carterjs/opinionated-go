# opinionated-go

A single subject, stated consistently across every layer. These rules are prescriptive — apply them as written.

## Naming

- **Full words only.** `Document` not `Doc`, `Request` not `Req`, `Response` not `Resp`, `Configuration` not `Cfg`, `Message` not `Msg`, `Error` not `Err` (as a name — `err` as a variable is correct).
- **Initialisms always uppercase.** `ID`, `URL`, `HTTP`, `API`, `JSON` — never `Id`, `Url`, `Http`.
- **Variable length scales with scope.** Single letters only in the tightest loops where meaning is unambiguous. Variables used across more than a few lines must be fully descriptive: `partitionKey` not `pk`, `errorCount` not `ec`.
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
- **No magic numbers.** All numeric literals beyond 0 and 1 must be named constants.

## Concurrency

- **Synchronous by default.** Never hide goroutines, channels, or async I/O inside library functions. Let the caller decide when to add concurrency.
- **No fire-and-forget goroutines.** Every goroutine must have a clear owner and defined lifetime, managed with `sync.WaitGroup` or a done channel.
- **`context.WithCancelCause` over `context.WithCancel`** when cancellation reason is meaningful to the caller.

## Global State & Configuration

- **`os.Getenv` only in `main`** or a `config` package loaded exclusively by `main`. Domain packages — stores, services, adapters — must never read environment variables directly.
- **No global `slog` functions.** Never call `slog.Info`, `slog.Error`, `slog.Debug` etc. at the package level. Inject a `*slog.Logger` via constructor or parameter.
- **No `init()` functions.** Initialization logic belongs in constructors or `main`.
- **No `errgroup`.** Use explicit goroutine creation, `sync.WaitGroup` for lifecycle management, and `context.WithCancelCause` when cancellation with a cause is appropriate.
- **Dependency injection via constructor arguments or receiver fields.** Never closures capturing external state.

## Package & File Organization

- **Dependencies flow strictly downward:** Presentation → Service → Data. Never import upward or across layers.
- **Presentation-layer packages under `internal/`.** HTTP handlers, CLI commands, and other I/O boundaries are not reusable and must not be importable externally.
- **One purpose per package.** If naming a package is difficult, it needs splitting.
- **File organization within a package:**
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

## Testing

See `references/testing.md` for complete testing conventions.
