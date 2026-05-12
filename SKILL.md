# go-fugue

A single subject, stated consistently across every layer. These rules are prescriptive — follow them exactly. When in doubt, do not hedge: apply the rule.

---

## Naming

- **Full words only.** `Document` not `Doc`, `Request` not `Req`, `Response` not `Resp`, `Configuration` not `Cfg`, `Message` not `Msg`, `Error` not `Err` (as a name — `err` as a variable is correct).
- **Initialisms always uppercase.** `ID`, `URL`, `HTTP`, `API`, `JSON` — never `Id`, `Url`, `Http`.
- **Variable length scales with scope.** Single letters only in the tightest loops where meaning is unambiguous. Variables used across more than a few lines must be fully descriptive: `partitionKey` not `pk`, `errorCount` not `ec`.
- **`ctx` always `ctx`.** Any `context.Context` parameter is always named `ctx`. Never `c`, `context`, or anything else.
- **`err` always `err`.** Any `error` variable is always named `err`. Never `e`, `erro`, or anything else.
- **Receiver names.** A short, readable word derived from the type name — `store`, `mock`, `adapter`, `service`. Never a single letter. Never `s`, `m`, `a`. Consistent across all methods on the type. Warn on 1–2 character receivers unless the type name is also 1–2 characters.
- **Package names.** Lowercase, single word, no underscores. Must match the directory name. Never `util`, `common`, `helpers`, `shared`, or similar generic names.
- **File names.** No underscores except `_test.go` and `_<platform>_test.go` patterns. Name files after their primary concept (`store.go`, `schema.go`), never their role (`helpers.go`, `utils.go`).

---

## Comments

- All exported identifiers must have a godoc comment beginning with the identifier's name.
- Unexported identifiers: comment only when the purpose is not clear from the name and context alone.
- Inline comments: only when the *why* is non-obvious — a hidden constraint, subtle invariant, or known workaround. Never use comments to label sections of code.

---

## Error Handling

- **Always wrap with `%w`.** Use `fmt.Errorf("doing X: %w", err)`. Never return a naked `err` directly. The `%w` verb preserves the error chain.
- **Sentinel errors at package level only.** `var ErrNotFound = errors.New("not found")` at the top of the file. Never `return errors.New("...")` inline.
- **Typed error structs** when callers need to inspect details beyond a sentinel check.
- **`errors.Is` / `errors.As` only.** Never string-match on error messages.
- **Errors always last return value.**
- **Error strings:** lowercase, no trailing punctuation. `"reading file"` not `"Reading file."`.
- **Indent error flow.** Return early. Keep the happy path at the left margin.
- **No `panic` in library code.** Only acceptable in `main` or test setup helpers. Always prefer returning an error.

---

## Function & Method Design

- **Maximum 4 parameters.** When more are needed, use a config struct or functional options — only for truly optional configuration, not as a workaround for required arguments.
- **`context.Context` always first** if present.
- **`*slog.Logger` always second** if present, consistently named `log` or `logger`.
- **No boolean parameters.** A boolean parameter means the function does two things. Split it or use a typed option.
- **No named return values.** Ever. They obscure control flow and invite bare returns.
- **No `func` parameters.** Never pass a `func()` or callback as a parameter. Define an interface with a method instead. Exception: single-use stdlib callbacks like `sort.Slice` are acceptable.
- **Return concrete types.** Always return concrete types from functions and constructors. The only exception is `error`. Never return an interface type.
- **Function length: 60 lines maximum** (excluding tests). A function approaching 60 lines is a signal it is doing too much regardless of whether the limit is reached.

---

## Interfaces

- **Interfaces belong to the consumer.** Define interfaces in the package that uses them, not the package that implements them. The data layer never defines the interfaces it satisfies — the service layer does.
- **Never define an interface unused in the current package.** Speculative interfaces are banned.
- **Keep interfaces small.** The bigger the interface, the weaker the abstraction.
- **No `any` / `interface{}` in public APIs** — warn. Use a concrete type or a well-defined interface.
- **No channels, `sync.WaitGroup`, or `func` types in exported function signatures** — warn. Wrap coordination primitives behind a concrete type or interface.

---

## Global State & Configuration

- **`os.Getenv` only in `main`** or a `config` package loaded exclusively by `main`. Domain packages — stores, services, adapters — must never read environment variables directly.
- **No global `slog` functions.** Never call `slog.Info`, `slog.Error`, `slog.Debug` etc. at the package level. Inject a `*slog.Logger` via constructor or parameter.
- **No `init()` functions** — warn. Initialization logic belongs in constructors or `main`.
- **No `errgroup`.** It is banned. Use explicit goroutine creation, `sync.WaitGroup` for lifecycle management, and `context.WithCancelCause` when cancellation with a cause is appropriate.
- **Dependency injection via constructor arguments or receiver fields.** Never closures capturing external state.

---

## Structs & Types

- **No exported fields on structs that have methods.** If a type has behavior, control access through methods.
- **Constructors required** when a struct has unexported fields or requires custom zero-value initialization. Name them `New<Type>`.
- **Zero value must be valid and usable** without a constructor for simple value types.
- **Config structs use `<= 0` checks for defaults.** Domain packages own their defaults; callers that don't need tuning pass `Config{}`.
- **Typed constants over raw string/int constants.** `type Status string` with typed constants beats `const StatusActive = "active"`.
- **No magic numbers.** All numeric literals beyond 0 and 1 must be named constants.

---

## Concurrency

- **Synchronous by default.** Never hide goroutines, channels, or async I/O inside library functions. Let the caller decide when to add concurrency.
- **No fire-and-forget goroutines.** Every goroutine must have a clear owner and defined lifetime, managed with `sync.WaitGroup` or a done channel.
- **`context.WithCancelCause` over `context.WithCancel`** when cancellation reason is meaningful to the caller.

---

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
- **No `_test.go` file without a corresponding `.go` source file** — error.

---

## Layered Architecture

```
Presentation  →  Service  →  Data
```

- **Service layer** owns all business logic and all interface definitions. No I/O or persistence logic lives here.
- **Data layer** contains database adapters, external API clients, file I/O, and other persistence concerns. Satisfies interfaces defined by the service layer. No business logic lives here.
- **Presentation layer** composes service calls and formats output. Imports only service packages — never data-layer packages directly. No business logic lives here.

---

## Testing

If writing or modifying test files, **read `references/testing.md` before proceeding.**
