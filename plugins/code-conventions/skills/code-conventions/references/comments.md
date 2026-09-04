# Comments

Owns when a comment earns its place and when it is deleted, in any
language. Runs under `SKILL.md`'s precedence and routing.

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
