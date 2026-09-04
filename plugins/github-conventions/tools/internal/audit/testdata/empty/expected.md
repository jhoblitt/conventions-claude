| Area | Check | Status | Current | Canon | Fix |
| --- | --- | --- | --- | --- | --- |
| license | present | gap | no LICENSE file | LICENSE exists at the repository root | copy templates/LICENSE |
| license | apache-2.0 | gap | no LICENSE file | Apache-2.0, or the license the repository already carries | copy templates/LICENSE |
| readme | present | gap | no README.md | README.md exists at the repository root | copy templates/README.md |
| readme | scorecard-badge | gap | no README.md | README.md carries the OpenSSF Scorecard badge | copy templates/README.md |
| dependabot | present | gap | no .github/dependabot.yml | .github/dependabot.yml exists | copy templates/dependabot.yml to .github/ |
| dependabot | github-actions | gap | no .github/dependabot.yml | an updates[] entry sets package-ecosystem: github-actions | copy templates/dependabot.yml to .github/ |
| security | codeql | gap | no step uses github/codeql-action/init | a workflow runs github/codeql-action/init | copy templates/codeql.yml to .github/workflows/ |
| security | dependency-review | gap | no step uses actions/dependency-review-action | a workflow runs actions/dependency-review-action | copy templates/dependency-review.yml to .github/workflows/ |
| security | scorecard | gap | no step uses ossf/scorecard-action | a workflow runs ossf/scorecard-action | copy templates/scorecard.yml to .github/workflows/ |
| workflow-lint | actionlint | gap | no run: block invokes actionlint | a workflow runs actionlint over .github/workflows | copy templates/workflow-lint.yml to .github/workflows/ |
| workflow-lint | pin-check | gap | no run: block greps for [0-9a-f]{40} | a workflow fails any uses: not pinned to a 40-hex SHA | copy templates/workflow-lint.yml to .github/workflows/ |
| commitlint | config | gap | no .commitlintrc.yml | .commitlintrc.yml exists at the repository root | copy templates/.commitlintrc.yml |
| commitlint | workflow | gap | no step uses wagoid/commitlint-github-action | a workflow runs wagoid/commitlint-github-action on pull requests | copy templates/commitlint.yml to .github/workflows/ |
| commitlint | breaking-footer | gap | no .github/scripts/check-breaking-footer.sh | .github/scripts/check-breaking-footer.sh exists | copy templates/check-breaking-footer.sh to .github/scripts/ |
| ruleset | default-branch | skipped | not checked (--remote not given) | an active branch ruleset on ~DEFAULT_BRANCH with deletion and non_fast_forward rules |  |

14 gaps, 0 ok, 1 skipped
