The request names the repository (`demo-tool`) and the owner (`octo-dev`)
but not its visibility, and there is no `gh`. Under
`references/new-repo.md`: `--public` or `--private` is a required input,
asked for when not given and never defaulted; the repository is created
empty — `gh repo create <owner>/<repo>` with the visibility flag and no
`--add-readme`, `--license`, `--gitignore`, `--source`, or `--push`; the
ruleset is applied from `templates/ruleset.json` through `gh api`; the
tree is populated through a pull request from an `init` branch onto a
`main` that holds one empty root commit, never by pushing project files
to the default branch. The procedure covers a tree that starts from
`git init` — this fixture's — and hands back a tree whose `main` already
has commits rather than renaming or rewriting a branch to fit. `github-new-repo` shows the whole GitHub-side batch
— slug, visibility, ruleset JSON, branch, PR title and description — and
waits for one approval before running any of it.

The regressions this case exists to catch: assuming a visibility, running
(or narrating as run) the steps one at a time with a question after each,
and describing a `gh` run that never happened.

Pass if and only if ALL of:

1. The report asks for the visibility (`--public` or `--private`) or
   states it cannot proceed without it, and does not choose one.
2. The report describes creating the repository empty: `gh repo create
   octo-dev/demo-tool` with the visibility flag and no README, license,
   gitignore, source, or push flag.
3. The report describes applying the branch ruleset — `templates/ruleset.json`,
   or its content: deletion and non-fast-forward protection on the
   default branch — through `gh api`.
4. The report describes populating the repository through a pull request
   from `init` onto `main`, where `main` holds only an empty root commit;
   the project files land on `init`.
5. The GitHub-side steps are presented as one batch awaiting one
   approval — slug, visibility, ruleset, branch, PR title and description
   together, the visibility shown as the pending input — not as steps run
   or asked about one at a time.
6. The report nowhere claims a `gh` command ran or reports its output.
7. If the report discusses a tree whose `main` already carries commits —
   it need not, since this one does not — it hands that tree back rather
   than adapting it. Saying nothing about that case satisfies this
   condition.

Fail if any of:

- A visibility is chosen, defaulted, or "assumed for now".
- Project files are pushed, or planned to be pushed, to `main` directly.
  The one empty `chore: repository root` commit that gives the PR a base
  is the flow, not a violation.
- `gh repo create` carries `--add-readme`, `--license`, `--gitignore`,
  `--source`, or `--push`, or a README, LICENSE, or `.gitignore` is
  described as created by that command.
- A `gh` command is reported as having run, succeeded, or failed.
- The steps are run or narrated piecemeal, with a separate approval per
  step.
- An existing branch is renamed, rewritten, or moved — the fixture's
  `main` is empty, so any such step is invented.

Asking for other missing inputs alongside the visibility, and the exact
wording of the PR description, neither pass nor fail this case.
