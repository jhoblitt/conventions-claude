# Disclosure gists

Owns how a security-vulnerability disclosure is delivered as a GitHub gist,
on the owner's instruction to put one there. The package's shape — its
severity-first opening, its body, its fix, and its links — is
code-conventions' `references/security-disclosures.md`; this reference adds
only the GitHub instance: the files, their names, the section order of
`BUG-REPORT.md`, which holds for a gist alone, and the `gh` commands. Runs
under `SKILL.md`'s precedence and routing.

- A disclosure gist is secret and multi-file. `gh gist create` makes a
  secret gist by default — `--public` is its only visibility flag and is
  never passed. The report's title is the gist description (`-d`); the files
  are the positional arguments. An existing gist is updated with
  `gh gist edit`, not recreated.
- Creating or editing a gist runs only on the owner's instruction for that
  act, given after they have seen the final local files and the exact `gh`
  command line; deleting one, after they have seen the command.
- The file set is fixed, so every gist reads the same way:
  - `BUG-REPORT.md` — the package's document, shaped by code-conventions'
    `references/security-disclosures.md`.
  - `fix.diff` — the proposed fix as a standalone unified diff against a
    named release.
  - `repro.<ext>` — the self-contained reproducer, its extension the
    upstream project's language (code-conventions'
    `references/bug-reports.md`): `repro.cc` for a radosgw defect.
  - `repro-notes.md` — the exact build and run command, and the output the
    run produced.
- `BUG-REPORT.md` carries its sections in this order: the title (`#`), the
  severity estimate as `## Estimated CVSS`, linking down to
  `## Impact and CVSS`, then `## Summary`, `## Affected versions`,
  `## Component`, `## Vulnerability`, `## Reachability`, `## Impact and CVSS`
  (the fuller scoring), `## Reproduction` (pointing to `repro.<ext>` and
  `repro-notes.md`), `## Prior art`, `## Suggested fix` (pointing to
  `fix.diff`), and `## Disclosure`.
- A code link is a GitHub blob URL, at the ref and with the line anchor
  code-conventions' `references/code-citations.md` sets.
- A secret gist is link-private, not access-controlled: anyone with the URL
  reads it, and GitHub does not embargo it. It is a staging artifact for the
  owner to share by link — never the project's private disclosure channel,
  which code-conventions' `references/bug-reports.md` selects — and it
  carries nothing the owner has not cleared for link-based sharing.
