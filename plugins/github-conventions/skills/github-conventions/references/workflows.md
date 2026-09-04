# Workflows

Owns GitHub Actions hygiene — pinning and pinact, actionlint, permissions,
concurrency, timeouts, checkout credentials — the template placeholders,
the Dependabot rule, and how a workflow diff is reviewed. Runs under
`SKILL.md`'s precedence and routing. Which security workflows exist and
the constraints on them are `references/security.md`; the commitlint
workflow is `references/commits.md`; where each template lands is
`references/new-repo.md`, "What lands where".

## Pinning

- Every `uses:` names a full 40-hex commit SHA followed by a `# vX.Y.Z`
  comment naming the tag it resolves: `actions/checkout@<sha> # v7.0.1`.
  Local (`./…`) and `docker://` refs are exempt. A tag or branch ref
  (`@v4`, `@main`) is a gap; a comment disagreeing with its SHA is a
  lesser one.
- `pinact run` after any edit that adds or bumps an action, before the
  commit (`GITHUB_TOKEN=$(gh auth token) pinact run` — it resolves tags
  through the API and needs a token). In CI, the `pin-check` job of
  `templates/workflow-lint.yml` fails on any unpinned ref.
- The `# vX.Y.Z` comment is what Dependabot bumps ("Dependabot" below); a
  pin without one never updates.
- Files under `templates/` carry major tags (`@v7`) on purpose — Dependabot
  scans only `.github/workflows/`, so a pin inside a plugin would rot; the
  skill that copies one runs `pinact run` immediately after, so a target
  repository only ever holds pins. Never pin a template.

## actionlint

`actionlint` on every changed workflow before it is committed, with
`shellcheck` on `PATH` so inline `run:` scripts are linted too; fix
everything it reports. In CI, the `actionlint` job of
`templates/workflow-lint.yml` runs the same check.

## Permissions

Top-level `permissions: contents: read` on every workflow. A job that
needs more declares exactly that scope at job level — never wider, never
at the top. `write-all`, a top-level write, or a workflow with no
`permissions` block (it inherits the repository default) is a gap.

## Concurrency

Every workflow carries a top-level `concurrency` block with
`cancel-in-progress: true`, keyed to vary by ref:

```yaml
concurrency:
  group: ${{ github.workflow }}-${{ github.event.pull_request.number || github.sha }}
  cancel-in-progress: true
```

(`|| github.sha` matters only when the workflow also runs on push; a
workflow with no pull-request trigger keys on `${{ github.ref }}`.) A
constant key serializes every run. The one exception is a release
workflow: a release must never be cancelled mid-run, so it keys on a
constant group with `cancel-in-progress: false` — serialized, never
cancelled.

## Timeouts

Every job sets `timeout-minutes`. The templates' values are the scale:
5 for a linter, 15 for Scorecard, 30 for CodeQL.

## Checkout credentials

Every `actions/checkout` step sets `persist-credentials: false`.

## Template placeholders

A `{{NAME}}` token is filled when the template is copied; a target
repository never contains one.

| Placeholder | Value | Used by |
|---|---|---|
| `{{OWNER}}` | the GitHub owner, user or organization | `templates/README.md` |
| `{{REPO}}` | the repository name | `templates/README.md` |
| `{{CI_WORKFLOW}}` | file stem of the repository's primary CI workflow (`ci`, `validate`), or `workflow-lint` when it has none of its own; the README's first badge names it | `templates/README.md` |
| `{{DESCRIPTION}}` | one paragraph on what the repository is | `templates/README.md` |
| `{{INSTALL}}` | the Install section's body | `templates/README.md` |
| `{{USAGE}}` | the Usage section's body | `templates/README.md` |
| `{{CODEQL_LANGUAGES}}` | the one CodeQL language of that job (`references/security.md`, "CodeQL") | `templates/codeql.yml` |
| `{{CODEQL_BUILD_MODE}}` | the CodeQL build mode for that language (`references/security.md`, "CodeQL") | `templates/codeql.yml` |

## Dependabot

`.github/dependabot.yml` is `templates/dependabot.yml`: `version: 2`, then
an `updates:` list whose first entry is the `github-actions` ecosystem,
weekly, grouped so every action bump arrives as one PR. A language plugin
owns its own `package-ecosystem` entry: it appends the entry when the file
is already there, and writes the file carrying it when it is not, in which
case this canon's `github-actions` entry is appended to the existing
`updates:` list; this plugin writes only its own entry and leaves the
others alone. A language plugin's template is a bare `updates:` list item,
so whichever plugin writes the file first writes the `version: 2` and
`updates:` preamble above the entry — a file that starts at the list item
is not a valid Dependabot config.

## Reviewing a workflow

Triggers: a diff under `.github/workflows/**`, or a script a workflow
invokes. Injection — `pull_request_target`, event strings in `run:`,
secrets on fork PRs — is `references/security.md`, "Workflow injection".

### Linters first

Where the repository has no `workflow-lint` workflow (the `actionlint`
and `pin-check` jobs above), run both linters yourself before judging
anything (a bare `pinact run` rewrites in place; `--check` does not):

```sh
actionlint <changed workflow files>      # shellcheck on PATH lints run: blocks too
GITHUB_TOKEN=$(gh auth token) pinact run --check --verify <changed workflow files>
```

Report what they find — actionlint's diagnostics, and pinact's diff (a
tag or branch ref comes back as a `-`/`+` rewrite; a `# vX.Y.Z` comment
disagreeing with its SHA comes back as a mismatch error) — then the
judgment checks below. Where the repository has `workflow-lint`, its
output is already known: review only what the linters miss.

### Dependabot entry

The diff does not drop the `github-actions` entry from
`.github/dependabot.yml` ("Dependabot" above).

### Logic (what actionlint misses)

- **Trigger/context mismatch**: `github.event` payload fields differ by
  trigger (`pull_request` vs `push` vs `workflow_dispatch` vs `schedule`)
  — a step reading `github.event.pull_request.*` in a workflow also
  triggered by push gets empty strings, which then flow into conditions
  silently.
- **`if:` semantics**: expressions vs strings (`if: ${{ false }}` vs
  `if: "false"`), missing `always()`/`failure()` where cleanup must run,
  `cancelled()` handling; job-level vs step-level condition placement.
- **Output plumbing**: `$GITHUB_OUTPUT` written names match
  `needs.<job>.outputs.<name>` / `steps.<id>.outputs.<name>` readers; a
  renamed output with a stale reader fails silently to empty-string.
- **Concurrency groups**: "Concurrency" above.
- **`permissions:`**: "Permissions" above.
- **Cache correctness**: `actions/cache` keys include the
  lockfile/toolchain hash they guard; restore-keys don't resurrect
  poisoned/stale caches across major toolchain bumps.
- **Artifacts**: upload/download names paired; retention deliberate.

### Matrix logic

- Every matrix dimension is actually CONSUMED by the job (an unused
  dimension multiplies CI cost for nothing).
- `include`/`exclude` produce the intended combinations — enumerate them
  when non-trivial; `fail-fast` and `max-parallel` are deliberate choices
  (burn-in style matrices want `fail-fast: false`).
