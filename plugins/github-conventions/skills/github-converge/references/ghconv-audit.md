# ghconv-audit output contract

Owns what `ghconv-audit` prints and how it is read: the usage and exit
codes, the row fields, the markdown and JSON renderings, the remote
ruleset lookup, and the check inventory with the canon each check
measures. Runs under `SKILL.md`'s procedure, and under its reading of
`references/<file>` and `templates/<file>`.

`tools/cmd/ghconv-audit/` implements this specification of its output.

- Usage: `ghconv-audit [--json|--markdown] [--remote] [dir]`; `dir` defaults
  to `.`, markdown is the default rendering, `--json` and `--markdown` exclude
  each other. Exit 0 whenever the audit ran, whatever it found; exit 1 on a
  usage or I/O error (a bad flag, an unreadable `dir`), with the reason on stderr.
- Row fields: `area`, `check`, `status` (`ok`, `gap`, or `skipped`), `current`
  (what the repository has), `canon` (what it should have), `fix` (what closes
  the gap; empty unless `gap`; a `templates/…` path in it is under the canon skill).
- Markdown: the header `| Area | Check | Status | Current | Canon | Fix |`, the
  separator `| --- | --- | --- | --- | --- | --- |`, one row per check in the
  order below, a blank line, then `N gaps, M ok, K skipped`; in a cell `|` is
  escaped as `\|` and a newline becomes a space. JSON:
  `{"rows":[…],"gaps":N,"ok":M,"skipped":K}`, the cell text as written.
- Every check but the last reads the tree; nothing is executed. The five
  `workflows` checks repeat per `.yml`/`.yaml` under `.github/workflows/` by
  file name (`timeout` skips a job that calls a reusable workflow); a file that
  does not parse as YAML gets all five as gaps against "the workflow parses as YAML".
- `--remote` reads `owner/repo` from `git remote get-url origin`, lists rulesets
  with `gh api --paginate repos/<owner>/<repo>/rulesets`, and fetches each by id
  (the list carries neither conditions nor rules); a ruleset whose fetch fails
  keeps its list entry and can only report a gap, never a false pass. A lookup
  that fails outright (no `origin`, no `gh`, an API error) is a `gap` whose
  `current` begins `gh api failed:`; without `--remote` the row is `skipped`.
  The row's `fix` is the tool's rendering of the ruleset command, valid in a
  checkout with an `origin`: `{owner}/{repo}` is a placeholder `gh api` fills
  from that remote, and `--method` is `-X` spelled long; a repository with no
  remote yet uses the form in `references/new-repo.md`, "Creation", step 2.
- The checks are a subset of the canon: an `ok` row means that check
  passed, not that the owning section, cited per row, is satisfied.

| Area / check | Canon | Owner |
| --- | --- | --- |
| `license` / `present` | LICENSE exists at the repository root | `references/new-repo.md`, "LICENSE" |
| `license` / `apache-2.0` | Apache-2.0, or the license the repository already carries | `references/new-repo.md`, "LICENSE" |
| `readme` / `present` | README.md exists at the repository root | `references/new-repo.md`, "README" |
| `readme` / `scorecard-badge` | README.md carries the OpenSSF Scorecard badge | `references/security.md`, "Badge" |
| `dependabot` / `present` | .github/dependabot.yml exists | `references/workflows.md`, "Dependabot" |
| `dependabot` / `github-actions` | an updates[] entry sets package-ecosystem: github-actions | `references/workflows.md`, "Dependabot" |
| `workflows` / `pinned:<file>` | every uses: ends in a 40-hex commit SHA with a # v… comment | `references/workflows.md`, "Pinning" |
| `workflows` / `permissions:<file>` | a top-level permissions key, with no write scope | `references/workflows.md`, "Permissions" |
| `workflows` / `timeout:<file>` | every job sets timeout-minutes | `references/workflows.md`, "Timeouts" |
| `workflows` / `concurrency:<file>` | a top-level concurrency key | `references/workflows.md`, "Concurrency" |
| `workflows` / `checkout-credentials:<file>` | every actions/checkout step sets with.persist-credentials: false | `references/workflows.md`, "Checkout credentials" |
| `security` / `codeql` | a workflow runs github/codeql-action/init | `references/security.md`, "CodeQL" |
| `security` / `dependency-review` | a workflow runs actions/dependency-review-action | `references/security.md`, "Dependency review" |
| `security` / `scorecard` | a workflow runs ossf/scorecard-action | `references/security.md`, "Scorecard" |
| `workflow-lint` / `actionlint` | a workflow runs actionlint over .github/workflows | `references/workflows.md`, "actionlint" |
| `workflow-lint` / `pin-check` | a workflow fails any uses: not pinned to a 40-hex SHA | `references/workflows.md`, "Pinning" |
| `commitlint` / `config` | .commitlintrc.yml exists at the repository root | `references/commits.md`, "Conventional Commits" |
| `commitlint` / `workflow` | a workflow runs wagoid/commitlint-github-action on pull requests | `references/commits.md`, "Conventional Commits" |
| `commitlint` / `breaking-footer` | .github/tools/breaking-footer/main.go exists | `references/commits.md`, "Conventional Commits" |
| `ruleset` / `default-branch` | an active branch ruleset on ~DEFAULT_BRANCH with deletion and non_fast_forward rules | `references/new-repo.md`, "Creation" |
