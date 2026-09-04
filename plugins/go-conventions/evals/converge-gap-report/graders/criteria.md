The fixture is a flat Go service: `go.mod` with a patch `go` directive
(`go 1.25.0`) and a `toolchain` line; `main.go` at the repository root, in
package `main`, importing the standard `flag` and `log` and carrying a
package-level `version` variable; `main_test.go` using testify's `require`;
a Makefile whose `build` target stamps `-X main.version`; and no
`.golangci.yml`, `.github/`, `.goreleaser.yaml`, `CLAUDE.md`, or `.gitignore`.

Under `go-converge`: `goconv-audit` produces the gap table, but the subject
cannot run it here, so the coverage gap has to be recorded rather than the
table invented. Converge then runs in two phases. The **tooling** phase is
applied first, without waiting: the `go` directive becomes `go 1.27` and the
`toolchain` line goes; `.golangci.yml` is generated; the Makefile, `ci.yml`,
`release.yml`, `.goreleaser.yaml`, `.gitignore`, the gomod Dependabot entry,
`internal/version/`, and the `CLAUDE.md` pointer block land from the
templates. The **migration** phase waits for the user to name an area, and
its areas run in dependency order because they share files: `layout` alone,
then `cli` + `logging` + `version` as one dispatch, then `errors`, then
`testing`.

The regressions this case exists to catch: running every migration at once,
claiming to have run the audit tool or applied a change, and treating the
`-X main.version` stamp as something to preserve or add.

Pass if and only if ALL of:

1. The plan names both `go.mod` fixes: the directive raised to a minor-only
   `go 1.27`, and the `toolchain` line dropped.
2. The plan names the missing `.golangci.yml` as something converge creates.
3. The plan names the Makefile, the CI workflow(s), the release/goreleaser
   setup, and the Dependabot gomod entry as files it lands.
4. Every item in conditions 1-3 is placed in the tooling phase, and that
   phase is stated to run BEFORE the migrations.
5. `layout` (the root `package main` moving under `cmd/<name>`) is named as
   the first migration.
6. `cli`, `logging`, and `version` are named as one dispatch — migrated
   together, after `layout`. (They share `main.go`, so grouping them is the
   plan, not a shortcut.)
7. `errors` and/or `testing` (testify to Ginkgo and Gomega) are named as
   later migrations, and the migrations as a whole are stated to wait for the
   user to name an area.
8. The answer says the audit tool would have produced the table and records
   that it could not be run here.

Fail if any of:

- The plan proposes all the migrations in one pass, or dispatches `layout`
  together with another area, or puts `layout` after `cli`/`logging`/`version`.
  (`cli` + `logging` + `version` as one dispatch is required, not a failure.)
- The answer claims to have run `goconv-audit`, `make check`, or any command,
  or to have edited a file; the subject can only read.
- The plan keeps, adds, or recommends `-X` ldflags anywhere.
- The plan changes the repository's license, or adds one as part of the Go
  convergence rather than through the github-conventions half.

Detail beyond these — naming the counterfeiter `tool` directive, the
`internal/cli` package, the stderr JSON handler, or noting that
`github-conventions:github-converge` runs first — neither passes nor fails
this case.
