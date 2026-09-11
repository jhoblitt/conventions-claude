---
name: code-conventions
description: Use when navigating unfamiliar code — locating a symbol's definition, its callers, its type, or the diagnostics on a file; when writing, reviewing, or deleting a comment; when writing or editing a README or any other doc; when judging whether a piece of code needs explanation; or when a chat reply names a repository, pull request, merge request, or issue.
---

# Code conventions

The canon for the rules that hold whatever the language or host is — how
code is navigated, when a comment is written or deleted, what a comment or
a doc describes, and how a repository, pull request, merge request, or
issue is referred to in chat. It is consulted, not run: no skill executes
it, and a language or host canon points here rather than restating it,
adding only its own instance of a rule. A path of the form
`references/<file>` is relative to this skill's directory.

## Precedence

This canon is the ladder's lowest rung: anything that speaks more
specifically wins on conflict, a language or host canon's own statement of
a rule included, because a rule written for one language or host knows
more about it than a rule written for all of them.

The rungs above it are enumerated once, in github-conventions' `SKILL.md`,
"Precedence". Read that where it is installed; this plugin does not require
it, and the sentence above is all this canon needs to be obeyed correctly on
its own.

## Reference routing

Read a reference when the work touches its trigger; skip the rest.

| Doing this | Read |
|---|---|
| locating a definition, its callers, or a symbol's type; reading a file's diagnostics; choosing between a language server and grep | `references/navigation.md` |
| writing, reviewing, or deleting a comment; writing or editing a README or other doc; judging whether code needs explanation | `references/comments.md` |
| naming a repository, pull request, merge request, or issue in a chat reply | `references/chat-links.md` |
