# Release targets and the image opt-in

**Status:** accepted, 2026-09-11. **Non-normative** — see `AGENTS.md`,
"Design records", for what that means and what this file may hold.
Revisits the release decision in `2026-09-03-go-conventions-design.md`.

## Why

The 2026-09-03 record chose goreleaser with ko, cosign, and an SBOM, and
rejected goreleaser without images. Applying that default to a daemon bound
to one host's session bus showed its cost: such a project has no use for
darwin or windows archives or a container image, every archive built is one
more checksummed and signed on every tag, and the image drags
`packages: write` and a registry login into every release job.

## Decided

- **Linux-only targets as the starting point.** A project adds operating
  systems it actually ships to. Rejected: keeping the multi-platform
  matrix as the default.
- **The container image is an opt-in.** A project that publishes an image
  turns it on; one that does not carries none of its permissions or
  steps. Rejected: the image on by default with a per-project override,
  which is what gnome-monitor-pin had to do by hand; making the image a
  scaffold-time question rather than an opt-in.
- **The tool renders the two release files, not the model.** The pre-PR
  review found that applying the opt-in was a deterministic transform the
  model was performing from prose. Rejected: the model copying the
  templates and editing the opt-in by hand.
