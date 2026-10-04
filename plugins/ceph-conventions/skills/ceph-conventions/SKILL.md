---
name: ceph-conventions
description: Use when contributing to ceph/ceph or tracker.ceph.com — opening or updating a Ceph pull request, filling its checklist, writing a Fixes trailer or a DCO sign-off; clearing the config-diff "Check Ceph config changes" check or reading the Jenkins ceph-pr-pipeline; preparing a stable-branch backport, tracking its Redmine Backport issue, or answering the Backport Audit; filing, searching, commenting on, or editing a tracker.ceph.com issue through the Redmine API; or handling a Ceph defect that may be a security vulnerability.
---

# Ceph conventions

The canon for how a contribution reaches `ceph/ceph` and `tracker.ceph.com`
the house way — the pull-request workflow, the Jenkins and config-diff CI
checks, the stable-branch backport machinery, filing and searching on the
Redmine tracker, and the embargoed path a security defect takes. It sits
on top of github-conventions, which owns repository hygiene, commit, and
pull-request practice, and code-conventions, which owns what a bug report
carries and where a vulnerability goes; this canon adds only Ceph's own
instance of each rule and points to the sibling for the rest. It is
consulted, not run. A path of the form `references/<file>` is relative to
this skill's directory; one prefixed with a sibling plugin's name —
"github-conventions' `references/<file>`", "code-conventions'
`references/<file>`" — is that plugin's.

## Precedence

The ladder is github-conventions' `SKILL.md`, "Precedence"; for
`ceph/ceph` and `tracker.ceph.com` work this canon is the project-specific
conventions plugin its first rung names. It depends on both sibling
plugins: if the Skill tool lists no `github-conventions:*` or no
`code-conventions:*` skill, stop and say to install it. Where Ceph's
process differs from the generic rule — the verbatim PR checklist, the
`Fixes:` trailer, the DCO sign-off, the bare bot command, the Jenkins and
config-diff checks, the backport tracker, the reproduction bar — Ceph's
instance wins, because it is written for this project. Everything the
generic rule still governs — the draft-and-assign, the description's shape
and length ceiling, the closing-keyword hazard, the comment gate and agent
marker, what a bug report carries — is inherited unchanged.

## Always

- A `ceph/ceph` PR carries the upstream checklist verbatim, never a custom
  one — `references/pull-requests.md`.
- Every commit is `Signed-off-by:` (DCO), and a tracker issue is referenced
  by a `Fixes: <tracker URL>` trailer, never a bare `#N` —
  `references/pull-requests.md`.
- A backport's Redmine Backport issue comes from Ceph's automation, not by
  hand — `references/backports.md`.
- A defect is filed on `tracker.ceph.com` only after it reproduces on a
  running system — `references/tracker.md`, "Before filing".
- A memory-safety, unauthenticated-crash, or authorization-bypass defect
  goes to `security@ceph.io` first and stays embargoed —
  `references/tracker.md`, "Security disclosure".
- No `tracker.ceph.com` write — an issue, a comment, a field, a status, a
  relation — is made without an explicit instruction for it —
  `references/tracker.md`, "Writing to the tracker". A GitHub post is
  github-conventions' `references/pull-requests.md`, "Comments".

## Reference routing

Read a reference when the work touches its trigger; skip the rest. The rules
above apply to every trigger.

| Doing this | Read |
|---|---|
| opening or updating a `ceph/ceph` PR, filling its checklist, writing a `Fixes:` trailer, a DCO sign-off, or `Co-Authored-By:`; posting a bot command (`/config check ok`, `/audit retest`, `jenkins test`) | `references/pull-requests.md` |
| clearing the config-diff "Check Ceph config changes" check, reading the Jenkins ceph-pr-pipeline, or watching Ceph CI | `references/ci.md` |
| preparing a stable-branch backport PR, tracking its Redmine Backport issue, or answering the Backport Audit | `references/backports.md` |
| filing, searching, commenting on, or editing a `tracker.ceph.com` issue; judging the reproduction bar; a possible security vulnerability | `references/tracker.md` |
