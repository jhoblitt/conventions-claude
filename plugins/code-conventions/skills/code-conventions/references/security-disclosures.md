# Security disclosures

Owns the shape of the disclosure package — the document that carries a
security vulnerability to the channel `references/bug-reports.md` selects.
That reference owns what any report carries, when it is filed, and where a
vulnerability goes; this one adds only what a vulnerability disclosure adds
on top of a report: a severity estimate up front, the proposed fix, and a
form the owner reviews before it leaves. Host- and tracker-agnostic — a
secret gist, a draft advisory, a private tracker ticket, and a vendor email
all carry the same package. Runs under `SKILL.md`'s precedence and routing.

- Severity first. The package opens, directly under the title, with an
  estimated severity scored with CVSS — a range when the vector is not yet
  pinned — and nothing else, so a reader triages from the first screen. The
  reasoning that produced it (the vector, each metric's justification, the
  aggravators, and the open questions that widen the range) is a later
  section the opening links to, never inlined into it. Each score is
  computed from its vector by a CVSS calculator — such as the `cvss` Python
  package's `CVSS3(vector).scores()` or `CVSS4(vector).scores()` — never by
  hand.
- The report body is a bug report. What it carries — the environment, a
  reproduction that was run, the observed result beside the expected one, a
  control where one is needed, and the source citation — is
  `references/bug-reports.md`, and the reproducer follows that reference's
  language rule. This reference does not restate any of it.
- The fix travels with the report. A disclosure carries the fix it proposes
  as a unified diff against a named release, readable on its own — the patch
  itself, not a prose description of it. A disclosure with no fix yet says so
  in place of the diff rather than omitting the section.
- Every reference is a link. Each mention the package makes of code, a
  commit, a pull request, an issue or tracker ticket, or a repository is a
  link to its canonical page, so a reader reaches it in one click; a code
  citation's file-and-line form is `references/code-citations.md`.
- Staged until released. The package is assembled locally and staged — a
  secret gist, or a draft advisory on a repository the owner administers —
  only on the owner's explicit instruction, given after they have reviewed
  its final files. It stays unpublished until the owner directs it to a
  channel. Which channel, and the gate on sending it, are
  `references/bug-reports.md`; this reference governs the document, not its
  release.
