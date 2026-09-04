---
name: github-new-repo
description: Use when creating a new GitHub repository, publishing a local project to GitHub for the first time, or asked to make a repo for this project.
---

# Create a GitHub repository

Every rule this procedure applies — visibility, the empty repository, the
ruleset, populating through a pull request, what lands where — and the
precedence ladder it runs under live in the canon skill,
`${CLAUDE_PLUGIN_ROOT}/skills/github-conventions/SKILL.md` with its
`references/` and `templates/`, and are not repeated here; this file only
sequences them. A `references/<file>` or `templates/<file>` path below is
under that skill's directory; `references/new-repo.md` owns the rules this
procedure runs.

A project with a language plugin scaffolds first: a Go project runs
`go-conventions:go-new-project`, which lays down the Go files and hands off
here with the inputs it already knows. This skill lands only the
github-conventions files and never touches a language plugin's.

## Inputs

Ask for each of these that was not given; never assume one:

- The slug, `<owner>/<name>`.
- Visibility, `--public` or `--private` (`references/new-repo.md`,
  "Visibility").
- The README's Install and Usage text, or "none yet", for `{{INSTALL}}`
  and `{{USAGE}}`.
- The CodeQL language for `{{CODEQL_LANGUAGES}}`, or that the repository
  has no language CodeQL supports (`references/security.md`, "CodeQL").

`{{DESCRIPTION}}` is drafted from the project. Every placeholder, and the
`{{CI_WORKFLOW}}` value for a repository with no CI workflow of its own,
is `references/workflows.md`, "Template placeholders".

## Procedure

1. **Prepare the tree** locally. It starts from `git init -b main` with no
   commits: `main` gets the one empty root commit,
   `git commit --allow-empty -m "chore: repository root"`, and everything
   else lands on `init` (`git switch -c init`) — `references/new-repo.md`,
   "Creation". Where `main` already has commits, that section hands the
   tree back: say so and stop before anything is created. On `init`, land
   LICENSE, README, dependabot, the workflows, and commitlint's config and script —
   every row of `references/new-repo.md`, "What lands where", at its path,
   from `${CLAUDE_PLUGIN_ROOT}/skills/github-conventions/templates/`,
   placeholders filled; a dependabot file the scaffold already wrote gets
   the `github-actions` entry per `references/workflows.md`, "Dependabot".
   Then `GITHUB_TOKEN=$(gh auth token) pinact run` and `actionlint` over
   `.github/workflows`, and commit as a logical series
   (`references/commits.md`) — everything the tree carries, including the
   files a language plugin's scaffold rendered and left uncommitted for
   this step.
2. **Show the batch — Gate: one approval.** In one message: the slug and
   visibility; the ruleset JSON, `templates/ruleset.json` verbatim; the
   branch, `init` onto `main`; the PR title and description, drafted per
   `references/pull-requests.md`, "The description". Nothing in step 3
   runs before the user approves the batch as a whole; a change to any
   item re-shows the batch.
3. **Execute**, in the canon's order (`references/new-repo.md`,
   "Creation"), stopping at the first failure:
   1. `gh repo create <owner>/<name> --public` or `--private`, and nothing
      else on the line.
   2. The ruleset: the `gh api` line of `references/new-repo.md`,
      "Creation", step 2, with `--input` at
      `${CLAUDE_PLUGIN_ROOT}/skills/github-conventions/templates/ruleset.json`.
   3. `git remote add origin https://github.com/<owner>/<name>.git` (or the
      SSH form the user's other clones use), then `git push origin main` —
      the root commit, the only push the default branch ever gets.
   4. `git push origin init`.
   5. `gh pr create --draft --assignee @me` with the approved title and
      description (`references/pull-requests.md`, "Opening a PR"). A
      failed assignment is non-fatal: the PR exists; note it, do not
      retry.
   6. Start the CI watcher in this same turn (`references/pull-requests.md`,
      "Watching CI").
4. **Report** the repository URL, the PR URL, and the watcher's status,
   plus anything skipped: the assignee, CodeQL.
