# Ceph backports

Owns what a stable-branch backport needs that a forward fix does not: its
commit messages and PR title, its Redmine Backport issue, and a passing
Backport Audit. Runs under `SKILL.md`'s precedence and routing. The PR
itself follows `references/pull-requests.md`; every tracker write named here
is `references/tracker.md`, "Writing to the tracker". Upstream's own account
is `SubmittingPatches-backports.rst` in `ceph/ceph`; read it from the
default branch for a case this reference does not cover.

## The tracker chain

The stable branches change with each release, so the current set comes from
<https://docs.ceph.com/en/latest/releases/>, never from memory. A fix bound
for them is tracked from its parent issue on `main`, a Bug or a Feature, and
Ceph's automation, not the contributor, creates the Backport issues:

- The parent's "Backport" field lists the target releases. Its "Pull
  request ID" field, and what Redmine Upkeep does with it when the `main`
  PR merges, are `references/pull-requests.md`, "The tracker reference".
- Once the parent reaches Pending Backport, the automation creates one
  Backport issue per listed release — tracker Backport, in the parent's
  project, its "Ceph Release" field set, linked from the parent by a
  "Copied to" relation. Upstream discourages creating them by hand.
- The backport is a branch off the stable branch carrying the `main` PR's
  commits as `git cherry-pick -x`, opened as a PR under
  `references/pull-requests.md`; its number goes into its Backport issue's
  "Pull request ID" field.
- Upstream's helpers — `src/script/ceph-backport.sh` for those steps, and
  `src/script/backport-create-issue` for when the automation fails to create
  the Backport issues — send the Redmine key in URLs and credentials on the
  command line, against `references/tracker.md`, "The Redmine API", and
  write to GitHub and the tracker without showing each write first. Running
  one is the owner's explicit decision, never the agent's default.
- Without permission to change a field, upstream's route is a tracker
  comment describing the change.

## Backport commits and PR title

`SubmittingPatches-backports.rst`, "Cherry-picking rules" and "Conflict
resolution", allows a backport commit two forms; which of their rules the
Backport Audit checks is "The Backport Audit" below.

- A commit taken from `main` keeps everything down to its `(cherry picked
  from commit …)` line as `git cherry-pick -x` wrote it, the original
  `Signed-off-by:` included. Below that line it gains a `Conflicts` section
  naming every file changed by hand and describing the change, and upstream
  allows other information to be appended there. A clean cherry-pick takes
  neither the `Assisted-by:` trailer nor the human's sign-off that
  `references/pull-requests.md`, "Sign-off, AI assistance, and
  co-authorship", asks of a new commit; the original sign-off satisfies the
  Jenkins "Signed-off-by" status. Ceph's draft AI policy does not say
  whether an agent-resolved conflict takes an `Assisted-by:` and the
  backporter's sign-off below that line: raise it with the user, and push
  nothing until they have reviewed the resolution.
- A commit made only on the stable branch, not cherry-picked from `main`, has
  a title that starts with the branch name, then the subsystem —
  `squid: qa: …`, `squid: doc/<path>: …`. Otherwise it is a new commit under
  `references/pull-requests.md`, but the human signs it alone, for instance
  with `git commit --amend -s` while it is `HEAD`, because `--signoff` over
  the branch would sign the cherry-picks too.
- The PR's title is the `main` PR's title prefixed with the branch name,
  `squid: <main PR title>`, as `src/script/ceph-backport.sh` sets it.

## The Backport Audit

The "Backport Audit" check (the `releng-audit` workflow), required on the
stable branches, fails a backport PR unless it merges cleanly, carries every
commit of the `main` PR as a `cherry-pick -x` that matches a clean dry-run
cherry-pick, titles every other commit with the branch name, and links up
on the tracker: the parent's "Pull request ID" is the `main` PR, the parent
has a "Copied to" Backport issue whose "Ceph Release" is the PR's base
branch, and that issue's "Pull request ID" is this PR.

- A failure applies the `releng-audit-fail` label, and while it is set a
  push does not re-run the audit. Fix what the bot's review names, then
  remove that label or post `/audit retest`, a bot command
  (`references/pull-requests.md`, "Bot commands"); either waits for the
  user's instruction.
- A conflict resolution fails the audit however well its `Conflicts`
  section documents it, and a retest fails it again: only the override
  below clears it, once a reviewer has checked the resolution.
- An override is the `releng-audit-override` label or an `/audit override`
  comment, honoured only from a repository admin or maintainer or a
  `@ceph/ceph-release-manager` member, and revoked by the next push. It is
  never the agent's to apply: surface the need to the user.
