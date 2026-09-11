# Comments

Owns, in any language, when a comment earns its place and when it is
deleted, and what a comment or a doc — a README, any doc in the
repository — describes. Runs under `SKILL.md`'s precedence and routing.

- A comment earns its place by saying what the code cannot: a non-obvious
  why, a constraint, a gotcha, an external reference. Code that needs
  narration to be understood is rewritten, not annotated.
- A comment that restates the code is deleted on sight. The signals: it
  restates the signature or the next line; it narrates control flow ("loop
  over the items"); it is a block comment out of proportion to the file's or
  module's comment density; it is process or prompt residue ("as requested",
  "updated per review", `TODO(ai)`, "the user"). The remedy is deletion, not
  rewording: keep any real why, cut the rest.
- A comment orphaned by an edit — the predicate, workaround, or constraint
  it explained is gone — is deleted in the same edit; deleting beats
  updating.
- A comment or a doc describes the software as it is, not how it got that
  way. What it supersedes, what it was modelled on, which ticket prompted
  it, what a later commit brings: that belongs in the commit message or
  the PR description, where a reader looking for history goes —
  github-conventions' `references/commits.md`, "What a message says",
  where that plugin is installed. Deleted on sight in any file already
  under edit, true though it is: provenance ("modelled on X's
  Containerfile", "copied from Y"); supersession ("replaces the vendored
  copy that lived in utils/foo", "carried over from the tool this
  replaces"); motivation by ticket ("the failure that motivated this",
  "see !432"); a forward reference ("the exporter itself lands in a later
  commit"). The part that still constrains the code stays: "scanned line
  by line rather than unmarshalled: these files carry YAML anchors" is a
  why that still holds; "carried over from the tool this replaces" in
  front of it is not.
- Behavior that changes between stable releases is documentation, not
  history, and belongs in the README: a changed default, a deprecation, a
  compatibility note, what a migration requires. The test is whether it
  tells a reader something they need in order to use the version in front
  of them: "since v2.0 the default is X" does; "this replaces the old
  tool" does not.
