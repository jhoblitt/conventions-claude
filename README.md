# conventions-claude

[![validate](https://github.com/jhoblitt/conventions-claude/actions/workflows/validate.yml/badge.svg)](https://github.com/jhoblitt/conventions-claude/actions/workflows/validate.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/jhoblitt/conventions-claude/badge)](https://scorecard.dev/viewer/?uri=github.com/jhoblitt/conventions-claude)

A [Claude Code plugin marketplace](https://code.claude.com/docs/en/plugin-marketplaces)
with two plugins that stop convention drift: **go-conventions**, the house
canon for how Go code is written, tested, logged, linted, built, and
released, and **github-conventions**, the canon for how a GitHub repository
is created, kept hygienic, and worked on through commits and pull requests.
go-conventions depends on github-conventions; installing the first installs
both.

## Install

Inside Claude Code:

```text
/plugin marketplace add jhoblitt/conventions-claude
/plugin install go-conventions@conventions-claude
```

or from a shell:

```sh
claude plugin marketplace add jhoblitt/conventions-claude
claude plugin install go-conventions@conventions-claude
```

A repository with no Go in it installs `github-conventions@conventions-claude`
on its own.

To pick up updates later:

```sh
claude plugin marketplace update conventions-claude   # refresh the index
claude plugin update go-conventions@conventions-claude  # install the new version
```

or inside Claude Code:

```text
/plugin marketplace update conventions-claude
/plugin update go-conventions@conventions-claude
/reload-plugins
```

After the shell form, run `/reload-plugins` in running sessions — a
restart also works. The marketplace step alone only refreshes the
index; it does not update the installed plugin.

## What's inside

### github-conventions

| Skill | What it does |
| --- | --- |
| `github-conventions` | The canon. Loads when you create a repository, edit a workflow or `dependabot.yml`, commit, rebase, open or update a pull request, watch CI, or post a GitHub comment. Owns the precedence ladder both plugins share. |
| `/github-conventions:github-converge` | Audits an existing repository against the canon and applies the file half on a branch. |
| `/github-conventions:github-new-repo` | Creates a repository the house way: empty, ruleset-protected, populated by a draft pull request. |

### go-conventions

| Skill | What it does |
| --- | --- |
| `go-conventions` | The canon. Loads when you write or refactor Go, edit `go.mod`, a lint config, a Makefile, a goreleaser config or a Go workflow, or choose a Go library. |
| `/go-conventions:go-new-project` | Scaffolds a new module: cobra and viper, `log/slog`, Ginkgo, the house lint tier, goreleaser with ko, then hands off to `github-new-repo`. |
| `/go-conventions:go-converge` | Audits an existing Go repository, applies the tooling half, and migrates one named area at a time. |
| `/go-conventions:go-review` | Reviews a working tree, a branch, or a pull request against the review canon. |

Two hooks come with go-conventions: after a `.go` file is written or edited it
is formatted with gofumpt (or gofmt), and a session that starts in a Go module
is told the module path and to load the canon. Both cost nothing anywhere else
— the launcher checks the path and the presence of `go.mod` before it builds
anything, and fails open.

### Skill workflows

`github-converge`:

```mermaid
flowchart TD
  A["ghconv-audit --markdown<br/>print the table as rendered"] --> B["Branch github-conventions/converge"]
  B --> C["Next gap, in table order"]
  C --> D{"File exists?"}
  D -- no --> E["Copy the template, fill placeholders,<br/>pinact, actionlint, commit the area"]
  D -- yes --> F{"User approves the diff"}
  F --> E
  E --> C
  C --> G["GitHub-side fixes: show the exact command"]
  G --> H{"User approves the command"}
  H --> I["Re-run the audit, report<br/>what landed and what waits"]
```

`github-new-repo`:

```mermaid
flowchart TD
  A["Collect inputs:<br/>slug, visibility, README text,<br/>CodeQL language"] --> B["Prepare the tree on init<br/>from git init -b main"]
  B --> C{"main already has commits?"}
  C -- yes --> D["Hand back to the user, create nothing"]
  C -- no --> E["Show the whole batch"]
  E --> F{"User approves once"}
  F --> G["gh repo create, ruleset,<br/>push main, push init,<br/>draft PR, CI watcher"]
  G --> H["Report the repository and PR URLs"]
```

`go-new-project`:

```mermaid
flowchart TD
  A["Collect inputs:<br/>module, binary, description,<br/>env prefix"] --> B["git init -b main,<br/>go mod init, set go 1.27"]
  B --> C["Render every template<br/>to its destination"]
  C --> D["Pin the house dependencies,<br/>go mod tidy"]
  D --> E["Write the first spec"]
  E --> F2["Write dependabot with the gomod entry"]
  F2 --> G2{"make check green"}
  G2 --> H2["Hand off"]
  H2 -.->|github-conventions:github-new-repo| I2["Repository, ruleset, draft PR"]
  I2 --> J2["Report"]
```

`go-converge`:

```mermaid
flowchart TD
  A["goconv-audit --markdown<br/>print the table as rendered"] --> B["Repository hygiene first"]
  B -.->|github-conventions:github-converge| C["Hygiene files land"]
  C --> D["Stay on the branch github-converge cut<br/>from the default branch"]
  D --> E["Tooling phase: one commit per area,<br/>diff gate on any file that exists"]
  E --> F{"User names a migration area"}
  F --> G["layout"]
  G --> H["cli + logging + version"]
  H --> I["errors"]
  I --> J["testing"]
  G -.->|go-migrator| K["Area migrated, tree left uncommitted"]
  H -.->|go-migrator| K
  I -.->|go-migrator| K
  J -.->|go-migrator| K
  K --> L["Regenerate the lint config,<br/>make check, commit the area"]
  L --> M["Re-run the audit, report"]
```

`go-review`:

```mermaid
flowchart TD
  A{"Target?"} -- working tree --> B["git diff plus untracked files"]
  A -- branch or range --> C["git diff base...HEAD"]
  A -- PR number --> D["Fetch the head into a throwaway worktree,<br/>lint config from the base branch,<br/>no go test on a fork head"]
  B --> E["Fence everything the target authored<br/>under a marker drawn fresh per review"]
  C --> E
  D --> E
  E --> F{"Diff over 1500 lines?"}
  F -- yes --> G[["One go-reviewer per package group,<br/>disjoint file lists, at most 4"]]
  G --> H["Merge by file:line, one finding per site"]
  F -- no --> I["Route the changed files to references and read them"]
  I --> J["Oracles once: build, vet, go fix -diff, golangci-lint"]
  J --> K["Judgment pass over the canon's classes"]
  K --> L{"Candidate argued both ways?"}
  L -- no --> M["Refute it inline"]
  L -- yes --> N["One go-reviewer, given that finding<br/>and the code it names"]
  M --> O["Report: findings, oracles, verdict"]
  N --> O
  H --> O
```

### How the skills fit together

```mermaid
flowchart TD
  GOC["go-conventions<br/>the Go canon"]
  GHC["github-conventions<br/>the repository canon"]
  GOC -.->|precedence ladder, workflow hygiene, commits, pull requests| GHC
  NP["/go-conventions:go-new-project"] --> GOC
  NP -.->|hands off| NR["/github-conventions:github-new-repo"]
  NR --> GHC
  CV["/go-conventions:go-converge"] --> GOC
  CV -.->|delegates the hygiene half| GCV["/github-conventions:github-converge"]
  GCV --> GHC
  RV["/go-conventions:go-review"] --> GOC
  RV -.->|fan-out over 1500 lines| RA["go-reviewer"]
  CV -.->|one area per dispatch| MA["go-migrator"]
```

## Example prompts

- "Write a Go CLI that reconciles two inventories" — the canon loads on its own.
- "/go-conventions:go-new-project" — scaffold a new module and its repository.
- "/go-conventions:go-converge" — see what an existing repository is missing, then migrate an area at a time.
- "/go-conventions:go-review 1234" — review a pull request against the review canon.
- "/github-conventions:github-converge" — pin the actions, add the security workflows, protect the default branch.

## Scope

The plugins report and propose; they do not act outside the working tree
without saying so first. Converge creates a file that is absent, and shows a
diff before touching one that exists. Anything that reaches GitHub — a
repository, a ruleset, a push, a pull request, a comment — happens only as a
command you have seen and approved. A review reports; it never edits.

## Development

Validate after changes:

```sh
claude plugin validate --strict .
```

Content changes land via PR. Commit messages follow
[Conventional Commits](https://www.conventionalcommits.org/) — commitlint
enforces this on every PR — and releasing is automated: on each merge to
`main`, [semantic-release](https://github.com/semantic-release/semantic-release)
derives the next version from the commit types in the changeset (`fix:`,
`docs:`, `refactor:`, `perf:` → patch; `feat:` → minor; a breaking change →
major; other types cut no release), writes it into both plugin manifests and
`CHANGELOG.md`, tags, and publishes a GitHub release. Never bump a plugin
version in a PR — the release commit owns that field.

## License

[Apache-2.0](LICENSE)
