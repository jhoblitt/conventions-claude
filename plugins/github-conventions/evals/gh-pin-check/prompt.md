There is no checkout of the repository under review, no network, and no
`gh` in this environment, and nothing can be installed: `actionlint` and `pinact` are
not on PATH. The workflow below is the whole diff — a new file,
`.github/workflows/ci.yml`, in a repository that has no `workflow-lint`
workflow. Treat it as complete; there is nothing further to fetch.

Review this workflow against our conventions. Your entire final answer is
the review: what the linters would have checked, each finding with the step
or key it lands on and its fix, and what is fine as it stands. Nothing else.

```yaml
name: ci

on:
  pull_request:
  push:
    branches: [main]

concurrency:
  group: ${{ github.workflow }}-${{ github.event.pull_request.number || github.sha }}
  cancel-in-progress: true

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Check out repository
        uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 22

      - name: Lint
        uses: ./.github/actions/lint

      - name: Test
        run: npm test
```
