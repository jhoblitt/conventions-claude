# Chat links

Owns how a repository, pull request, merge request, or issue is referred to
in a chat reply, whatever host serves it. Runs under `SKILL.md`'s precedence
and routing.

- Every mention in a chat reply's prose, tables, or lists — a repeat of
  one already linked included — is one Markdown link: the display text is
  the mention exactly as it would have been written bare, and the target is
  the page of the item the mention names —
  `[#42](https://github.com/o/r/pull/42)`,
  `[!42](https://gitlab.example/g/p/-/merge_requests/42)`,
  `[g/p](https://gitlab.example/g/p)`, and
  `[o/r#42](https://github.com/o/r/pull/42)`, because a mention that
  carries the repository name still names the pull request. A bare number
  in a terminal is something the user has to go and look up; a link is the
  lookup done.
- The page each item lives at:

  | Host | Item | Its page |
  |---|---|---|
  | GitHub | pull request `#N` | `https://github.com/<owner>/<repo>/pull/N` |
  | GitHub | issue `#N` | `https://github.com/<owner>/<repo>/issues/N` |
  | GitHub | repository `<owner>/<repo>` | `https://github.com/<owner>/<repo>` |
  | GitLab | merge request `!N` | `https://<host>/<path>/-/merge_requests/N` |
  | GitLab | issue `#N` | `https://<host>/<path>/-/issues/N` |
  | GitLab | project `<path>` | `https://<host>/<path>` |

  The `owner/repo` shape carries no host; the remote does. With remote
  `https://gitlab.com/g/p.git`, the project mention `g/p` is
  `[g/p](https://gitlab.com/g/p)`. A branch or file has no row:
  `g/p/cmd/main.go` is that project mention followed by a path, so it is
  `[g/p](https://gitlab.com/g/p)/cmd/main.go`, the path plain. A host
  canon adds only its host's rows; the rule and its table live here.
- The host and project path come from the repository the item was filed
  against: the upstream remote of the checkout under discussion, not a
  fork that only carries a pushed branch — or the name or URL the user
  gave. The host is the remote's hostname alone; userinfo in the URL
  (`oauth2:<token>@`) is not part of it and never appears in a link. No
  host is assumed: a self-hosted instance is whatever the remote names, a
  nested GitLab path (`group/subgroup/project`) is carried whole, and a
  mention whose host neither source supplies stays bare. A fact the remote
  cannot supply — which repository a project's shorthand name means —
  belongs to a rung above this canon, not here.
- A code span or fenced block is not prose: a `gh` command line, an import
  path, a diff shows verbatim, because a link inside it would corrupt what
  the block is quoting.
