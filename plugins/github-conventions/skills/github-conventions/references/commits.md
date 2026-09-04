# Commits

Owns commit messages — Conventional Commits, commitlint, the
breaking-change footer — branch history as a logical series, the rewrite
proof, and what a message says. Runs under `SKILL.md`'s precedence and
routing. The PR description is a different artifact with its own shape:
`references/pull-requests.md`, "The description".

## Conventional Commits

- Every subject is `type(scope)?: description`, the types
  `@commitlint/config-conventional` allows. commitlint enforces it on
  every pull request: `templates/commitlint.yml` runs
  `wagoid/commitlint-github-action` against `templates/.commitlintrc.yml`
  (config-conventional with `body-max-line-length` off; its comment says
  why). Where each lands: `references/new-repo.md`, "What lands where".
- A `BREAKING CHANGE:` or `BREAKING-CHANGE:` footer requires `!` in the
  subject (`feat!:`, `fix(api)!:`). The commitlint workflow runs
  `templates/breaking-footer/main.go` with `go run` to reject a footer the
  subject never declared; the program's package doc says why commitlint
  cannot.

## Branch history

A branch is a logical series that reads as though written correctly the
first time: each commit one coherent unit, the split reflecting the final
change rather than the path to it, no commit existing to correct an
earlier one on the same branch. Before any push that opens or updates a
PR — the first push and every repush — check the history and fix it.

Tells: `fixup!`/`squash!` subjects; "address review", "review feedback",
"fix typo", "oops", "wip"; a commit whose whole content patches files only
an earlier commit on the branch touched.

- Mechanical — do it, then report: a commit whose entire content belongs
  to exactly one earlier commit on the branch is squashed into it.
- Judgment — propose and wait: splitting a commit, regrouping across
  concerns, reordering, a fixup spanning two parents, or any case with
  more than one defensible grouping. Show which commits become which and
  why, then stop.

A surviving commit's message describes the final state; that it absorbed
a fix is process ("What a message says" below).

## The rewrite proof

After any history rewrite, before pushing:

1. `git diff <old-head> HEAD` is empty — say so. A non-empty diff means
   the rewrite went wrong: stop and report, never push.
2. Every `Signed-off-by:` survived.
3. Push with `--force-with-lease`.

## What a message says

A commit message documents what changed and why, for a future reader of
the history — never how the change was produced. Leave out process:
sanity checks that found nothing, that the PR was opened as a draft,
labels added, "rebased onto main", which remote it was pushed from. A
finding that changed the diff is part of the change and stays. Whose
voice an AI-assistance disclosure is in, and where the agent marker does
not go, is `references/pull-requests.md`, "Comments".
