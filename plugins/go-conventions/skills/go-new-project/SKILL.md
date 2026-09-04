---
name: go-new-project
description: Use when starting a new Go project, binary, CLI, service, or module; asked to scaffold a Go repo or lay out a new Go project; or running `go mod init` for something new.
---

# Scaffold a new Go project

One project per invocation. Every rule this procedure applies lives in the canon
skill (`${CLAUDE_PLUGIN_ROOT}/skills/go-conventions/SKILL.md`, its `references/`
and `templates/`), not here: this file sequences them and owns the template
mapping table. `references/<file>` and `templates/<file>` are under that skill.
This skill works in a local tree only: it creates no remote, pushes nothing,
and opens no pull request — `github-conventions:github-new-repo` does, in step 6.

## Inputs

Ask for each of these that was not given; never assume one:

- **Module path** — default `github.com/<gh login>/<name>`.
- **Binary name** — default the module path's last element.
- **One-line description** of what the binary does.
- **Environment prefix** — defaulted from the binary name.

`{{OWNER}}` and `{{REPO}}` are the module path's last two elements for a
`github.com/…` path and asked for otherwise. What each placeholder is filled
with is `references/layout.md`, "Template placeholders"; this file owns only
which template takes which, and where each lands.

Scaffold into a directory whose `main` carries no commits. `github-new-repo`
covers a tree that starts from `git init` and hands back one whose `main`
already has history rather than rewriting it (github-conventions'
`references/new-repo.md`, "Creation"), so say so before writing, not after.

## Procedure

1. **Tree, then module.** `git init -b main` first — step 5's gates and step
   6's hand-off both need exactly this tree. Then `go mod init <module>`,
   rewrite the directive to `go 1.27`, and delete any `toolchain` line:
   `go mod init` writes it one minor low (`references/toolchain.md`, "Go version").

2. **Render** each template below from
   `${CLAUDE_PLUGIN_ROOT}/skills/go-conventions/templates/` to its destination,
   placeholders filled from the inputs. This table is the one home of that
   mapping.

   | Template | Lands at | Placeholders |
   | --- | --- | --- |
   | `main.go` | `cmd/<binary>/main.go` | `{{MODULE}}`, `{{BINARY}}`, `{{DESCRIPTION}}` |
   | `root.go` | `internal/cli/root.go` | `{{MODULE}}`, `{{BINARY}}`, `{{DESCRIPTION}}`, `{{ENV_PREFIX}}` |
   | `version.go` | `internal/version/version.go` | none |
   | `suite_test.go` | `internal/cli/cli_suite_test.go` | `{{PACKAGE}}`, `{{PACKAGE_TITLE}}` |
   | `.golangci.yml` | `.golangci.yml` | `{{MODULE}}` |
   | `Makefile` | `Makefile` | `{{MODULES}}` |
   | `.goreleaser.yaml` | `.goreleaser.yaml` | `{{BINARY}}`, `{{OWNER}}`, `{{REPO}}` |
   | `ci.yml` | `.github/workflows/ci.yml` | none |
   | `release.yml` | `.github/workflows/release.yml` | `{{BINARY}}`, `{{OWNER}}`, `{{REPO}}` |
   | `.gitignore` | `.gitignore` | none |
   | `CLAUDE-pointer.md` | `CLAUDE.md` | `{{MODULE}}`, `{{BINARY}}`, `{{ENV_PREFIX}}` |
   | `dependabot-gomod.yml` | `.github/dependabot.yml` | none |

   What each placeholder is filled with, and which `{{ … }}` token is not a
   placeholder at all, is `references/layout.md`, "Template placeholders".
   `CLAUDE.md` is a new file here; where one exists already the block goes at
   its end (`references/layout.md`, "Tree", owns its markers).
   `dependabot-gomod.yml` is one entry, not a file: how it lands with or
   without an existing `.github/dependabot.yml` is `references/ci.md`,
   "Dependabot". The two workflows land unpinned — the hand-off's `pinact` and
   `actionlint` pass covers them (github-conventions' `references/workflows.md`, "Pinning").

3. **Dependencies**, then tidy:

   ```sh
   go get github.com/spf13/cobra@v1.10.2 github.com/spf13/viper@v1.21.0
   go get github.com/onsi/ginkgo/v2@v2.32.1 github.com/onsi/gomega@v1.43.0
   go get -tool github.com/maxbrunsfeld/counterfeiter/v6@v6.12.2
   go mod tidy
   ```

   Every pin is checkable in this repository: four in
   `plugins/go-conventions/tools/go.mod`, viper's in the compliant audit
   fixture, `plugins/go-conventions/tools/internal/audit/testdata/compliant/go.mod`.

4. **A first spec**, `internal/cli/root_test.go` in package `cli_test`: one
   `Describe` that calls `cli.Run` with `--name` and a `bytes.Buffer` for
   stdout and expects the line `root.go` writes there, so the suite is not
   empty. Its shape is `references/testing.md`, "Test quality".

5. **Gate — `make check` green** before the hand-off (`references/ci.md`,
   "Gates"). It needs golangci-lint on `PATH`, which `make tools` arranges
   (`references/ci.md`, "Makefile").

6. **Hand off** to `github-conventions:github-new-repo`, passing the inputs
   already known — the slug from `{{OWNER}}` and `{{REPO}}`, and
   `{{DESCRIPTION}}` — and asking the user for the ones only that skill needs.
   It lands every row of github-conventions' `references/new-repo.md`, "What
   lands where" — except the dependabot file step 2 already wrote, which gets
   its `github-actions` entry appended instead (github-conventions'
   `references/workflows.md`, "Dependabot") — then creates the empty
   repository, applies the ruleset, and opens the draft PR. Leave everything
   rendered above uncommitted: that skill's step 1 commits the whole tree, this
   scaffold's files included, on `init`. The canon depends on that plugin (`SKILL.md`, "Precedence"): if the
   Skill tool lists no `github-conventions:*` skill, stop and say to install it.

7. **Report** the module path and the tree as scaffolded, then pass through
   what step 6 reported — repository, PR, CI watcher, skips — rather than
   restating it (github-conventions' `skills/github-new-repo/SKILL.md`, step 4).
