# Toolchain

Owns the Go version and the `go` directive, what each toolchain release
changed that this canon leans on, `go fix`, code navigation, and the
plugin's hooks. Runs under `SKILL.md`'s precedence and routing. The gates
`go fix` feeds are `references/ci.md`, "Gates"; a `go fix` versus
`modernize` disagreement is `references/lint.md`, "modernize and go fix".

## Go version

- Go 1.27 (released 2026-08-19) is the default for a new or converged
  module.
- `go.mod` carries `go 1.27` — the minor only, never a patch — and no
  `toolchain` line: a patch directive forces a bump on every patch release,
  and a `toolchain` line pins what `go` downloads.
- `go mod init` writes a directive one minor below the installed toolchain
  (Go 1.26 and later); a scaffold rewrites it to `go 1.27` right after and
  drops any `toolchain` line.
- CI builds with the directive's version (`references/ci.md`, "CI").

## What each release changed

Go 1.26:

- `go fix` is the analyzer-based modernizer ("go fix" below).
- `new(expr)` allocates and initializes in one construct; it is the form
  for a pointer to a value (`references/errors-and-style.md`, "Modern Go").
- `go mod init` writes the directive one minor lower ("Go version" above).

Go 1.27:

- `encoding/json` is backed by its v2 implementation; `omitzero` is the tag
  to reach for (`references/errors-and-style.md`, "Standard-library defaults").
- Methods may declare type parameters.
- `go test` runs vet's `stdversion` check, which reports a standard-library
  symbol newer than the module's `go` directive; the directive is what
  gates standard-library use.
- `go mod tidy` merges `require` blocks.
- `uuid` and `crypto/mldsa` are standard-library packages.

## go fix

`go fix ./...` applies the modernizers; `go fix -diff ./...` prints the
patch and writes nothing. `go fix` shares its analyzers with golangci-lint's
`modernize` but tracks a different `golang.org/x/tools` version, so the two
can disagree; the toolchain's fixer wins (`references/lint.md`, "modernize
and go fix"). The gate on its output is `references/ci.md`, "Gates"; the
review oracle built on it is `references/review-checks.md`, "Modernization".

## Navigation

- A semantic question — a symbol's definition, its references, its type,
  the diagnostics on a file — goes to gopls through the LSP tool, not to
  grep.
- The LSP tool is usually deferred, not absent: load it with `ToolSearch`
  (`select:LSP`) before concluding no language server covers the file.
  Reading its absence from the tool list as "no server" silently downgrades
  navigation to grep; lower-tier agents get this wrong without the hint.
- grep stays for what it is better at: string and pattern searches, and
  non-Go files.

## Hooks

`hooks/hooks.json` registers two hooks, run through `hooks/goconv-hook.sh`
and the stdlib-only binary it builds on first use (the carve-out is
`references/cli.md`, "Hook and launcher binaries"). This section is the
output spec the binary's package doc names. The binary fails open: every
error path is silent, nothing is logged, and its exit status is always 0, so
a hook that cannot do its job never disturbs the tool call or session it
wraps. `GOCONV_HOOK=off` silences both hooks.

- PostToolUse on `Write|Edit` of a `.go` file: the file is formatted in
  place with `gofumpt -w` when gofumpt is on `PATH`, else `gofmt -w`. When
  the bytes changed, `additionalContext` is
  `<tool> rewrote <path> on disk; re-read it before the next Edit`, with
  `<path>` absolute. An unchanged file, a non-`.go` path, or any error
  produces nothing.
- SessionStart, when the event's `cwd` (falling back to
  `CLAUDE_PROJECT_DIR`) holds a `go.mod` whose `module` and `go` lines
  parse: `additionalContext` is
  `Go module <path> (go <version>): load go-conventions:go-conventions before writing Go`.
  A module path outside `[A-Za-z0-9-._~/]`, or with an empty, `.`, `..`,
  or `-`-led element; a `go` directive with a suffix (`1.27rc1`) or a
  trailing comment; no `go.mod` at all — each produces nothing.
