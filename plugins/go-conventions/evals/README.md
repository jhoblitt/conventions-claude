# Eval cases

Regression gates for the go-conventions plugin's own behavior: that a
subject reading this checkout's canon writes a new binary the way
`references/cli.md`, `references/logging.md`, `references/layout.md`, and
`references/testing.md` say, plans a convergence the way `go-converge` says,
and does not reach for the things the canon exists to keep out — stdlib
`flag`, testify, `log.Printf`, `-X` ldflags.

`claude plugin eval` is not a documented command yet, so a case runs through
`evals/run-manual.sh` at the repository root, in two separate `claude -p`
passes — a subject and a grader:

```sh
./evals/run-manual.sh go-conventions canon-new-binary       # both passes
./evals/run-manual.sh go-conventions canon-new-binary -s    # subject only
```

The passes stay separate on purpose: a session that has read
`graders/criteria.md` writes an answer that satisfies it, and a session that
just wrote the canon cannot grade it honestly. The runner points the subject
at this checkout rather than the installed plugin, tells it not to read the
`graders/` directories, and tells it that a command it cannot run is a
coverage gap to record, not a result to invent.

Every case here is hermetic: no repository checkout, no network, no Go
toolchain, no `golangci-lint`, and no `allowed-tools` file — the subject has
Read, Grep, and Glob, and everything it judges or writes about is in
`prompt.md`. A subject writes Go into its answer; it never builds or runs it.
No case runs `goconv-audit`; the tool's output is pinned by the golden tests
under `tools/internal/audit/testdata/`.

| Case | Guards |
| --- | --- |
| `canon-new-binary` | Asked for a small CLI with no framework, logger, layout, or Go version named, the subject writes a cobra tree with `SilenceUsage`/`SilenceErrors`, viper under a `SUMIT` prefix read through `v.Get*`, a `log/slog` JSON handler on stderr with stdout carrying only the sum, `cmd/sumit/main.go` over `internal/` with the version from `debug.ReadBuildInfo`, `signal.NotifyContext` and exit 1 in `main`, `go 1.27`, and a Ginkgo bootstrap with `RandomizeAllSpecs`/`FailOnPending` plus a real spec — instead of the flat `flag`-and-`fmt.Println` program with a testify test the request invites. The `-X` ldflags and `pkg/` conditions bite only on an answer that volunteers a build file or that layout. |
| `converge-gap-report` | Asked what converging a flat Go service would change, the subject records that it cannot run `goconv-audit`, puts the go-directive, toolchain line, lint config, Makefile, CI, release and Dependabot files in a tooling phase applied first, then orders the migrations `layout` → `cli`+`logging`+`version` → `errors`/`testing` and says they wait for the user; it never claims to have run a command and never keeps or adds `-X` ldflags. |
| `review-silent-failure` | A review anchors a blocker or changes-requested finding on the `err` shadowed inside the write loop, naming the nil return the stale outer check produces; raises the `%v` wrap separately and ties it to the caller's `errors.Is` never matching; records the unrunnable oracles as a coverage gap; and leaves the already-modern `for i := range len(...)` line, unchanged context, and linter-owned nits alone. |
| `precedence-project-plugin` | With the repository's `CLAUDE.md` keeping `github.com/pkg/errors` and zerolog and the package already using both, a review raises nothing about either, states in a line that the repo `CLAUDE.md` and the incumbent package idiom outrank the canon, and still reports the `defer` inside the loop — recommending `log/slog` or `%w` anywhere, including as a converge gap, fails. |
