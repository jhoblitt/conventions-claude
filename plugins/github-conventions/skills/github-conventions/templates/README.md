# {{REPO}}

[![{{CI_WORKFLOW}}](https://github.com/{{OWNER}}/{{REPO}}/actions/workflows/{{CI_WORKFLOW}}.yml/badge.svg)](https://github.com/{{OWNER}}/{{REPO}}/actions/workflows/{{CI_WORKFLOW}}.yml)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/{{OWNER}}/{{REPO}}/badge)](https://scorecard.dev/viewer/?uri=github.com/{{OWNER}}/{{REPO}})

{{DESCRIPTION}}

## Install

{{INSTALL}}

## Usage

{{USAGE}}

## Development

Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/);
commitlint enforces this on every pull request. Every GitHub Action is pinned
to a commit SHA: run `pinact run` after editing a workflow and `actionlint`
before committing it. Where a `Makefile` is present, `make check` is the local gate,
`make tools` installs the pinned linter, and that pin lives in the `Makefile`.

## License

[Apache-2.0](LICENSE)
