# Code citations

Owns how a durable document cites code by file and line, whatever host
serves the code. Runs under `SKILL.md`'s precedence and routing.

A durable document is one read after the session that wrote it: a spec, a
plan, a registry entry, a review finding, a PR description. A bare
`file.cc:123-145` leaves the reader to find the repository, the tree, and
the lines by hand, and a citation verified against one tree but read
against another opens unrelated code without saying so.

- A file:line citation in a durable document is a Markdown link to that
  line range. The display text is the citation exactly as written; the
  target is the file at the ref the citation was verified against — the
  document's declared base, or a marker in the text naming it.
- The ref is a tag or a commit, never a branch: a branch moves, and the
  lines move with it.
- A continuation — a bare `:N-M` after a file citation — links into the
  same file at the same ref, unless the text names another.
- The line anchor follows the host: `#L10-L20` on GitHub, `#L10-20` on
  GitLab. A single line is `#L10` on both.
- A code span or fenced block stays verbatim, by the rule
  `references/chat-links.md` sets for a link. The one exception here: a
  code span holding nothing but the citation is linked whole.
- A citation whose file or ref cannot be settled stays plain. No link is
  better than a wrong one, which a reader trusts.
