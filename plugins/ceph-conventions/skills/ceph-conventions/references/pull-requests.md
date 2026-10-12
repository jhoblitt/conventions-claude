# Ceph pull requests

Owns what a `ceph/ceph` pull request carries that a generic GitHub PR does
not: the verbatim upstream checklist, the subsystem-prefixed commit title,
the `Fixes:` tracker trailer, the DCO sign-off, the `Assisted-by:` trailer,
and `Co-Authored-By:`. Runs under `SKILL.md`'s precedence and routing. The
parts a Ceph PR inherits unchanged — draft-and-assign, the description's
shape and length ceiling, the closing-keyword hazard, the comment gate and
agent marker — are github-conventions' `references/pull-requests.md`, and
the commit series itself, its "Conventional Commits" section aside, is
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

## Commit titles

A `ceph/ceph` commit title is `<subsystem>: <description>`, not a
Conventional Commits subject. `SubmittingPatches.rst`, "Commit title", asks
for one line of at most 72 characters, in the imperative mood, prefixed with
the subsystem or module the change touches — `mds: add perf counter for
finisher of MDSRank`, or `doc/<path>: …` for a change only under `doc/`.
Take the prefix from the recent subjects on the files the commit touches,
`git log --format=%s -n 20 -- <files>`, rather than inventing one;
`src/test/rgw/s3-tests/`, for instance, is `s3tests:`. A PR title follows
the same convention and describes the whole set of changes
(`SubmittingPatches.rst`, "PR title"). `SubmittingPatches.rst` sets this
format, so github-conventions' `references/commits.md`, "Conventional
Commits", does not apply to a Ceph commit; the rest of that file does. A
backport's commits and PR title are `references/backports.md`, "Backport
commits and PR title".

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

## Sign-off, AI assistance, and co-authorship

Ceph's AI policy is a draft:
[ceph/ceph#66788](https://github.com/ceph/ceph/pull/66788), at head
`4a7c0b8bf547` on 2026-10-12, the text Ceph Steering Committee motion
[#79745](https://tracker.ceph.com/issues/79745) puts to a vote. It adds
`doc/dev/developer_guide/ai-coding-assistants.rst` and a pointer to it in
`SubmittingPatches.rst`. This section, and the cherry-pick bullet of
`references/backports.md`, "Backport commits and PR title", follow that
draft at that head; revising them against the merged text is this canon's
upkeep, not a step of a Ceph change. A backport's cherry-pick takes its
trailers and sign-off from that bullet rather than from the bullets below,
whose prohibitions — no agent `Signed-off-by:`, no assistant in a
`Co-Authored-By:` — still hold.

- Every commit carries `Signed-off-by: Name <email>` (DCO). The Jenkins
  "Signed-off-by" status, required on `main` and the stable branches, fails a
  PR with an unsigned commit. The sign-off survives any history rewrite
  (github-conventions' `references/commits.md`, "The rewrite proof").
- An agent never adds `Signed-off-by:` — no `git commit -s`, no
  `--signoff`, not in the human's name, and not on their confirmation. Only a
  human can certify the DCO, and the human's sign-off also certifies that
  every AI assistance is disclosed. The agent leaves its commits unsigned and
  says so; the human signs them before anything is pushed, for instance with
  `git rebase --signoff --keep-base <upstream>/<branch>` on a branch of new
  commits only, which signs every commit it replays without moving the base.
- A rewrite that folds agent-written content into a commit the human already
  signed — a squashed fixup, an amend — carries their sign-off over content
  they have not certified. The agent says so, and nothing is pushed until the
  human has reviewed that commit.
- AI assistance is disclosed by an `Assisted-by:` trailer on each commit it
  went into, in the draft's format, where square brackets mark the optional
  parts:

  ```text
  Assisted-by: AGENT_NAME-AGENT_VERSION[:MODEL_NAME] [TOOL1] [TOOL2]
  ```

  `AGENT_VERSION` is required. `MODEL_NAME` is the underlying model, where
  there is one. The tools are specialised analysis tools such as
  `clang-tidy` or `mypy`, never git, a compiler, make, or an editor. For
  Claude Code, follow the draft's own example
  (`Claude-v2.1.199:claude-4-5-sonnet`): the agent name is `Claude`, the
  version is the number `claude --version` prints prefixed with `v`, and the
  model is its model ID — `Assisted-by: Claude-v2.1.296:claude-opus-5-5`.
  The draft asks for no disclosure in the PR description; the trailers are
  the disclosure.
- The draft names no other attribution. This canon adds that an assistant is
  never named in a `Co-Authored-By:`, which names a human author; a harness
  default that asks for one yields to this rule (github-conventions'
  `SKILL.md`, "Precedence"). A commit with more than one human author carries
  a `Co-Authored-By: Name <email>` trailer per additional author.
- The agent writes its trailers as `Assisted-by:`, then any
  `Co-Authored-By:`, then any `Fixes:`, so the sign-off the human appends
  lands just after `Fixes:`, where "The tracker reference" puts it. The
  draft sets no order of its own.
- A review by an AI tool is allowed, but a human's review is still required
  before a change merges. The draft does not require transcripts of the AI
  interaction; the change and its documentation must be reviewable on their
  own.
