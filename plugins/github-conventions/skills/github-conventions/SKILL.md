---
name: github-conventions
description: Use when creating a GitHub repository; editing a workflow under .github/workflows/, .github/dependabot.yml, or a repository ruleset; committing, rebasing, squashing, or otherwise rewriting branch history; opening or updating a pull request; watching or retrying CI after a push; or before posting or replying to any GitHub comment, review, or issue.
---

# GitHub repository conventions

The canon for how a GitHub repository is created, kept hygienic, and worked
on through commits and pull requests — the rules a language plugin
(`go-conventions`) and a repository with no language plugin share. It is
consulted, not run: `github-converge` audits and converges an existing
repository against it, `github-new-repo` creates one from it. A path of the
form `references/<file>` or `templates/<file>` is relative to this skill's
directory.

## Precedence

On any conflict, the higher rung wins:

1. The user's own instructions — a global or repository `CLAUDE.md` — or
   a project-specific conventions plugin such as `rook-maintainer` for
   `github.com/rook/*`.
2. Any rung a language plugin adds between that and this canon, stated in
   its own `SKILL.md` (`go-conventions` adds one).
3. This canon.

This section is the ladder's one home; a plugin that shares it points here.

## Reference routing

Read a reference when the work touches its trigger; skip the rest. The rules
below this table apply to every trigger.

| Doing this | Read |
|---|---|
| creating a repository, choosing its visibility, applying its ruleset, or writing its first files; auditing or converging a repository's license or README | `references/new-repo.md` |
| adding or editing a workflow under `.github/workflows/`, pinning an action, editing `.github/dependabot.yml`, or filling a template placeholder | `references/workflows.md` |
| reviewing a workflow diff | `references/workflows.md`, "Reviewing a workflow" |
| adding or changing CodeQL, dependency review, Scorecard, or the Scorecard badge; judging a workflow for injection | `references/security.md` |
| writing or fixing a commit message, amending, squashing, rebasing, or otherwise rewriting a branch's history, or diagnosing a commitlint failure | `references/commits.md` |
| opening or updating a PR, writing its description, or running a multi-PR campaign | `references/pull-requests.md` |
| watching CI after a push, or retrying a failed job | `references/pull-requests.md`, "Watching CI" |
| about to post or reply to a GitHub comment, review, or issue | `references/pull-requests.md`, "Comments" |

## Always

- Every action `uses:` is pinned to a full commit SHA before it is
  committed — `references/workflows.md`, "Pinning".
- A changed workflow passes actionlint, with shellcheck, before it is
  committed — `references/workflows.md`, "actionlint".
- Workflow permissions are least privilege — `references/workflows.md`,
  "Permissions".
- A pull request opens as a draft, assigned to its author —
  `references/pull-requests.md`, "Opening a PR".
- No comment, review, or reply is posted without an explicit instruction
  for that post, and every post made on the user's behalf opens with the
  agent marker — `references/pull-requests.md`, "Comments".

## Scripts

```sh
bash "${CLAUDE_PLUGIN_ROOT}/tools/run.sh" ghconv-audit [--json|--markdown] [--remote] [dir]
```

Audits a repository against this canon and prints the gap table. The
launcher fails loud: a non-zero exit is a real failure, never an empty
result. The output contract — rows, statuses, the summary line — is owned
by `github-converge` (`skills/github-converge/SKILL.md`), not defined here.
