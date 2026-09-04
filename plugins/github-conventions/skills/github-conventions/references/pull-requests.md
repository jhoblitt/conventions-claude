# Pull requests

Owns a pull request's life after the commits exist: opening it, the
description, the multi-PR campaign budget, watching CI, and posting to
GitHub — the gate and the agent marker. Runs under `SKILL.md`'s precedence
and routing. The commits themselves — messages, branch history, the
rewrite proof — are `references/commits.md`.

## Opening a PR

- The branch is a logical series before the push
  (`references/commits.md`, "Branch history").
- `gh pr create --draft --assignee @me`: always a draft; assigned to its
  author, best-effort — a failed assignment (no permission) is skipped,
  not retried.

## The description

In order:

1. **Motivation** — the problem as the author experienced it. When a
   feature or behavior change's request stated none, ask for it before
   drafting; never reconstruct one from the diff — it reads plausible
   while missing the actual reason.
2. **What changed** — the new user-visible behavior.
3. **Notable decisions** — only a choice a reviewer would otherwise
   question, one or two sentences each. Usually none; omit the section
   rather than fill it.
4. Whatever the repository requires: its PR template, its AI-assistance
   disclosure ("Signing" below).

A reviewer gets the point from the first paragraph. Hard limit: 100 words
across items 1–3 (`wc -w`, markup included) — a ceiling, not a target;
required disclosures and checklists do not count. Omit any section with
nothing to say. When the body outgrows the limit, detail moves into commit
messages, not into the description. Process stays out
(`references/commits.md`, "What a message says").

## Campaign budget

In a multi-PR campaign, open at most ~3 PRs before checking in with the
user.

## Watching CI

After a PR is opened or pushed, watching is a concrete action, not a
promise: in the same turn as the push, start ONE combined background
watcher for every run being tracked, and say so. Never end the turn
having only said you will watch.

Polling:

- `gh api repos/<owner>/<repo>/actions/runs/<id> --jq .status` on a
  minutes-scale interval (~3 min); jobs run for tens of minutes, so finer
  granularity buys nothing. Never `gh run watch`: it polls every few
  seconds and exhausts the API quota.
- Fetch job details only after a run completes.
- On an HTTP 403 that names a rate limit, check `gh api rate_limit` (free)
  and sleep past the reset.
- The harness's own limits — a sandbox's API quota, which clone or
  worktree is in use — are the user's configuration, not this canon's.

Triage each failing check by whether the PR plausibly caused it:

- Plausibly caused: diagnose it and push a fix.
- Unlikely, and plausibly flaky (a transient network error, a suite known
  to flake): restart the job, up to 3 times; a failure that survives the
  retries is called out as a surviving flake, never "fixed".

Stop watching when CI is green, the user says stop, a fix needs a decision
only the user can make, or only retried flakes remain.

## Comments

### Posting requires an instruction

No comment, review, or reply is posted to a GitHub PR or issue without an
explicit instruction for that specific post — including a "done" reply or
a re-review ping, and including PRs the user authored. Addressing feedback
in code (edits, commits, pushes) needs no instruction. An ambiguous
instruction ("respond to item B") is not authorization: it means "handle
it in code" or "draft the reply for me" — ask which. Default to drafting
the text in chat for the user to post.

### Signing

Every conversational post made on the user's behalf — a PR or issue
comment, a review body, a review reply, an issue filed — opens with the
exact line

```text
> This is @<login>'s AI agent.
```

where `<login>` is `gh api user --jq .login`, never hardcoded. The marker
is verbatim, never paraphrased, and is the whole attribution: no trailing
sign-off. It does not apply to PR descriptions or commit messages, which
carry the repository's AI-assistance disclosure in the user's voice.
