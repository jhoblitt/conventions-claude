# Release

Owns the release contract: goreleaser v2 on `v*` tags, the build options,
archives and checksums, images through ko, SBOMs, signing, and the release
workflow. Runs under `SKILL.md`'s precedence and routing.
`templates/.goreleaser.yaml` and `templates/release.yml` are the forms; the
version stamp they depend on is `references/layout.md`, "Version"; the
placeholders they are rendered with are `references/layout.md`, "Template
placeholders". Pinning, permissions, concurrency, timeouts, and checkout
credentials belong to github-conventions' `references/workflows.md`, which
the rendered workflow complies with; what appears below is only where a
release departs from the ordinary shape — the scopes the job escalates to,
and a group that is never cancelled.

## goreleaser

- `version: 2`; one `builds` entry per binary, id and `binary`
  `{{BINARY}}`, `main: ./cmd/{{BINARY}}`, with `CGO_ENABLED=0`,
  `-trimpath`, and `ldflags: [-s -w]` — that list replaces goreleaser's
  default `-X main.version=…` ldflags rather than extending them, because
  the version is the build stamp (`references/layout.md`, "Version").
- Archives (`tar.gz`; `zip` on Windows), `checksums.txt`, and a GitHub
  release with a GitHub-generated changelog.
- `kos`: ko builds the image from the same checkout with the same build
  options and no `-buildvcs=false`, so the image's binary carries the same
  VCS stamp as the archives. Base `gcr.io/distroless/static:nonroot`;
  `linux/amd64` and `linux/arm64`; SPDX SBOM; repository
  `ghcr.io/{{OWNER}}/{{REPO}}` with `bare: true`; tags `latest` and
  `{{ .Version }}` — goreleaser's own template, not a placeholder
  (`references/layout.md`, "Template placeholders").
- `sboms`: over the archives, with syft, which is goreleaser's default SBOM
  tool and is installed by the workflow ("The release workflow" below).
- `signs`: cosign keyless `sign-blob --bundle` over the checksum file.
  `docker_signs`: cosign keyless `sign` over the image manifests.

## Images

An image is never built by a Dockerfile that runs `go build` without `.git`
present: that binary reports `(devel)`. ko inside goreleaser is how the image
gets the stamp.

## The release workflow

`templates/release.yml`, on `push` of tags matching `v*`:

- Fork-guarded: `if: github.repository == '{{OWNER}}/{{REPO}}'`, so a fork
  never publishes under the upstream name.
- The `goreleaser` job escalates to `contents: write` (the release),
  `packages: write` (ghcr.io), and `id-token: write` (keyless signing).
- Checkout with `fetch-depth: 0`: the stamp needs the tag.
- Before goreleaser runs, the job builds the binary and asserts
  `go version -m` reports the tag as the module version; a mismatch fails
  the release before anything is published.
- cosign and syft are installed, and the job logs in to ghcr.io with
  `GITHUB_TOKEN`.
- The concurrency form — a constant group, never cancelled — is the release
  exception in github-conventions' `references/workflows.md`, "Concurrency".
