# Release

Owns the release contract: goreleaser v2 on `v*` tags, the build targets
and options, archives and checksums, SBOMs, signing, the container image a
project opts into, the release workflow, and the repository without
binaries that the contract does not cover. Runs under `SKILL.md`'s
precedence and routing. `templates/.goreleaser.yaml` and
`templates/release.yml` are the forms; the version stamp they depend on is
`references/layout.md`, "Version"; the placeholders and blocks they are
rendered with are `references/layout.md`, "Template placeholders". Pinning,
permissions, concurrency, timeouts, and checkout credentials belong to
github-conventions' `references/workflows.md`, which the rendered workflow
complies with; what appears below is only where a release departs from the
ordinary shape — the scopes the job escalates to, and a group that is never
cancelled.

## goreleaser

- `version: 2`; one `builds` entry per binary, id and `binary`
  `{{BINARY}}`, `main: ./cmd/{{BINARY}}`, with `CGO_ENABLED=0`,
  `-trimpath`, and `ldflags: [-s -w]` — that list replaces goreleaser's
  default `-X main.version=…` ldflags rather than extending them, because
  the version is the build stamp (`references/layout.md`, "Version").
- Targets: `goos: [linux]` and `goarch: [amd64, arm64]` are the starting
  point, and both lists are the project's to widen or narrow. A binary bound
  to one platform's APIs — a daemon on a session bus — has no use for a
  darwin or windows archive, and every archive built is one checksummed and
  signed on every tag, so the default is the narrow one and cross-platform
  is the exception a project opts into. A project that has changed either
  list has decided, not drifted: the audit does not measure them, and
  `goconv-audit --emit-goreleaser` carries an existing file's lists forward
  (go-converge's `references/goconv-audit.md`).
- Archives (`tar.gz`; a project that adds windows adds the `zip`
  `format_overrides` itself), `checksums.txt`, and a GitHub release with a
  GitHub-generated changelog.
- `sboms`: over the archives, with syft, which is goreleaser's default SBOM
  tool and is installed by the workflow ("The release workflow" below).
- `signs`: cosign keyless `sign-blob --bundle` over the checksum file.
- With an image ("Images" below), `kos` and `docker_signs` as a pair. `kos`
  builds the image from the same checkout with the same build options and
  no `-buildvcs=false`, so the image's binary carries the same VCS stamp as
  the archives: base `gcr.io/distroless/static:nonroot`; `linux/amd64` and
  `linux/arm64`; SPDX SBOM; repository `ghcr.io/{{OWNER}}/{{REPO}}` with
  `bare: true`; tags `latest` and `{{ .Version }}` — goreleaser's own
  template, not a placeholder (`references/layout.md`, "Template
  placeholders"). `docker_signs`: cosign keyless `sign` over the image
  manifests.

## Without binaries

A repository that publishes no binaries — its consumers build the main
package from its checkout, as an action's helper is built — has nothing
for goreleaser to archive, checksum, sign, or describe in an SBOM. It may
release with semantic-release instead, cutting tags and GitHub releases
from the commit history; its release workflow gates on CI by calling
`templates/ci.yml` (`references/ci.md`, "CI"), concurrency keyed as
github-conventions' `references/workflows.md`, "Concurrency", sets for a
release workflow that calls CI. None of this file's
goreleaser contract applies to it, and the audit skips those rows
(go-converge's `references/goconv-audit.md`). A repository that does
publish binaries builds them with goreleaser as above, whatever tool cuts
its tags.

## Images

A release publishes a container image only when the project decides to;
the default is none. An image of a binary that cannot run in a container —
a daemon on one host's session bus, a desktop tool — is unusable by
construction, and the image drags `packages: write` and a registry login
into the release job, so no project pays for one it did not ask for. A
project has opted in when it asked for an image or its `.goreleaser.yaml`
carries either of the two blocks; both blocks are the complete form, one
without the other is the gap the audit names, and neither is no image, not
drift. The blocks are the `{{IMAGE}}` blocks: `kos` and `docker_signs` in
`templates/.goreleaser.yaml`, and `packages: write` and the ghcr.io login
in `templates/release.yml`, rendered by `goconv-audit` (go-converge's
`references/goconv-audit.md`); what the markers are is
`references/layout.md`, "Template placeholders".

An image is never built by a Dockerfile that runs `go build` without `.git`
present: that binary reports `(devel)`. ko inside goreleaser is how the image
gets the stamp.

## The release workflow

`templates/release.yml`, on `push` of tags matching `v*`:

- Fork-guarded: `if: github.repository == '{{OWNER}}/{{REPO}}'`, so a fork
  never publishes under the upstream name. The guard must name the
  repository the releases are tagged in: one naming any other leaves the
  release job never running there. How `goconv-audit` derives it is
  go-converge's `references/goconv-audit.md`.
- The `goreleaser` job escalates to `contents: write` (the release) and
  `id-token: write` (keyless signing), and, with an image, `packages: write`
  (ghcr.io).
- Checkout with `fetch-depth: 0`: the stamp needs the tag.
- Before goreleaser runs, the job builds the binary and asserts
  `go version -m` reports the tag as the module version; a mismatch fails
  the release before anything is published.
- cosign and syft are installed; with an image, the job logs in to ghcr.io
  with `GITHUB_TOKEN`.
- goreleaser-action runs one exact goreleaser release, `version: v2.x.y`,
  never a range and never the input left to its `~> v2` default. The action
  resolves a range to the newest match on every run, so a release nobody
  tested against the project would get the job's write and OIDC tokens, and
  two tags of one project could build with different goreleaser versions.
  The input is the version's one home. Dependabot bumps the action's pin
  but never the input, so a bump edits it by hand.
- The concurrency form — a constant group, never cancelled — is the release
  exception in github-conventions' `references/workflows.md`, "Concurrency".
