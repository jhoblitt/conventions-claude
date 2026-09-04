---
name: code-conventions
description: Use when navigating unfamiliar code — locating a symbol's definition, its callers, its type, or the diagnostics on a file; when writing, reviewing, or deleting a comment; or when judging whether a piece of code needs explanation.
---

# Code conventions

The canon for the rules that hold whatever the language is — how code is
navigated, and when a comment is written or deleted. It is consulted, not
run: no skill executes it, and a language canon points here rather than
restating it, adding only that language's instance of a rule. A path of the
form `references/<file>` is relative to this skill's directory.

## Precedence

This canon is the ladder's lowest rung: anything that speaks more
specifically wins on conflict, a language canon's own statement of a rule
included, because a rule written for one language knows more about that
language than a rule written for all of them.

The rungs above it are enumerated once, in github-conventions' `SKILL.md`,
"Precedence". Read that where it is installed; this plugin does not require
it, and the sentence above is all this canon needs to be obeyed correctly on
its own.

## Reference routing

Read a reference when the work touches its trigger; skip the rest.

| Doing this | Read |
|---|---|
| locating a definition, its callers, or a symbol's type; reading a file's diagnostics; choosing between a language server and grep | `references/navigation.md` |
| writing, reviewing, or deleting a comment; judging whether code needs explanation | `references/comments.md` |
