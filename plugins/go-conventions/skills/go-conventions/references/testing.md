# Testing

Owns the test framework and its house rules: Ginkgo v2 and Gomega for every
test, the suite bootstrap, the runner, focus and pending, test doubles,
integration specs, and what a spec must show a reviewer. Runs under
`SKILL.md`'s precedence and routing. How to write specs is the official
Ginkgo plugin's (`SKILL.md`, "Ginkgo how-to"); coverage adequacy is
`references/review-checks.md`, "Coverage adequacy"; the stdlib-`testing`
carve-out for hook binaries is `references/cli.md`, "Hook and launcher
binaries".

## Framework

- Ginkgo v2 and Gomega for all tests — unit, integration, end-to-end. No
  testify (depguard denies it, `references/lint.md`) and no bare `testing`
  assertions.
- One suite file per package, `<pkg>_suite_test.go`, rendered from
  `templates/suite_test.go`: `RegisterFailHandler(Fail)`, then
  `GinkgoConfiguration()` with `RandomizeAllSpecs = true` and
  `FailOnPending = true`, then `RunSpecs`. Randomization and the pending
  gate live in the bootstrap, not on the command line.

## Runner

- `go test -race -count=1 ./...` (`templates/Makefile`, `test`); CI adds the
  coverage flags (`references/ci.md`, "CI").
- Ginkgo rejects `-count` above 1 and `-parallel`; `-shuffle` reorders only
  top-level `Test` functions, a no-op for specs — randomization comes from
  the bootstrap.
- A committed `FIt`, `FDescribe`, or `FContext` in an otherwise passing
  suite exits 197 inside the test binary, so plain `go test` fails on it;
  ginkgolinter (`forbid-focus-container`, `forbid-spec-pollution`,
  `force-expect-to`, `validate-async-intervals`) reports it at lint time. A
  committed `PIt` or `XIt` fails through `FailOnPending`.
- The `ginkgo` CLI is optional — focus, labels, watch, `--repeat`:
  `go run github.com/onsi/ginkgo/v2/ginkgo <args>` pins it to `go.mod`'s
  version.

## Test doubles

counterfeiter, through the `tool` directive:
`go get -tool github.com/maxbrunsfeld/counterfeiter/v6`, then per package
and per interface:

```go
//go:generate go tool counterfeiter -generate

//counterfeiter:generate . Store
type Store interface {
	Get(ctx context.Context, key string) ([]byte, error)
}
```

- One `//go:generate go tool counterfeiter -generate` line per package; one
  `//counterfeiter:generate . <Iface>` per interface.
- Fakes generate into `<pkg>fakes/` beside the package and are committed;
  `make generate-check` fails a stale one (`references/ci.md`, "Gates").
- The interface is declared where it is consumed
  (`references/layout.md`, "Interfaces"), so the fake sits with the consumer.
- Not gomock, not mockery, not a hand-written fake where a generated one
  serves.

## Integration specs

- A spec that needs a real dependency carries `Label("integration")` and
  lives in a file behind `//go:build integration`; plain `go test` skips
  the file, `-tags integration` includes it.
- `templates/.golangci.yml` declares the `integration` build tag so the
  analyzers see those files (`references/lint.md`).
- Neither `make test` nor any CI job passes `-tags integration`
  (`references/ci.md`), so integration specs run on demand —
  `go test -tags integration -race ./...` — until a repository adds a job
  that needs them. That is deliberate: the default suite stays hermetic and
  fast, and nothing in CI depends on a real dependency being reachable.
- When integration coverage is expected at all is
  `references/review-checks.md`, "Coverage adequacy".

## Test quality

What a spec in a diff must show a reviewer; `references/review-checks.md`
applies these (tag `test-quality`) and owns severity and verification.

- **Useful failures.** A bare `Expect(ok).To(BeTrue())` tells a CI reader
  nothing. Carry got/want and the identifying key: a matcher whose failure
  prints both (`Equal`, `HaveKeyWithValue`, `ConsistOf`) over a boolean, and
  the key in the description — `Expect(got).To(Equal(want), "user %q", name)`.
  An `ok` is tested through the value it guards, never the boolean.
- **Table entries that differ.** Every `Entry` of a `DescribeTable` asserts
  something no sibling asserts; copy-paste rows asserting the same thing are
  deleted. The `Entry` name is what a failure prints, so each names its case
  ("empty input", "expired token"), never its index.
- **Fakes seeded per spec.** Every fake, fixture, and client is built in
  `BeforeEach` or the spec body, never in a package-level `var` that specs
  mutate: `RandomizeAllSpecs` runs specs in any order, and shared mutable
  state turns an order change into a flake.
- **No bare `time.Sleep` synchronization.** Asynchronous behavior is observed
  with `Eventually`/`Consistently` carrying an explicit timeout and polling
  interval (`.WithTimeout(d).WithPolling(p)`), or through a clock the spec
  injects and advances.
