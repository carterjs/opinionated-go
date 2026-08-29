# opinionated-go

An opinionated Go analyzer, linter, and agent skill that enforces a coherent philosophy for writing Go. Not a collection of best practices to pick and choose from — a single subject, stated consistently across every layer of your codebase.

The rules exist because they work. They are prescriptive by design.

## Philosophy

A well-structured Go codebase follows a clear architectural pattern. A data layer states the foundational theme, the only connection to the outside world. Service layers restate it at different pitches, building on the foundation. Presentation layers — APIs, handlers — enter last, combining all voices into something coherent.

Each layer is independent. Each follows the same rules. The complexity of the whole emerges from their combination, not from any individual layer being complicated.

This is also the philosophy of cellular automata, of Unix pipes, of CSP. Simple rules, consistently applied, producing emergent order.

## What's Included

- **`SKILL.md`** — an agent skill that teaches these conventions to AI agents working on your codebase
- **Analyzer** — a `go/analysis`-based linter that enforces the rules mechanically, runnable standalone or as a golangci-lint plugin
- **`.golangci.yml`** — an opinionated configuration enabling existing linters that complement the custom analyzer
- **Integration hooks** — Claude Code hook configurations that surface violations in real time during agentic sessions

## Usage

### As an agent skill

**With the `npx skills` CLI:**

```bash
npx skills add carterjs/opinionated-go
```

**Or manually:** add to your agent's configuration (e.g., `.claude/settings.json` for Claude Code):

```json
{
  "skills": ["github.com/carterjs/opinionated-go"]
}
```

### As a standalone analyzer

```bash
# Install
go install github.com/carterjs/opinionated-go/analyzer@latest

# Check
go run github.com/carterjs/opinionated-go/analyzer@latest ./...

# Fix (where possible)
go run github.com/carterjs/opinionated-go/analyzer@latest -fix ./...
```

### As a git pre-commit hook

```bash
#!/bin/sh
go run github.com/carterjs/opinionated-go/analyzer@latest ./...
```

### With agent hooks

Add a hook to `.claude/settings.json` in your own project so Claude Code runs the analyzer against a file right after writing or editing it:

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [
          {
            "type": "command",
            "command": "jq -r '.tool_response.filePath // .tool_input.file_path' | { read -r f; case \"$f\" in *.go) go run github.com/carterjs/opinionated-go/analyzer@latest \"$f\" ;; esac; }"
          }
        ]
      }
    ]
  }
}
```

(This repo's own `.claude/settings.json` runs a different set of hooks: `gofmt`/`go build` after editing a file under `analyzer/`, and `go build && go vet && go test` before Claude stops — for developing the analyzer itself, not for linting a project with it. If you're contributing here, that's the config to look at instead of the snippet above.)

## Opinions

opinionated-go deliberately disagrees with some common Go advice. Where it does, the reasoning is documented. The short version:

- Named return values are banned, not just discouraged
- Global `slog` functions are banned — inject `*slog.Logger`
- Interfaces belong to the consumer package, not the producer
- Boolean parameters are banned — they mean a function does two things
- Third-party HTTP frameworks are banned — `net/http` is the router
- Absence has exactly one spelling per kind — `nil` for records, `(T, bool)` for values, and never both
- Test functions are counted by API surface, not by behavior — new behavior is a new table row, not a new `Test*` function

## License

MIT
