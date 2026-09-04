# Security workflows

Owns CodeQL, dependency review, Scorecard, the Scorecard badge, and
workflow injection. Runs under `SKILL.md`'s precedence and routing.
Pinning, permissions, concurrency, timeouts, and checkout credentials are
`references/workflows.md`; where each template lands is
`references/new-repo.md`, "What lands where".

## CodeQL

`templates/codeql.yml`, advanced setup: on pull requests, pushes to
`main`, and a weekly schedule. Default setup stays off in the repository's
settings — it rejects SARIF from an advanced-setup workflow.

- `{{CODEQL_LANGUAGES}}` is one language; a second language is a second
  job, never a comma list in one.
- `{{CODEQL_BUILD_MODE}}` follows the language: `none` for interpreted
  languages and for C/C++, C#, Java, and Rust; `autobuild` for Go;
  `manual` for a compiled language with a custom build.
- `security-events: write` on the job is the one escalation.
- A repository with no language CodeQL supports lands no `codeql.yml`.

Uploading results needs code scanning, which a private repository has only
with GitHub Advanced Security, so the job carries the same
`if: github.event.repository.private == false` guard as dependency review
below: the analysis would otherwise succeed and only the upload fail.

## Dependency review

`templates/dependency-review.yml`: `actions/dependency-review-action` on
every pull request, `contents: read` only. The action needs the dependency
graph, which a private repository has only with GitHub Advanced Security, so
the job carries `if: github.event.repository.private == false`: without the
guard the step errors on such a repository rather than passing, and the
check reports red for a feature that is simply unavailable.

## Scorecard

`templates/scorecard.yml`: `ossf/scorecard-action` weekly, on push to
`main`, and on `branch_protection_rule`, with `publish_results: true`.
scorecard-action v2 rejects a workflow outside its constraints, so the
template is copied unedited apart from placeholder fill and pinning, and
any later edit keeps to:

- no `env:` or `defaults:` at workflow or job level;
- write permissions only on the scorecard job, exactly
  `security-events: write` and `id-token: write`;
- steps limited to `actions/checkout`, `ossf/scorecard-action`,
  `actions/upload-artifact`, and `github/codeql-action/upload-sarif`.

`step-security/harden-runner` is the one other step v2 permits; it was
considered and not adopted.

## Badge

`templates/README.md` carries the Scorecard badge under the title, beside
the CI badge:
`[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/{{OWNER}}/{{REPO}}/badge)](https://scorecard.dev/viewer/?uri=github.com/{{OWNER}}/{{REPO}})`.
It renders once the first Scorecard run has published.

## Workflow injection

- `pull_request_target` combined with a checkout of the PR head runs
  untrusted code with a write token — never.
- A `${{ github.event.* }}` string (title, body, branch name) is never
  interpolated into `run:`; pass it through `env:` and quote the variable.
- Secrets are not reachable from a fork's pull request, and a workflow is
  never restructured to make them so.
