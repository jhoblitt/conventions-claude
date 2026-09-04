# Lint

Owns the lint policy: the golangci-lint tier, why each deliberate setting is
what it is, the formatters, the exclusions, how to disagree with a linter,
and the rule for a `modernize` versus `go fix` disagreement. Runs under
`SKILL.md`'s precedence and routing. `templates/.golangci.yml` is the
enforced form; its comments carry the per-linter and per-setting rationale
for a reader in the target repository, and this file states the policy rather
than repeating them. The pinned golangci-lint version has one home,
`templates/Makefile` (`references/ci.md`, "golangci-lint version"); the gates
that run the linter are `references/ci.md`, "Gates".

## The house tier

- golangci-lint v2 (`version: "2"`), one `.golangci.yml` at the repository
  root, rendered from `templates/.golangci.yml` with `{{MODULE}}` filled
  (`references/layout.md`, "Template placeholders"). Converge rewrites it
  rather than editing it into shape.
- `linters.default: none` with an explicit `enable` list — the template's,
  grouped there by purpose: correctness, modern Go, dependency hygiene,
  style, and the two framework linters. golangci-lint's `standard` set is
  errcheck, govet, ineffassign, staticcheck, unused; gosec is NOT in it and
  is enabled here.
- govet with `enable-all` (`fieldalignment` excepted); staticcheck `all`.
- Formatters gofmt, gofumpt, and goimports; goimports' `local-prefixes` is
  a list holding the module path. `make fmt` applies them through
  `golangci-lint fmt`. `golangci-lint run` also reports them, as
  `formatters:` issues, which is what gates formatting in CI
  (`references/ci.md`, "Gates").
- `issues.max-issues-per-linter: 0` and `max-same-issues: 0`: every issue
  is reported, none truncated.
- `run.build-tags` lists `integration`, so specs behind that tag are
  analyzed (`references/testing.md`, "Integration specs").
- `exclusions.generated: lax`, and test files are excluded from the four
  style linters — gosec, unparam, revive, gocritic; the correctness linters
  still apply there.
- depguard denies `github.com/pkg/errors`, `io/ioutil`, `math/rand` (v1),
  `golang.org/x/exp/slices`, `golang.org/x/exp/maps`, testify, logrus, zap,
  and zerolog; each entry's `desc` names the replacement, and the rules are
  `references/errors-and-style.md`, `references/testing.md`, and
  `references/logging.md`. `gomodguard` is never used: deprecated since
  golangci-lint v2.12, and depguard covers the deny list.

## Deliberate settings, and why

- `contextcheck` is not enabled: it misreads cobra's `cmd.Context()`
  closures as dropping the context (verified 2026-09-03 on
  `templates/root.go`).
- errcheck's `exclude-functions` covers `fmt.Fprint`, `fmt.Fprintf`, and
  `fmt.Fprintln`; the template's comment carries why a failed write to an
  output stream is not an error worth checking.
- gofumpt's setting is the nested mapping

  ```yaml
  gofumpt:
    extra:
      group-params: true
  ```

  and not the `extra-rules` key it replaces: golangci-lint 2.13.2 loads
  `extra-rules` but logs
  ``[formatter] gofumpt: `extra-rules` is deprecated, please use
  `extra.group-params` instead.`` (observed 2026-09-04).
- staticcheck's `dot-import-whitelist` holds Ginkgo and Gomega, which are
  dot-imported by design; without it ST1001 fires on every suite.

## modernize and go fix

`modernize` (golangci-lint's) and `go fix` (the toolchain's,
`references/toolchain.md`, "go fix") run the same analyzers from different
`golang.org/x/tools` versions and can disagree. When they do, `go fix` wins:
the analyzer goes into `modernize.disable` with a comment naming the
disagreement, and this section records it. Recorded disagreements: none.

## Disagreeing with a linter

- A suppression is `//nolint:<linter> // <reason>` on the line it covers:
  the linter named, the reason on the same line, never a bare `//nolint`.
  The reason says why the report is wrong here, not what the code does.
- A config-level change — a disabled linter, a new exclusion — carries a
  comment beside it saying why.
- In a diff, a suppression or config change is judged under the pre-change
  config (`references/review-checks.md`, "Scope").
