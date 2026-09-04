| Area | Check | Status | Phase | Current | Canon | Fix |
| --- | --- | --- | --- | --- | --- | --- |
| toolchain | go-directive | ok | none | go 1.27 | go 1.27, the minor only, and no toolchain line |  |
| lint | config | gap | tooling | v1 schema: no version key, run.deadline, linters.disable-all | a .golangci.yml at the root, golangci-lint v2 (version: "2") | regenerate: goconv-audit --emit-golangci |
| lint | linters | skipped | none | no v2 lint config | every linter templates/.golangci.yml enables |  |
| lint | formatters | skipped | none | no v2 lint config | the formatters templates/.golangci.yml enables |  |
| lint | max-issues | skipped | none | no v2 lint config | issues.max-issues-per-linter and max-same-issues are both 0 |  |
| makefile | present | ok | none | Makefile | a Makefile at the repository root |  |
| makefile | targets | gap | tooling | missing: help, print-golangci-version, tools, generate-check, fmt-check, fix, fix-check, tidy, tidy-check, check | every target templates/Makefile defines | copy templates/Makefile |
| makefile | golangci-version | gap | tooling | GOLANGCI_LINT_VERSION is not defined | GOLANGCI_LINT_VERSION is defined in the Makefile | copy templates/Makefile |
| ci | setup-go | gap | tooling | no workflow sets go-version-file: go.mod | a workflow sets Go up with go-version-file: go.mod | copy templates/ci.yml to .github/workflows/ |
| ci | race | gap | tooling | no run: block runs go test -race or make test | a workflow runs the suite with the race detector | copy templates/ci.yml to .github/workflows/ |
| ci | lint | gap | tooling | no step uses golangci/golangci-lint-action | a workflow runs golangci/golangci-lint-action | copy templates/ci.yml to .github/workflows/ |
| ci | checks | gap | tooling | no run: block runs fix-check and tidy-check | a workflow runs the go fix and go mod tidy checks | copy templates/ci.yml to .github/workflows/ |
| ci | govulncheck | gap | tooling | no step or run: block runs govulncheck | a workflow runs govulncheck over ./... | copy templates/ci.yml to .github/workflows/ |
| release | goreleaser | gap | tooling | no .goreleaser.yaml | .goreleaser.yaml at the root, version: 2 | copy templates/.goreleaser.yaml |
| release | kos | gap | tooling | no readable .goreleaser.yaml | a kos block builds the image from the same checkout | copy templates/.goreleaser.yaml |
| release | sign-sbom | gap | tooling | no readable .goreleaser.yaml | signs and sboms blocks: cosign over the checksums, syft over the archives | copy templates/.goreleaser.yaml |
| release | workflow | gap | tooling | no workflow runs goreleaser on a v* tag | a workflow on v* tags runs goreleaser/goreleaser-action | copy templates/release.yml to .github/workflows/ |
| release | no-ldflags-x | ok | none | no -X ldflags | no -X ldflags anywhere; the version is the build stamp |  |
| dependabot | gomod | gap | tooling | no .github/dependabot.yml | .github/dependabot.yml carries a gomod updates entry | append templates/dependabot-gomod.yml under the existing updates: list |
| gitignore | present | ok | none | .gitignore | .gitignore at the repository root |  |
| claude | pointer | gap | tooling | no CLAUDE.md | CLAUDE.md carries the go-conventions pointer block | insert templates/CLAUDE-pointer.md between its markers |
| version | package | gap | tooling | no internal/version | internal/version reads debug.ReadBuildInfo | copy templates/version.go to internal/version/ |
| layout | cmd | skipped | none | controller-runtime | every package main lives in cmd/<name> |  |
| layout | pkg-dir | ok | none | no pkg/ | no pkg/ in a module that ships a binary |  |
| deps | cli | skipped | none | controller-runtime | cobra builds every command tree |  |
| deps | config | skipped | none | controller-runtime | viper resolves flags, environment, and config file |  |
| deps | testing | skipped | none | no test files | Ginkgo v2 and Gomega in every _test.go |  |
| deps | fakes | skipped | none | no test files | a tool directive for counterfeiter generates the fakes |  |
| deps | logging | skipped | none | controller-runtime | log/slog only, in every non-test file that logs |  |
| logging | stderr-json | skipped | none | controller-runtime | the JSON handler writes to stderr; stdout is program output |  |

15 gaps (15 tooling, 0 migration), 5 ok, 10 skipped
