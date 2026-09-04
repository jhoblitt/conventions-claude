# Layout

Owns the repository and package layout, the shape of `main`, package naming,
interfaces, the version package, and the placeholders a template is rendered
with. Runs under `SKILL.md`'s precedence and routing. A kubebuilder module
keeps its scaffold (`references/kubernetes.md`, "Layout").

## Tree

- `cmd/<bin>/main.go` per binary, thin ("main" below); cobra commands in
  `internal/cli` (`references/cli.md`); logic in `internal/<pkg>`, one
  package per concern.
- No `pkg/` unless the module is a library other modules import; a
  binary's code is `internal/`.
- No `util`, `common`, or `helpers` package: a package is named for what it
  provides.
- Package names are short, lowercase, one word, no underscores or
  mixedCaps; the directory name matches.
- `templates/.gitignore` — `/bin/`, `/dist/`, `coverage.out` — is the
  repository's, merged into an existing one.
- The repository's `CLAUDE.md` carries `templates/CLAUDE-pointer.md`'s block
  between its `go-conventions:begin` and `go-conventions:end` markers,
  replaced in place, never duplicated.

## main

`templates/main.go`: `main` is `os.Exit(run())`; `run` opens a
`signal.NotifyContext` for `os.Interrupt` and `SIGTERM`, calls
`cli.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr)`, prints a
returned error once to stderr as `<bin>: <err>`, and returns 1. Nothing else
lives in `main`: no flag parsing, no logger setup, no logic.

## Interfaces

- Small, and declared at the consumer, not beside the implementation; one
  or two methods, named for what it does.
- Accept interfaces, return structs: a function takes the narrowest
  interface it needs and returns the concrete type.
- A test double follows the consumer (`references/testing.md`, "Test doubles").

## Version

`templates/version.go`, package `internal/version`:

- `Read() Info` reads `debug.ReadBuildInfo` — `Info{Version, Revision,
  Time, Modified, GoVersion}`, zero-valued outside a module build.
- `Info.String()` renders the fields it has: the version, then the
  revision in parentheses truncated to twelve characters and suffixed
  `, dirty` when `Modified`, then `built <time>`, then the Go version —
  each segment omitted when its field is empty.
- `String()` feeds cobra's `Version` field, which wires `--version`
  (`references/cli.md`).
- No `-X` ldflags anywhere — not in a Makefile, a workflow, or
  `.goreleaser.yaml` (`references/release.md`). Go 1.24 and later stamp
  `Main.Version` from the git tag: the tag for a tagged commit, a
  pseudo-version otherwise, `+dirty` when the tree is modified; `-trimpath`
  does not disable it; `go install pkg@vX` records the version it fetched.
- The stamp needs `.git` at build time: a build without it reports
  `(devel)` (`references/release.md`, "Images").

## Template placeholders

A `{{NAME}}` token in a template under `templates/` is filled when the file
is copied; a rendered repository holds no `{{NAME}}` token. `go-new-project` fills them at
scaffold time and `go-converge` when it lands a template
(`skills/go-new-project/SKILL.md` owns which template takes which).
github-conventions' own placeholders are its
`references/workflows.md`, "Template placeholders".

| Placeholder | Filled with |
|---|---|
| `{{MODULE}}` | the module path from `go.mod` |
| `{{MODULES}}` | the module directories relative to the Makefile, space separated; `.` for a single-module repository. Assigned with `?=`, so a caller can override it |
| `{{BINARY}}` | the binary name — the `cmd/<bin>` directory, the goreleaser build id, and the archive name |
| `{{ENV_PREFIX}}` | the environment-variable prefix, the binary name upper-cased with `-` mapped to `_` (`references/cli.md`, "Configuration") |
| `{{OWNER}}` | the owner half of the `<owner>/<repo>` pair the release workflow's fork guard and the ko image repository are built from |
| `{{REPO}}` | the repository half of that same pair |
| `{{DESCRIPTION}}` | one line on what the binary does |
| `{{PACKAGE}}` | the package a suite file belongs to, as written in `package <name>_test` |
| `{{PACKAGE_TITLE}}` | that package name title-cased, so `Test{{PACKAGE_TITLE}}` is a valid exported test function |

`{{ .Version }}` in `templates/.goreleaser.yaml`'s ko tags is not a
placeholder: it is goreleaser's own template, evaluated at release time, and
it survives rendering verbatim — as do the `${{ … }}` expressions GitHub
Actions evaluates in a workflow. Neither is `{{NAME}}`-shaped, and neither
is filled at copy time.
