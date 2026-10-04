# tracker.ceph.com

Owns how a Ceph defect or feature is searched for, filed, and written to on
`tracker.ceph.com`, the project's Redmine, and where a Ceph security defect
goes instead. Runs under `SKILL.md`'s precedence and routing. What any bug
report carries, the prior-art search, the gate on filing, and where a
vulnerability goes are code-conventions' `references/bug-reports.md`; this
reference adds only Ceph's instances — the Redmine API and its ids, the
tracker's marker and write gate, the reproduction bar, and the
`security@ceph.io` channel.

## The Redmine API

`tracker.ceph.com` is Redmine, read and written through its REST API. Reads
of public issues, trackers, and projects need no key. A write, and the read
of the user's own account, are authenticated with the user's own API key in
the `X-Redmine-API-Key` header, read from an owner-only (`0600`) file that
holds that one header line, so the key never lands in a URL, an exported
variable, or a command line:

```sh
curl -sf https://tracker.ceph.com/issues.json \
  -H @<header-file> \
  -H 'Content-Type: application/json' \
  -d '{"issue":{"project_id":<PROJECT>,"tracker_id":<TRACKER>,"subject":"…","description":"…"}}'
```

- Tracker ids are global: Bug `1`, Feature `2`, Backport `9`. Confirm any
  other against `GET /trackers.json` rather than guessing.
- A project id is per component. Resolve it with
  `GET /projects/<identifier>.json`, the identifier being the slug in the
  project's tracker URL (`rgw` in `https://tracker.ceph.com/projects/rgw`),
  rather than hardcoding a number or reading the paged `GET /projects.json`.
- A Backport issue, and the parent fields that bring it into being, are
  `references/backports.md`.

## Writing to the tracker

A filed report is gated by code-conventions' `references/bug-reports.md`.
Every other write — a comment, a field, a status, a relation, a Backport
issue — goes out only on an explicit instruction for that write, given after
the user has seen what it sets. A comment or an issue description written on
the user's behalf opens with the line

```text
> This is @<login>'s AI agent.
```

where `<login>` is the user's tracker login, from `GET /users/current.json`
sent with the same header file — github-conventions'
`references/pull-requests.md`, "Signing", gives the marker's rules.

## Before filing

The prior-art search and what a match means are code-conventions'
`references/bug-reports.md`. Ceph's instances: the tracker search is
`tracker.ceph.com`'s own full-text search plus the component project's
recently updated issues, and the forge search is `ceph/ceph`'s open and
recently closed PRs touching the affected files.

This canon sets a stricter reproduction bar than code-conventions' generic
rule: a Ceph defect is filed only after it reproduces on a running system —
a disposable cluster counts, code analysis does not — with no exception for
a defect that cannot be exercised. Until it reproduces, it stays recorded
as unreproduced in the owner's own registry.

## Security disclosure

A Ceph defect that is, or may be, a memory-safety bug, an unauthenticated
crash, or an authorization bypass is a security vulnerability, and Ceph's
private channel for it is `security@ceph.io`:

- The report carries what Ceph's `SECURITY.md` asks for — a reproducer, the
  affected versions, a fix if one exists, who is to be credited, and any
  intended disclosure date — in the package code-conventions'
  `references/security-disclosures.md` shapes. The Ceph security team
  assigns the CVE once it confirms the report; the reporter does not
  request one.
- It stays embargoed until the Ceph security team coordinates disclosure: no
  `tracker.ceph.com` issue, no public PR, no pushed branch carrying the fix,
  and no public comment.
- Recording an embargoed defect in the owner's private registry and filing
  it upstream are different acts; the first may happen while nothing reaches
  Ceph's public tracker.
