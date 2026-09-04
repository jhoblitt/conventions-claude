# goconv-audit output contract

Owns what `goconv-audit` prints and how it is read: the usage and exit
codes, the row fields, the full check inventory, the markdown and JSON
renderings, and the `--emit-golangci` and `--files` flags. Runs under
`SKILL.md`'s procedure, and under its reading of `references/<file>` and
`templates/<file>`.

`tools/cmd/goconv-audit/` implements this specification.
- `dir` defaults to `.` and is a Go repository root (a `go.mod` there); markdown
  is the default rendering, and the four flags exclude each other. Exit 0
  whenever the audit ran; exit 1 on a usage or I/O error (a bad flag, a `dir`
  with no `go.mod`, an unknown area), the reason on stderr.
- Every answer is read statically and offline: `go.mod` through
  `golang.org/x/mod/modfile`; imports through `go/parser` with
  `parser.ImportsOnly` over every `*.go` outside `vendor/`, `testdata/`, and `.`-
  and `_`-led directories; YAML into loose maps; the Makefile and workflow `run:`
  blocks as text. Never `go list`, never `exec`.
- Row fields: `area`, `check`, `status` (`ok`, `gap`, `skipped`), `phase`,
  `current` (what the repository has), `canon` (what it should have), `fix`
  (empty unless `gap`; a `templates/…` path in it is under the canon skill).
  `phase` is `tooling` for a file-level gap this skill applies itself,
  `migration` for one that changes code and waits on the user, else `none`.
- Markdown: `| Area | Check | Status | Phase | Current | Canon | Fix |`, its
  separator, one row per check in the order below, a blank line, then
  `N gaps (T tooling, M migration), K ok, S skipped`; in a cell `|` is escaped as
  `\|`, a newline becomes a space. JSON: the same fields, verbatim, under
  `{"rows":[…],"gaps":N,"tooling":T,"migration":M,"ok":K,"skipped":S}`.
- A check that cannot apply is `skipped`, the reason in `current`: `library` (no
  main package, so no handler of its own to install), `controller-runtime` (the
  incumbent `references/kubernetes.md` keeps), `no logging`, `no test files`,
  `no v2 lint config`.
- Under controller-runtime the scaffold's own `flag` and `go.uber.org/zap/zapcore`
  imports produce no `deps/<import path>` row at all — not a `skipped` one. That
  is the wiring `references/kubernetes.md` keeps; a standalone `go.uber.org/zap`
  logger beside it still reports.
- `logging`/`stderr-json` reports a slog handler built over `os.Stdout`, and a
  binary that builds none at all. It is not a literal search for `os.Stderr`:
  `templates/root.go` builds the handler over the writer `main` hands it.
- A check measured against a template reads it from
  `${CLAUDE_PLUGIN_ROOT}/skills/go-conventions/templates/` at runtime, so the
  template stays its one home; unset, those rows `skip` and `--emit-golangci`
  exits 1. That flag prints `templates/.golangci.yml` with `{{MODULE}}` filled
  from `go.mod`, minus every depguard `deny` entry whose `pkg` matches an import
  the repository carries today, so it lands green; the area migration restores
  each by re-running it once those imports are gone.
- `--files <area>` prints, sorted and unique, one repository-relative path per
  line. `cli`: importers of `flag`, kong or urfave/cli, every `cmd/*/main.go`,
  everything under `internal/cli/`. `logging`: importers of `log`, logrus, zap or
  zerolog, callers of `slog.New`/`slog.SetDefault`, and a main package importing
  `log/slog` — the command that has to install the handler. `testing`: every
  `_test.go` importing testify, or `testing` without Ginkgo, plus any `*fakes/`
  or `mocks/` directory. `errors`: importers of `github.com/pkg/errors`.
  `layout`: every `.go` outside `cmd/` and `internal/`. `version`: package-level
  `version`/`Version`/`commit`/`date` declarations, `.goreleaser.y*ml`, the
  Makefile, and workflows whose `run:` carries `-X`.
- A text scan reads a commented-out line as a live one: a `#` line inside a
  workflow `run:` scalar, or in the Makefile, counts. The `-X` search reads those
  two and goreleaser `ldflags` values only — never a YAML comment or an action
  input such as `with: ldflags:` — and matches `-X` only where it takes
  `importpath.name=value`, through a make or shell variable as readily as
  literally, so `ssh -X` and `curl -X POST` are not findings.
- The checks are a subset of the canon: an `ok` row means that check passed, not
  that the section owning it is satisfied. Which reference owns an area is the
  canon skill's routing table, not restated here.

| Area | Checks | Phase |
| --- | --- | --- |
| `toolchain` | `go-directive` | tooling |
| `lint` | `config`, `linters`, `formatters`, `max-issues` | tooling |
| `makefile` | `present`, `targets`, `golangci-version` | tooling |
| `ci` | `setup-go`, `race`, `lint`, `checks`, `govulncheck` | tooling |
| `release` | `goreleaser`, `kos`, `sign-sbom`, `workflow`, `no-ldflags-x` | tooling |
| `dependabot`, `gitignore`, `claude`, `version` | `gomod`, `present`, `pointer`, `package` | tooling |
| `layout` | `cmd`, `pkg-dir` | migration |
| `deps` | `cli`, `config`, `testing`, `fakes`, `logging` | migration |
| `logging` | `stderr-json` | migration |
| `deps` | `<import path>`, one per forbidden import present, `fix` naming its area | migration |
