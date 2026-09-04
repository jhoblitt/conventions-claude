# conventions-claude design record

**Status:** accepted, 2026-09-03. **Non-normative** — see `AGENTS.md`, "Design
records", for what that means and what this file may hold.

## Why

A survey of eighteen of the author's Go repositories (2026-09-03) found
three golangci-lint tiers, two logging sinks, cobra in two repos and stdlib
`flag` in eleven, Ginkgo in the daemon family but stdlib `testing` in the
newest CLIs, three release mechanisms, two licenses, and no linter at all
in the two newest and largest repos. The generic Go review canon lived
inside rook-claude, reachable only from rook work, and the general GitHub
repository rules lived in a personal `CLAUDE.md` that no plugin could
point at. This marketplace gives both a normative home.

## Decided

- **Two plugins, one marketplace.** `github-conventions` (repository
  hygiene and pull-request practice) and `go-conventions` (the Go canon),
  the second depending on the first. Rejected: keeping repository hygiene
  inside the Go plugin (it would become the accidental home of rules that
  are not Go's, and Puppet and plugin repos could not use them); a separate
  `github-conventions-claude` repository (recommended by the structural
  review for its cleaner versioning — declined to land both together; the
  shared version and mixed changelog are the accepted cost).
- **Delivery in one pull request** with four internal phases, so the
  skill-review gate, which is expensive, runs once on the finished branch.
  Rejected: four sequential PRs.
- **Go 1.27, minor-only `go` directive, no `toolchain` line.** Rejected: an
  exact patch directive (forces a bump on every patch release).
- **cobra for every binary, viper for configuration.** Rejected: cobra only
  for subcommand trees (two idioms to teach); koanf (cleaner keys, unfamiliar);
  hand-rolled precedence helpers as in markdown-linkerator (re-implemented
  per repo). Recorded viper pitfalls drove the "read only through viper"
  and "required-ness in PreRunE" rules.
- **`log/slog`, JSON to stderr, format switch opt-in.** Rejected: JSON to
  stdout (breaks any tool whose stdout is data); text automatically on a
  TTY (format would depend on where the program runs).
- **Ginkgo and Gomega for every test, run through `go test`.** Rejected:
  Ginkgo for e2e only with stdlib units (the newest repos' practice); the
  `ginkgo` CLI as the canonical runner. Verified along the way: `-shuffle`
  does nothing for specs, so randomization moved into the suite bootstrap;
  a committed focus fails plain `go test`. How-to is deferred to the
  official onsi/ginkgo plugin rather than restated.
- **counterfeiter fakes** through the Go 1.24 `tool` directive. Rejected:
  hand-written fakes (the author's current practice; no generator, but every
  repo re-implements them); gomock/mockery.
- **The rich golangci-lint tier**, seeded from ceph-fleet-mcp's config.
  Rejected: the minimal tier (standard set plus modernize and ginkgolinter)
  and the middle tier. Adjusted after running the scaffold through it:
  `contextcheck` dropped (misreads cobra closures), `fmt.Fprint*` excluded
  from errcheck, gofumpt's `extra-rules` replaced by `extra.group-params`
  (deprecated in 2.13), Ginkgo dot imports whitelisted for staticcheck.
- **`go fix -diff` as a CI gate** alongside the `modernize` linter, with
  `go fix` winning disagreements. Rejected: advisory only.
- **`debug.ReadBuildInfo` only, no `-X` ldflags.** Rejected: the hybrid
  (ldflags win, buildinfo backfills) and pure ldflags. The consequence —
  images must be built from a checkout with `.git` — is met by ko
  rebuilding inside goreleaser.
- **goreleaser with ko, cosign, and an SBOM**, ghcr images, fork-guarded.
  Rejected: goreleaser without images; go-release-action matrices; a
  hand-rolled `go build` loop.
- **Apache-2.0 by default**, owned by github-conventions. Rejected: GPL-3.0;
  asking at scaffold time.
- **A small Makefile whose `check` is exactly CI.** Rejected: no Makefile;
  a tool-caching Makefile. The lint version lives in the Makefile and CI
  reads it, so it has one home.
- **ubuntu-only CI on the go.mod version**, coverage as an artifact with no
  gate. Rejected: OS matrices; Go N/N-1 matrices; Codecov; a threshold.
- **CodeQL by workflow**, dependency review, Scorecard, commitlint.
  Rejected: CodeQL default setup (a settings change the skill would have to
  ask for); harden-runner (not adopted).
- **github-conventions covers repo setup and the commit, PR, CI-watching,
  and comment rules.** Rejected: repo setup only (those rules would have
  stayed in a personal file no plugin can cite).
- **Precedence**: a project-specific plugin or the repository's own
  CLAUDE.md beats the canon; the surrounding package's idiom is the Go
  canon's extra rung. The ladder has one home, github-conventions, and the
  Go canon points at it.
- **A Kubernetes reference** loaded only when go.mod pulls k8s dependencies.
  Rejected: a k8s-free plugin.
- **A repository CLAUDE.md pointer block**, idempotent between markers.
  Rejected: nothing in the repo; a full generated CLAUDE.md.
- **Converge applies tooling without asking and migrates code per named
  area, serially in dependency order.** The first draft fanned migrations
  out; the structural review showed cli, logging, and version rewrite the
  same files and every area edited the lint config, so the fan-out became a
  chain with a tool-regenerated lint config and fan-out only on proven
  disjointness.
- **PR review fetches the head into a throwaway worktree**, reads the lint
  config from the base, and never runs `go test` on a fork head. The first
  draft ran oracles against nothing in particular.
- **Templates carry major tags; live copies carry SHA pins**, reconciled by
  a diff in CI. The first draft pinned inside templates, which dependabot
  never scans.
- **The audit tools parse with `go/parser` and read YAML statically**,
  never `go list` (which honours `toolchain` and downloads).
- **Hook binaries are stdlib-only**, so the first build works offline, and
  the launcher filters on the file path before building anything.
- **The eval harness is the manual two-pass runner**: `claude plugin eval`
  is not a documented command.
