# Navigation

Owns how code is navigated: which questions go to a language server and
which stay with grep. Runs under `SKILL.md`'s precedence and routing.

- A semantic question — a symbol's definition, its references, its type, the
  diagnostics on a file — goes to the language server through the LSP tool,
  not to grep. Which server answers follows the file: gopls, pyright,
  clangd, a yaml server, marksman, whatever covers the language at hand.
- The LSP tool is usually deferred, not absent: load it with `ToolSearch`
  (`select:LSP`) before concluding no language server covers the file.
  Reading its absence from the tool list as "no server" silently downgrades
  navigation to grep; lower-tier agents get this wrong without the hint.
- grep stays for what it is better at: string and pattern searches, and
  files no language server covers.
