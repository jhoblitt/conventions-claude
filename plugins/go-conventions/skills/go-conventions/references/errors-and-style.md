# Errors, context, and style

Owns error handling, context, goroutine lifetimes, the modern-Go position
and idiom inventory, comments and godoc, and the standard-library defaults
for HTTP, JSON, and randomness. Runs under `SKILL.md`'s precedence and
routing. The review consequences of each rule are `references/review-checks.md`;
the linters that enforce the literal forms are `templates/.golangci.yml`
(`references/lint.md`).

## Errors

- Wrap with context and `%w`: `fmt.Errorf("reading config %s: %w", path, err)`;
  the prefix names the operation, not the function.
- Inspect with `errors.Is` and `errors.As`, combine with `errors.Join`;
  never `==` on an error that could be wrapped, never a string match on
  `err.Error()`.
- Sentinels are `var ErrNotFound = errors.New("not found")`, exported when
  callers match on them.
- Error strings are lowercase and unpunctuated — they get wrapped — and
  carry no double prefix: each layer adds its own context once, and a
  callee never pre-adds the caller's.
- `github.com/pkg/errors` is denied (`references/lint.md`); `errors` and
  `fmt` are the whole toolkit.

## Context

- `context.Context` is the first parameter, named `ctx`, never stored in a
  struct.
- No `context.TODO()` or `context.Background()` where a `ctx` is in scope;
  `Background` belongs to `main`, a suite bootstrap, and nowhere else.
- A blocking call honors cancellation: it selects on `ctx.Done()` or uses
  the `…Context` variant of the API.

## Goroutines

- Every goroutine has a stated stop condition — a context it watches, a
  channel that closes, a `WaitGroup` its owner waits on — and a reader can
  find it at the `go` statement. A goroutine with no stop condition is a
  leak, whatever it does.
- Fan-out is `golang.org/x/sync/errgroup`, `errgroup.WithContext`: the
  first error propagates and cancels the rest.

## Modern Go

New code is written in modern Go: modern idioms on every added or changed
line. Untouched code is left alone — modernizing what a change touches is
part of the change and never scope creep, while a whole-package sweep is its
own change. The inventory:

- `any`, not `interface{}`; `min` and `max`; `slices` and `maps` from the
  standard library, not `x/exp`; range-over-int, `for i := range n`, and
  `for range n` when the index is unused; `strings.CutPrefix` and
  `CutSuffix` over `HasPrefix` plus `TrimPrefix`; iterators —
  `strings.SplitSeq`, `maps.Keys`, a `range`-over-func — where a loop
  consumes them; `new(expr)` over `ptr.To(v)` and `tmp := v; &tmp`;
  `math/rand/v2`.
- `go fix` and `modernize` report most of the inventory
  (`references/toolchain.md`, "go fix"; `references/lint.md`); the rest is
  read by hand.

## Comments

- A comment earns its place by saying what the code cannot: a non-obvious
  why, a constraint, a gotcha, an external reference. Code that needs
  narration to be understood is rewritten, not annotated.
- A comment that restates the code is deleted on sight. The signals: it
  restates the signature or the next line; it narrates control flow
  ("loop over the items"); it is a block comment out of proportion to the
  package's comment density; it is process or prompt residue ("as
  requested", "updated per review", `TODO(ai)`, "the user"). The remedy is
  deletion, not rewording: keep any real why, cut the rest.
- A comment orphaned by an edit — the predicate, workaround, or constraint
  it explained is gone — is deleted in the same edit; deleting beats
  updating.

## Godoc

Every exported symbol and every package has a godoc comment: a full
sentence starting with the name — `// Read returns …`,
`// Package version reports …`. revive's `exported` reports absence
(`references/lint.md`).

## Standard-library defaults

Unless the package's incumbent differs (`SKILL.md`, "Precedence"):

- HTTP: `net/http` — Go 1.22 method-and-pattern routes
  (`mux.HandleFunc("GET /items/{id}", …)`), an `http.Server` with its
  timeouts set explicitly (`ReadHeaderTimeout`, `ReadTimeout`,
  `WriteTimeout`, `IdleTimeout`), and `Shutdown(ctx)` on cancellation. No
  third-party router or framework.
- JSON: `encoding/json`, v2-backed in Go 1.27 (`references/toolchain.md`);
  `omitzero` where `omitempty` would once have been reached for.
- Randomness: `math/rand/v2`; `crypto/rand` for anything security-sensitive.
