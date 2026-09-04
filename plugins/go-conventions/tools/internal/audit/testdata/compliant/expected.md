| Area | Check | Status | Phase | Current | Canon | Fix |
| --- | --- | --- | --- | --- | --- | --- |
| toolchain | go-directive | ok | none | go 1.27 | go 1.27, the minor only, and no toolchain line |  |
| lint | config | ok | none | .golangci.yml, version: 2 | a .golangci.yml at the root, golangci-lint v2 (version: "2") |  |
| lint | linters | ok | none | enabled: govet, staticcheck, errcheck, errorlint, nilerr, nilnesserr, bodyclose, noctx, fatcontext, copyloopvar, recvcheck, ineffassign, unused, unparam, modernize, exptostd, intrange, perfsprint, usetesting, depguard, gomoddirectives, gocritic, revive, misspell, unconvert, gosec, ginkgolinter, sloglint | every linter templates/.golangci.yml enables |  |
| lint | formatters | ok | none | enabled: gofmt, gofumpt, goimports | the formatters templates/.golangci.yml enables |  |
| lint | max-issues | ok | none | max-issues-per-linter: 0, max-same-issues: 0 | issues.max-issues-per-linter and max-same-issues are both 0 |  |
| makefile | present | ok | none | Makefile | a Makefile at the repository root |  |
| makefile | targets | ok | none | defined: help, print-golangci-version, tools, build, generate, generate-check, fmt, fmt-check, vet, lint, fix, fix-check, tidy, tidy-check, test, check | every target templates/Makefile defines |  |
| makefile | golangci-version | ok | none | GOLANGCI_LINT_VERSION defined | GOLANGCI_LINT_VERSION is defined in the Makefile |  |
| ci | setup-go | ok | none | .github/workflows/ci.yml, .github/workflows/release.yml | a workflow sets Go up with go-version-file: go.mod |  |
| ci | race | ok | none | .github/workflows/ci.yml | a workflow runs the suite with the race detector |  |
| ci | lint | ok | none | .github/workflows/ci.yml | a workflow runs golangci/golangci-lint-action |  |
| ci | checks | ok | none | .github/workflows/ci.yml | a workflow runs the go fix and go mod tidy checks |  |
| ci | govulncheck | ok | none | .github/workflows/ci.yml | a workflow runs govulncheck over ./... |  |
| release | goreleaser | ok | none | .goreleaser.yaml, version: 2 | .goreleaser.yaml at the root, version: 2 |  |
| release | kos | ok | none | present: kos | a kos block builds the image from the same checkout |  |
| release | sign-sbom | ok | none | present: signs, sboms | signs and sboms blocks: cosign over the checksums, syft over the archives |  |
| release | workflow | ok | none | .github/workflows/release.yml | a workflow on v* tags runs goreleaser/goreleaser-action |  |
| release | no-ldflags-x | ok | none | no -X ldflags | no -X ldflags anywhere; the version is the build stamp |  |
| dependabot | gomod | ok | none | ecosystems: github-actions, gomod | .github/dependabot.yml carries a gomod updates entry |  |
| gitignore | present | ok | none | .gitignore | .gitignore at the repository root |  |
| claude | pointer | ok | none | pointer block present | CLAUDE.md carries the go-conventions pointer block |  |
| version | package | ok | none | internal/version | internal/version reads debug.ReadBuildInfo |  |
| layout | cmd | ok | none | every main package is under cmd/<name> | every package main lives in cmd/<name> |  |
| layout | pkg-dir | ok | none | no pkg/ | no pkg/ in a module that ships a binary |  |
| deps | cli | ok | none | github.com/spf13/cobra | cobra builds every command tree |  |
| deps | config | ok | none | github.com/spf13/viper | viper resolves flags, environment, and config file |  |
| deps | testing | ok | none | ginkgo/v2, gomega | Ginkgo v2 and Gomega in every _test.go |  |
| deps | fakes | ok | none | tool github.com/maxbrunsfeld/counterfeiter/v6 | a tool directive for counterfeiter generates the fakes |  |
| deps | logging | ok | none | log/slog | log/slog only, in every non-test file that logs |  |
| logging | stderr-json | ok | none | slog.NewJSONHandler, never over os.Stdout | the JSON handler writes to stderr; stdout is program output |  |

0 gaps (0 tooling, 0 migration), 30 ok, 0 skipped
