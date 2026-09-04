# New repository

Owns creating a repository: the visibility input, the `gh repo create` form,
the ruleset, populating through a pull request, LICENSE, the README
skeleton, and which template lands where. Runs under `SKILL.md`'s
precedence and routing. `github-new-repo` is the skill that executes it.
The contents of the files it lands are owned elsewhere: workflows,
Dependabot, and placeholders by `references/workflows.md`; CodeQL,
dependency review, and Scorecard by `references/security.md`; commitlint
by `references/commits.md`.

## Visibility

`--public` or `--private` is a required input. Ask when it was not given;
never default it.

## Creation

1. `gh repo create <owner>/<repo> --public|--private` and nothing else: no
   `--add-readme`, `--license`, or `--gitignore`, no `--source`/`--push`.
   The repository starts empty.
2. Apply the ruleset, naming the repository explicitly — right after
   creation the cwd may have no remote:

   ```sh
   gh api -X POST repos/<owner>/<repo>/rulesets --input templates/ruleset.json
   ```

   `templates/ruleset.json` targets `~DEFAULT_BRANCH`, enforcement
   `active`, rules `deletion` and `non_fast_forward`.
3. Populate through a pull request, never a push to the default branch.
   The default branch is `main`. In a local `git init` tree whose
   `origin` is the new repository, it exists first as one empty root
   commit — `git commit --allow-empty -m "chore: repository root"` on
   `main`, pushed before any other branch — so the PR has a base; every
   file below lands on an `init` branch, and the PR opens per
   `references/pull-requests.md`, "Opening a PR". A tree whose `main`
   already carries commits is outside this canon and is handed back as it
   stands: nothing here renames a branch or rewrites history to manufacture
   that root commit. A converge target whose default branch is not `main`
   adjusts the `branches:` key of every copied workflow.

## What lands where

| Template | Lands at |
|---|---|
| `templates/LICENSE` | `LICENSE` |
| `templates/README.md` | `README.md` |
| `templates/dependabot.yml` | `.github/dependabot.yml` |
| `templates/workflow-lint.yml` | `.github/workflows/workflow-lint.yml` |
| `templates/codeql.yml` | `.github/workflows/codeql.yml`, when a supported language is present (`references/security.md`, "CodeQL") |
| `templates/dependency-review.yml` | `.github/workflows/dependency-review.yml` |
| `templates/scorecard.yml` | `.github/workflows/scorecard.yml` |
| `templates/commitlint.yml` | `.github/workflows/commitlint.yml` |
| `templates/.commitlintrc.yml` | `.commitlintrc.yml` |
| `templates/check-breaking-footer.sh` | `.github/scripts/check-breaking-footer.sh`, executable |
| `templates/ruleset.json` | the repository's rulesets, via `gh api` (step 2) — never the tree |

Placeholders are filled on copy (`references/workflows.md`, "Template
placeholders"); every copied workflow is pinned right after
(`references/workflows.md`, "Pinning"); `branches: [main]` in the copied
workflows is the default branch (step 3 above).

## LICENSE

Apache-2.0, `templates/LICENSE` verbatim. A repository that already has a
different license keeps it: this canon never relicenses, and only a
missing `LICENSE` is a gap.

## README

`templates/README.md` is the skeleton: title, CI badge, Scorecard badge
(`references/security.md`, "Badge"), description, Install, Usage,
Development — a rendering of the commit and pinning rules for a
contributor, not their home — and License. A repository with a README
keeps its own; only a missing README, or a README without the Scorecard
badge, is a gap.
