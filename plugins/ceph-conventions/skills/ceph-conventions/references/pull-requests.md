# Ceph pull requests

Owns what a `ceph/ceph` pull request carries that a generic GitHub PR does
not: the verbatim upstream checklist, the `Fixes:` tracker trailer, the DCO
sign-off, and `Co-Authored-By:`. Runs under `SKILL.md`'s precedence and
routing. The parts a Ceph PR inherits unchanged — draft-and-assign, the
description's shape and length ceiling, the closing-keyword hazard, the
comment gate and agent marker — are github-conventions'
`references/pull-requests.md`, and the commit series itself is
github-conventions' `references/commits.md`. The CI a Ceph PR triggers is
`references/ci.md`; a backport PR adds `references/backports.md`.

## Opening and the description

- Draft, assigned to its author — github-conventions'
  `references/pull-requests.md`, "Opening a PR".
- The description's shape and its length ceiling are github-conventions'
  `references/pull-requests.md`, "The description". The upstream checklist
  below is the repository's required template, so it is item 4 there and
  does not count against the ceiling.

## The checklist

`ceph/ceph` ships a pull-request template,
`.github/pull_request_template.md`, whose checklist has Tracker, Component
impact, Documentation, and Tests sections. Keep it verbatim: on a PR to
`main` the "Pull Request Checklist" check fails unless the Tracker,
Documentation, and Tests sections each have a box checked `[x]`, and
reviewers read the real boxes. Check the boxes that apply; never hand-write
a paraphrase or a trimmed custom checklist, and keep the template's "Show
available Jenkins commands" section.

From the CLI the template goes in by shell, never retyped: write the
drafted description to a file, append the default branch's template as it
stands now (`git show <upstream>/main:.github/pull_request_template.md`,
fetched first — the template changes), change only the boxes that apply
from `[ ]` to `[x]`, and pass the file with `gh pr create --body-file` or
`gh pr edit --body-file`.

## The tracker reference

Ceph links a change to its tracker issue with a trailer in the commit
message, just before `Signed-off-by:`, and in the PR description:

```text
Fixes: https://tracker.ceph.com/issues/<N>
```

- The reference is the full `https://tracker.ceph.com/issues/<N>`, never a
  bare `#<N>` (github-conventions' `references/pull-requests.md`, "The
  description").
- `fixes:` is one of GitHub's closing keywords; this trailer closes nothing
  on GitHub only because a tracker URL is not a GitHub reference. The
  closing-keyword rule and its scan still cover everything else in the
  title, description, and commits (github-conventions'
  `references/pull-requests.md`, "Closing keywords").
- On a PR against `main`, the tracker issue moves on merge, not by hand:
  Ceph's Redmine Upkeep workflow finds the issue whose "Pull request ID"
  field holds the PR's number and sets it to Pending Backport, or to
  Resolved when its "Backport" field is empty. So once the PR is open, that
  field carries its number — a tracker write under `references/tracker.md`,
  "Writing to the tracker". A backport PR's number goes on its Backport
  issue instead, never on the issue its `Fixes:` names
  (`references/backports.md`).

## Bot commands

A comment addressed to one of Ceph's bots — `/config check ok`,
`/audit retest`, `jenkins test <check>` — is the bare command and nothing
else, one per comment, because the bots read it from the comment's first
character: the Backport Audit checks that the comment starts with
`/audit retest`, and a Jenkins trigger phrase such as `jenkins test docs.*`
must match from the start. It is still a GitHub post that goes out only on
an explicit instruction for it (github-conventions'
`references/pull-requests.md`, "Comments"), but as a command to a bot
rather than conversation it carries no agent marker.

## Sign-off and co-authorship

- Every commit carries `Signed-off-by: Name <email>` (DCO), which
  `git commit -s` adds. The Jenkins "Signed-off-by" status, required on
  `main` and the stable branches, fails a PR with an unsigned commit. The
  sign-off survives any history rewrite (github-conventions'
  `references/commits.md`, "The rewrite proof").
- A commit with more than one author carries a `Co-Authored-By: Name
  <email>` trailer per additional author, below the sign-off. How AI
  assistance is disclosed is github-conventions'
  `references/pull-requests.md`, "The description" and "Signing", and
  `references/commits.md`, "What a message says".
