# opinionated-go

An opinionated Go analyzer, linter, and Claude Code skill that enforces a coherent philosophy for writing Go. Not a collection of best practices to pick and choose from — a single subject, stated consistently across every layer of your codebase.

The rules exist because they work. They are prescriptive by design.

## Philosophy

A well-structured Go codebase follows a clear architectural pattern. A data layer states the foundational theme, the only connection to the outside world. Service layers restate it at different pitches, building on the foundation. Presentation layers — APIs, handlers — enter last, combining all voices into something coherent.

Each layer is independent. Each follows the same rules. The complexity of the whole emerges from their combination, not from any individual layer being complicated.

This is also the philosophy of cellular automata, of Unix pipes, of CSP. Simple rules, consistently applied, producing emergent order.

## What's Included

- **`SKILL.md`** — a Claude Code skill that teaches these conventions to AI agents working on your codebase
- **Analyzer** — a `go/analysis`-based linter that enforces the rules mechanically, runnable standalone or as a golangci-lint plugin
- **`.golangci.yml`** — an opinionated configuration enabling existing linters that complement the custom analyzer
- **Claude Code hooks** — `PostToolUse` and `Stop` hook configurations that surface violations in real time during agentic sessions

## Usage

### As a Claude Code skill

Add to your project's `.claude/settings.json`:

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

### With Claude Code hooks

See `.claude/hooks.json` in this repository for the recommended hook configuration.

## Opinions

opinionated-go deliberately disagrees with some common Go advice. Where it does, the reasoning is documented. The short version:

- Named return values are banned, not just discouraged
- `errgroup` is banned — prefer explicit goroutine ownership with `sync.WaitGroup` and `context.WithCancelCause`
- Global `slog` functions are banned — inject `*slog.Logger`
- Interfaces belong to the consumer package, not the producer
- Boolean parameters are banned — they mean a function does two things

## License

MIT
