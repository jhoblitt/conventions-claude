| Area | Check | Status | Current | Canon | Fix |
| --- | --- | --- | --- | --- | --- |
| license | present | ok | LICENSE | LICENSE exists at the repository root |  |
| license | apache-2.0 | ok | MIT | Apache-2.0, or the license the repository already carries |  |
| readme | present | ok | README.md | README.md exists at the repository root |  |
| readme | scorecard-badge | gap | no scorecard badge | README.md carries the OpenSSF Scorecard badge | add the badge line from templates/README.md |
| dependabot | present | gap | no .github/dependabot.yml | .github/dependabot.yml exists | copy templates/dependabot.yml to .github/ |
| dependabot | github-actions | gap | no .github/dependabot.yml | an updates[] entry sets package-ecosystem: github-actions | copy templates/dependabot.yml to .github/ |
| workflows | pinned:ci.yaml | gap | unpinned: actions/checkout@v3, actions/setup-go@v4, actions/checkout@v3 | every uses: ends in a 40-hex commit SHA with a # v… comment | run pinact run over .github/workflows |
| workflows | permissions:ci.yaml | gap | no top-level permissions key | a top-level permissions key, with no write scope | add a top-level permissions: contents: read |
| workflows | timeout:ci.yaml | gap | no timeout-minutes on: build, test | every job sets timeout-minutes | add timeout-minutes to each job |
| workflows | concurrency:ci.yaml | gap | no top-level concurrency key | a top-level concurrency key | add a top-level concurrency group keyed on the workflow and the ref |
| workflows | checkout-credentials:ci.yaml | gap | persist-credentials not false on 2 of 2 | every actions/checkout step sets with.persist-credentials: false | add with.persist-credentials: false to every actions/checkout step |
| workflows | pinned:stale.yml | gap | unpinned: actions/stale@v9 | every uses: ends in a 40-hex commit SHA with a # v… comment | run pinact run over .github/workflows |
| workflows | permissions:stale.yml | gap | permissions: write-all | a top-level permissions key, with no write scope | move the write scopes onto the jobs that need them |
| workflows | timeout:stale.yml | gap | no timeout-minutes on: stale | every job sets timeout-minutes | add timeout-minutes to each job |
| workflows | concurrency:stale.yml | gap | no top-level concurrency key | a top-level concurrency key | add a top-level concurrency group keyed on the workflow and the ref |
| workflows | checkout-credentials:stale.yml | ok | no actions/checkout steps | every actions/checkout step sets with.persist-credentials: false |  |
| security | codeql | gap | no step uses github/codeql-action/init | a workflow runs github/codeql-action/init | copy templates/codeql.yml to .github/workflows/ |
| security | dependency-review | gap | no step uses actions/dependency-review-action | a workflow runs actions/dependency-review-action | copy templates/dependency-review.yml to .github/workflows/ |
| security | scorecard | gap | no step uses ossf/scorecard-action | a workflow runs ossf/scorecard-action | copy templates/scorecard.yml to .github/workflows/ |
| workflow-lint | actionlint | gap | no run: block invokes actionlint | a workflow runs actionlint over .github/workflows | copy templates/workflow-lint.yml to .github/workflows/ |
| workflow-lint | pin-check | gap | no run: block greps for [0-9a-f]{40} | a workflow fails any uses: not pinned to a 40-hex SHA | copy templates/workflow-lint.yml to .github/workflows/ |
| commitlint | config | gap | no .commitlintrc.yml | .commitlintrc.yml exists at the repository root | copy templates/.commitlintrc.yml |
| commitlint | workflow | gap | no step uses wagoid/commitlint-github-action | a workflow runs wagoid/commitlint-github-action on pull requests | copy templates/commitlint.yml to .github/workflows/ |
| commitlint | breaking-footer | gap | no .github/tools/breaking-footer/main.go | .github/tools/breaking-footer/main.go exists | copy templates/breaking-footer/main.go to .github/tools/breaking-footer/ |
| ruleset | default-branch | skipped | not checked (--remote not given) | an active branch ruleset on ~DEFAULT_BRANCH with deletion and non_fast_forward rules |  |

20 gaps, 4 ok, 1 skipped
