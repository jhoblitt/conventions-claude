# Eval cases

Regression gates for the github-conventions plugin's own behavior: that a
subject reading this checkout's canon reviews a workflow the way
`references/workflows.md` says, creates a repository the way
`github-new-repo` says, and does not do the things the canon exists to
prevent.

`claude plugin eval` is not a documented command yet, so a case runs through
`evals/run-manual.sh` at the repository root, in two separate `claude -p`
passes — a subject and a grader:

```sh
./evals/run-manual.sh github-conventions gh-pin-check          # both passes
./evals/run-manual.sh github-conventions gh-new-repo-flow -s   # subject only
```

The passes stay separate on purpose: a session that has read
`graders/criteria.md` writes a report that satisfies it, and a session that
just wrote the canon cannot grade it honestly. The runner points the subject
at this checkout rather than the installed plugin, tells it not to read the
`graders/` directories, and tells it that a command it cannot run is a coverage gap to
record, not a result to invent.

Both cases are hermetic: no repository checkout, no network, no `gh`, no Go
toolchain, and no `allowed-tools` file — the subject has Read, Grep, and
Glob, and everything it judges is embedded in `prompt.md`. Neither case
runs `ghconv-audit`; the tool's output is pinned by the golden tests under
`tools/internal/audit/testdata/`.

| Case | Guards |
| --- | --- |
| `gh-pin-check` | A workflow review flags the `@v4` tag, the missing `permissions`, the missing `timeout-minutes`, and the checkout without `persist-credentials: false`; names `pinact run` (or the `pin-check` job) as the pin fix; leaves the `./` local action and the SHA-pinned step alone; recommends no harden-runner; and claims no edit it did not make. |
| `gh-new-repo-flow` | Asked for a repository with no visibility stated and no `gh`, the subject asks for the visibility rather than picking one, lays out the empty-repo, ruleset, `init`-branch, draft-PR flow as one batch for one approval, pushes no project file to the default branch, and claims no `gh` run. |
