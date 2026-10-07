# Ceph CI

Owns the checks a `ceph/ceph` PR triggers that a generic repository does not:
the config-diff check and the Jenkins ceph-pr-pipeline. Runs under
`SKILL.md`'s precedence and routing. How CI is watched, and how a failing
check is triaged and retried, is github-conventions'
`references/pull-requests.md`, "Watching CI"; this reference adds only what
Ceph's checks mean and how they are reached.

## The config-diff check

The "Check Ceph config changes" workflow runs on every PR and fails when the
PR changes a config option in `src/common/options/*.yaml.in`, asking the
author to update the release documentation where the change needs it. It is
not a required status.

- It is cleared by a `/config check ok` comment on the PR followed by a
  re-run — re-run the job, edit the PR description, or push — because the
  workflow does not trigger on comments; the comment alone changes nothing.
  One such comment anywhere in the thread passes every later run, so it
  goes out once the documentation is settled.
- The comment is a bot command (`references/pull-requests.md`, "Bot
  commands"). Draft it and the re-run plan, and wait.
- A PR that changes no option passes; do not post the comment
  speculatively.

## The Jenkins ceph-pr-pipeline

Ceph's CI runs on Jenkins (`jenkins.ceph.com`) as well as GitHub Actions.
The ceph-pr-pipeline reports back as commit statuses — among them "make
check" (the build and unit tests), "ceph API tests", "ceph windows tests",
and "Signed-off-by" — which `gh pr checks <N> --repo ceph/ceph` lists beside
the Actions runs, so that is what the watch polls.

- For an author without write access, Jenkins does not run until a
  maintainer applies the `ci-approved` label, and every push removes it. A
  run that never starts is waiting on that label: tell the user rather than
  waiting on it.
- Ceph's instance of the triage: a "make check" or "ceph API tests" failure
  in code the PR touched is plausibly caused. "ceph windows tests" is a
  required status on the stable branches, so on a backport a failure there
  is never left as a flake.
- Jenkins is retriggered by a bot command, `jenkins test <check>`, as the
  template's "Show available Jenkins commands" section lists them
  (`references/pull-requests.md`, "Bot commands").
