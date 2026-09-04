---
name: go-conventions
description: Use when writing, editing, generating, or refactoring Go code; creating or editing go.mod, .golangci.yml or .golangci.yaml, a Go Makefile, .goreleaser.yaml or .goreleaser.yml, or a Go CI workflow; choosing a Go library, CLI, config, logging, or test framework; or answering how something should be done in Go.
---

# Go conventions

The canon for how Go is written, tested, logged, linted, built, and
released — the house rules a Go repository follows on top of
github-conventions, which owns repository hygiene and pull-request practice.
It is consulted, not run: `go-converge` audits and converges an existing
repository against it, `go-review` reviews a Go diff against it,
`go-new-project` scaffolds a new one from it. A path of the form
`references/<file>` or `templates/<file>` is relative to this skill's
directory; "github-conventions' `references/<file>`" is the sibling plugin's.
A file under `templates/` is the enforced form of a rule, never its home;
where each lands is `skills/go-new-project/SKILL.md`'s.

## Precedence

The ladder is github-conventions' `SKILL.md`, "Precedence". This canon
depends on that plugin: if the Skill tool lists no `github-conventions:*`
skill, stop and say to install it. This canon adds one rung, and this
section is its home: between the repository's own `CLAUDE.md` and this
canon sits the surrounding package's consistent idiom. A change that
matches that idiom is correct where this canon disagrees; a new pattern
inconsistent with the package is the finding, not the package's deviation
from canon. `references/review-checks.md` applies the rung; nothing
restates it.

## Always

- Go 1.27, a minor-only `go` directive — `references/toolchain.md`, "Go version".
- gopls through the LSP tool before grep for semantic questions — `references/toolchain.md`, "Navigation".
- cobra and viper for every binary — `references/cli.md`.
- `log/slog`, JSON to stderr — `references/logging.md`.
- Ginkgo v2 and Gomega for every test — `references/testing.md`.
- `debug.ReadBuildInfo` for the version, no `-X` ldflags — `references/layout.md`, "Version".
- The house lint tier, `templates/.golangci.yml` — `references/lint.md`.
- Modern Go on every added or changed line — `references/errors-and-style.md`, "Modern Go".
- `make check` green before the work is called done — `references/ci.md`, "Gates".

## Reference routing

Read a reference when the work touches its trigger; skip the rest. The
rules above apply to every trigger.

| Working on | Read |
|---|---|
| a new binary or CLI, a cobra command, flags | `references/cli.md` |
| configuration, environment variables, a config file | `references/cli.md`, "Configuration" |
| logging, a logger, log output | `references/logging.md` |
| tests, specs, suites, fakes, or test doubles | `references/testing.md` |
| `go.mod`, the `go` directive, the toolchain, `go fix` | `references/toolchain.md` |
| `.golangci.yml`, a `//nolint`, a linter's complaint | `references/lint.md` |
| package layout, `main`, `internal/`, a new package | `references/layout.md` |
| versioning, a release, goreleaser, a container image | `references/release.md`, and `references/layout.md`, "Version" |
| the Makefile, a CI workflow, `.github/dependabot.yml` | `references/ci.md` |
| errors, contexts, goroutines, comments, godoc, HTTP, JSON | `references/errors-and-style.md` |
| a module importing `k8s.io/*` or `sigs.k8s.io/controller-runtime` | `references/kubernetes.md` |
| reviewing a Go diff | `references/review-checks.md` |

## Scripts

```sh
bash "${CLAUDE_PLUGIN_ROOT}/tools/run.sh" goconv-audit [--json|--markdown|--emit-golangci|--files <area>] [dir]
```

Audits a repository against this canon; emits the house lint config for the
repository's imports; lists the files an area's migration touches. The
launcher fails loud: a non-zero exit is a real failure, never an empty
result. Its contract is `skills/go-converge/references/goconv-audit.md`.

## Ginkgo how-to

Writing specs, tables, async assertions, parallelism, and debugging are the
official plugin's — `/plugin marketplace add onsi/ginkgo`, then
`/plugin install ginkgo@ginkgo`; `references/testing.md` carries house rules only.
