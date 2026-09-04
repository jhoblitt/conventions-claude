| Area | Check | Status | Current | Canon | Fix |
| --- | --- | --- | --- | --- | --- |
| license | present | ok | LICENSE | LICENSE exists at the repository root |  |
| license | apache-2.0 | ok | Apache-2.0 | Apache-2.0, or the license the repository already carries |  |
| readme | present | ok | README.md | README.md exists at the repository root |  |
| readme | scorecard-badge | ok | scorecard badge present | README.md carries the OpenSSF Scorecard badge |  |
| dependabot | present | ok | .github/dependabot.yml | .github/dependabot.yml exists |  |
| dependabot | github-actions | ok | ecosystems: github-actions | an updates[] entry sets package-ecosystem: github-actions |  |
| workflows | pinned:codeql.yml | ok | every uses: is SHA-pinned | every uses: ends in a 40-hex commit SHA with a # v… comment |  |
| workflows | permissions:codeql.yml | ok | contents: read | a top-level permissions key, with no write scope |  |
| workflows | timeout:codeql.yml | ok | 1 job(s), all with timeout-minutes | every job sets timeout-minutes |  |
| workflows | concurrency:codeql.yml | ok | concurrency group set | a top-level concurrency key |  |
| workflows | checkout-credentials:codeql.yml | ok | persist-credentials: false on all 1 | every actions/checkout step sets with.persist-credentials: false |  |
| workflows | pinned:commitlint.yml | ok | every uses: is SHA-pinned | every uses: ends in a 40-hex commit SHA with a # v… comment |  |
| workflows | permissions:commitlint.yml | ok | contents: read | a top-level permissions key, with no write scope |  |
| workflows | timeout:commitlint.yml | ok | 1 job(s), all with timeout-minutes | every job sets timeout-minutes |  |
| workflows | concurrency:commitlint.yml | ok | concurrency group set | a top-level concurrency key |  |
| workflows | checkout-credentials:commitlint.yml | ok | persist-credentials: false on all 1 | every actions/checkout step sets with.persist-credentials: false |  |
| workflows | pinned:dependency-review.yml | ok | every uses: is SHA-pinned | every uses: ends in a 40-hex commit SHA with a # v… comment |  |
| workflows | permissions:dependency-review.yml | ok | contents: read | a top-level permissions key, with no write scope |  |
| workflows | timeout:dependency-review.yml | ok | 1 job(s), all with timeout-minutes | every job sets timeout-minutes |  |
| workflows | concurrency:dependency-review.yml | ok | concurrency group set | a top-level concurrency key |  |
| workflows | checkout-credentials:dependency-review.yml | ok | persist-credentials: false on all 1 | every actions/checkout step sets with.persist-credentials: false |  |
| workflows | pinned:scorecard.yml | ok | every uses: is SHA-pinned | every uses: ends in a 40-hex commit SHA with a # v… comment |  |
| workflows | permissions:scorecard.yml | ok | contents: read | a top-level permissions key, with no write scope |  |
| workflows | timeout:scorecard.yml | ok | 1 job(s), all with timeout-minutes | every job sets timeout-minutes |  |
| workflows | concurrency:scorecard.yml | ok | concurrency group set | a top-level concurrency key |  |
| workflows | checkout-credentials:scorecard.yml | ok | persist-credentials: false on all 1 | every actions/checkout step sets with.persist-credentials: false |  |
| workflows | pinned:workflow-lint.yml | ok | every uses: is SHA-pinned | every uses: ends in a 40-hex commit SHA with a # v… comment |  |
| workflows | permissions:workflow-lint.yml | ok | contents: read | a top-level permissions key, with no write scope |  |
| workflows | timeout:workflow-lint.yml | ok | 2 job(s), all with timeout-minutes | every job sets timeout-minutes |  |
| workflows | concurrency:workflow-lint.yml | ok | concurrency group set | a top-level concurrency key |  |
| workflows | checkout-credentials:workflow-lint.yml | ok | persist-credentials: false on all 2 | every actions/checkout step sets with.persist-credentials: false |  |
| security | codeql | ok | codeql.yml | a workflow runs github/codeql-action/init |  |
| security | dependency-review | ok | dependency-review.yml | a workflow runs actions/dependency-review-action |  |
| security | scorecard | ok | scorecard.yml | a workflow runs ossf/scorecard-action |  |
| workflow-lint | actionlint | ok | workflow-lint.yml | a workflow runs actionlint over .github/workflows |  |
| workflow-lint | pin-check | ok | workflow-lint.yml | a workflow fails any uses: not pinned to a 40-hex SHA |  |
| commitlint | config | ok | .commitlintrc.yml | .commitlintrc.yml exists at the repository root |  |
| commitlint | workflow | ok | commitlint.yml | a workflow runs wagoid/commitlint-github-action on pull requests |  |
| commitlint | breaking-footer | ok | .github/tools/breaking-footer/main.go | .github/tools/breaking-footer/main.go exists |  |
| ruleset | default-branch | skipped | not checked (--remote not given) | an active branch ruleset on ~DEFAULT_BRANCH with deletion and non_fast_forward rules |  |

0 gaps, 39 ok, 1 skipped
