# Makefile and CI

Owns the Makefile target contract, the build gates, the golangci-lint
version's one home, the Go half of CI, and the gomod Dependabot entry. Runs
under `SKILL.md`'s precedence and routing. `templates/Makefile`,
`templates/ci.yml`, and `templates/dependabot-gomod.yml` are the forms.
Workflow hygiene — pinning, least privilege, concurrency, timeouts,
`persist-credentials`, actionlint — is github-conventions'
`references/workflows.md`, which the copies comply with and this file claims
none of. CodeQL, dependency review, and Scorecard are github-conventions'
`references/security.md`; commitlint is github-conventions'
`references/commits.md`; the release workflow is `references/release.md`.

## Makefile

`templates/Makefile`, one per repository, at the root. Targets: `build`,
`test`, `lint`, `fmt`, `fmt-check`, `vet`, `fix`, `fix-check`, `tidy`,
`tidy-check`, `generate`, `generate-check`, `check`, `tools`, `help`,
`print-golangci-version`; `help` is the default goal.

- Every per-module target loops over `MODULES` (`references/layout.md`,
  "Template placeholders"), running its command in each directory. The
  exceptions:
  `help`, `print-golangci-version`, and `tools` are repository-wide;
  `check` is a prerequisite list; `generate-check` loops through `generate`
  and then runs one `git diff` with `$(MODULES)` as its pathspec.
- golangci-lint is not a `go tool` dependency: `make tools` installs the
  pinned version into `GOBIN`, which has to be on `PATH` for `lint`,
  `fmt`, and `fmt-check` to find it.

## Gates

`make check` green before work is called done. It is
`generate-check fmt-check vet lint fix-check tidy-check test`; `build`
(`go build ./...`) is a convenience target outside it, since vet and test
both compile the packages.
Each gate and what it tests:

- `vet`: `go vet ./...`.
- `lint`: `golangci-lint run ./...` under the repository's `.golangci.yml`
  (`references/lint.md`); `fmt-check`: `golangci-lint fmt --diff ./...`
  prints nothing.
- `fix-check`: `go fix -diff ./...` prints nothing. The test is output
  emptiness, not exit status: `go fix -diff` exits 0 with a patch pending.
- `tidy-check`: `go mod tidy -diff` is clean.
- `generate-check`: `go generate ./...` leaves no tracked file changed
  (counterfeiter fakes, `references/testing.md`, "Test doubles").
- `test`: the runner is `references/testing.md`, "Runner".

`check` and CI overlap but are not equal, and neither is a superset:

- The `checks` job runs `make vet fix-check tidy-check generate-check` — the
  part of `check` that is plain `make`.
- The `test` and `lint` jobs run those gates their own way: `go test` with
  coverage flags added, and golangci-lint through its action rather than
  `make lint`.
- No CI job runs `fmt-check`, and formatting is still gated: golangci-lint
  v2's `run` reports `formatters:` issues as ordinary issues, so an
  unformatted file fails the `lint` job (`references/lint.md`, "The house
  tier"). That dependency is what makes the missing job safe.
- CI additionally runs `govulncheck`, which `make check` does not.

## golangci-lint version

`GOLANGCI_LINT_VERSION` is defined in the Makefile and nowhere else. CI reads
it with `make -s print-golangci-version`; `make tools` installs that version
through golangci-lint's official install script. The repository README says
where the pin lives — the rendering is github-conventions'
`references/new-repo.md`, "README".

## CI

`templates/ci.yml`, on pull requests and pushes to `main`, ubuntu only, no
OS or Go-version matrix. The `go` directive is the one version CI builds
with (`references/toolchain.md`, "Go version"): the `test`, `lint`, and
`checks` jobs reach it through `actions/setup-go` with
`go-version-file: go.mod` and `check-latest: true`, and `govulncheck`,
which has no setup-go step, passes `go-version-file` to its own action.
Jobs:

- `test`: `go test -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...`,
  the profile uploaded as an artifact; there is no coverage threshold.
- `lint`: a prior step of the same job reads the version from the Makefile
  into `GOLANGCI_LINT_VERSION`, then `golangci/golangci-lint-action` v9
  runs with `version: ${{ env.GOLANGCI_LINT_VERSION }}`.
- `checks`: `make vet fix-check tidy-check generate-check`.
- `govulncheck`: `golang/govulncheck-action` v1 over `./...`.

Only `checks` fans out over `MODULES`, through `make`; `test`, `lint`, and
`govulncheck` run at the repository root and so cover the root module alone.
A multi-module repository adjusts them — a matrix over the module
directories, or the `make` target in place of the direct command.

## Dependabot

`templates/dependabot-gomod.yml` — the `gomod` ecosystem, weekly, grouped as
`go-dependencies` — is the Go entry in `.github/dependabot.yml`. It is a bare
`updates:` list item; github-conventions' `references/workflows.md`,
"Dependabot" owns the file, the preamble a written one carries, and whether
this entry is appended or written with the file.
