# AGENTS

This file summarizes how to work in the Lume codebase.
It is intended for agentic coding tools.

## Project Snapshot

- Language: Go (module `lume`)
- Go version: 1.25.5 (from `go.mod`)
- Entry point: `main.go`
- CLI framework: `github.com/spf13/cobra`
- Core packages:
  - `cmd` (CLI wiring and flags)
  - `internal/timewarrior` (parsing Timewarrior data)
  - `internal/report` (report generation, markdown output)

## Build / Lint / Test

### Build

- Build binary:
  - `go build -o lume .`
- Run locally:
  - `./lume --help`

### Lint / Format

No explicit lint config is present.
Use standard Go tooling:

- Format all Go files:
  - `gofmt -w .`
- Vet (basic static checks):
  - `go vet ./...`

### Tests

No `*_test.go` files are present in this repository.
If tests are added, prefer these standard commands:

- Run all tests:
  - `go test ./...`
- Run a single package:
  - `go test ./internal/timewarrior`
- Run a single test by name:
  - `go test ./internal/timewarrior -run TestParseLine`

## Code Style Guidelines

These guidelines reflect current patterns in the codebase.
If new patterns are introduced, keep them consistent and minimal.

### Imports

- Use Go standard formatting (gofmt).
- Group imports as:
  1. Standard library
  2. Third-party
  3. Local module (`lume/...`)
- Keep import blocks sorted by gofmt.

### Formatting

- Use tabs for indentation (gofmt).
- Keep lines readable; prefer small helper functions.
- Use explicit variable names for clarity over brevity.
- Keep file-level organization top-to-bottom:
  - types
  - public functions
  - helpers

### Types

- Prefer small structs with clear ownership.
- Use concrete types unless an interface adds value.
- Keep time-related values as `time.Time` or `time.Duration`.
- Use slices and maps directly instead of custom wrappers.

### Naming Conventions

- Use Go naming conventions:
  - `CamelCase` for exported symbols
  - `camelCase` for unexported
- Keep file and package names short and descriptive.
- Avoid abbreviations unless they are standard (`id`, `url`).

### Error Handling

- Return errors up the stack; avoid panics.
- Wrap errors with context using `fmt.Errorf("...: %w", err)`.
- Check and return errors immediately after calls.
- Prefer early returns to reduce nesting.

### Data Parsing

- Keep parsing functions pure where possible.
- Validate parsed data before use.
- Handle malformed data by returning errors or skipping entries
  (current pattern: `parseLine` returns `(Entry, bool)`).

### Output / IO

- Create output directories with explicit permissions.
- Close files with `defer` immediately after open/create.
- Use streaming writes (e.g., `fmt.Fprintf`) for markdown output.

### Sorting and Aggregation

- When returning ordered slices, use `sort.Slice` consistently.
- Avoid hidden ordering; make sorting explicit near the output.
- Use maps for aggregation, then normalize into slices for ordering.

### CLI Behavior

- Keep user-facing output concise and informative.
- Prefer `RunE` to allow error returns to bubble up.
- Avoid side effects in `init` beyond flag setup.

## Conventions from Tooling Rules

- No Cursor or Copilot rules found in:
  - `.cursor/rules/`
  - `.cursorrules`
  - `.github/copilot-instructions.md`

## Agent Workflow Tips

- Read relevant files before editing.
- Keep changes scoped and avoid refactors unless requested.
- Prefer small, composable helpers in `internal/*` packages.
- Run `gofmt -w .` on Go changes.
- If adding tests, mirror existing package layout and use table-driven tests.
