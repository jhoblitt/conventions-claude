The fixture is one workflow with three `uses:` steps: `actions/checkout`
pinned to a 40-hex SHA with a `# v7.0.1` comment, `actions/setup-node@v4`
(a tag), and `./.github/actions/lint` (a local action). It has no
top-level `permissions`, no `timeout-minutes` on its one job, and a
checkout step that does not set `persist-credentials: false`. Its
`concurrency` block is correct. Under `references/workflows.md`: a tag or
branch ref is a gap and a local `./` ref is exempt; the fix for an unpinned
ref is `pinact run`, and in CI the `pin-check` job of `workflow-lint`;
every workflow carries top-level `permissions: contents: read`; every job
sets `timeout-minutes`; every `actions/checkout` step sets
`persist-credentials: false`. `references/security.md` records
`step-security/harden-runner` as considered and not adopted.

The regressions this case exists to catch: flagging the local action as
unpinned, inventing a defect in the correctly pinned step, recommending
harden-runner, and claiming to have applied a fix the subject could not
make.

Pass if and only if ALL of:

1. The report flags `actions/setup-node@v4` as not pinned to a commit SHA.
2. The report does NOT flag `./.github/actions/lint` as unpinned or as
   needing a pin.
3. The report flags the missing top-level `permissions` block.
4. The report flags the missing `timeout-minutes` on the `test` job.
5. The report flags the checkout step for not setting
   `persist-credentials: false`.
6. The pin finding names `pinact run`, or the `pin-check` job of the
   `workflow-lint` workflow, as the fix.
7. The report makes no finding against the checkout step's `uses:` ref —
   not its SHA, not its `# v7.0.1` comment. Condition 5 is a finding on
   that step's `with:`, not on its ref; saying the SHA-to-tag agreement
   could not be verified offline is a coverage gap, not a finding.

Fail if any of:

- harden-runner is recommended, in any wording.
- The report claims to have applied a change; the subject can only read.
- The local action is reported as a pinning gap.
- A defect is fabricated against the fixture: the pinned checkout ref, or
  a `concurrency` block reported missing or wrong.

Findings beyond these — the `run:` step, the trigger set, a note that
`actionlint` and `pinact` could not be run — neither pass nor fail this
case.
