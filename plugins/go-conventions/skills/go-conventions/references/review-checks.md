# Review checks

Owns the judgment layer above the oracles for a diff that touches Go — what a
reviewer reports, at what severity, on what evidence; the checks `go-review`
and `go-reviewer` run. Runs under `SKILL.md`'s precedence and routing.

## Scope

The oracles run first and are never re-reported: `go build ./...`, `go vet
./...`, `go fix -diff ./...`, `golangci-lint run` under `templates/.golangci.yml`
(its comments are the linter one-liners; `references/lint.md` the policy), and
the suites — `templates/Makefile`'s `check`, what CI runs. The reviewer's job is
what those cannot see. A target with no lint config gets the oracles run first:
`golangci-lint run --config <copy>` on a copy of `templates/.golangci.yml`
outside the tree, `{{MODULE}}` filled from `go.mod`.

- The literal form a roster linter flags is dropped: its report is the author's
  to fix, never re-derived here. What no linter sees is the reviewer's — the
  residue, named per class below.
- Suppressions are read from the pre-change config, never the diff's: a
  `//nolint` or lint-config edit in the diff is itself a `lint` finding unless
  justified per `references/lint.md`, and what it suppresses is judged as though
  the suppression were absent.

## The package wins

The rung and the ladder are `SKILL.md`'s, Precedence; this is its review-side
application. Census the package before judging the newcomer against any
reference here — the finding is "inconsistent with this package", never a
preference. The verb-prefix census:

```sh
grep -hoE '^func (\([^)]+\) )?[A-Za-z0-9_]+' --exclude='*_test.go' <pkg>/*.go |
  sed -E 's/.* //; s/^([a-z]+|[A-Z][a-z]*).*/\1/' | sort | uniq -c | sort -rn
```

A diff matching the package's consistent idiom is correct even where a
reference here disagrees — nothing, or a nit; the package's own distance from
canon is a converge gap, not this diff's finding. A diff introducing a pattern
the package does not use is the finding, at the severity of its class. A
package split between idioms leaves the rung empty; the reference decides.

## Correctness classes the oracles miss

Tag `correctness`, once the failure path is traced (Verification). Where an
oracle owns the literal form the residue is named; the Kubernetes specifics of
the nil and aliasing cases are `references/kubernetes.md`'s.

- **Goroutine lifetimes.** The stop condition `references/errors-and-style.md`
  requires is present AND reachable — trace what cancels, closes, or `Wait`s;
  one that outlives its caller is a leak, and usually a race on the next call.
- **In-band errors, ignored ok-values.** A zero value where an error belongs;
  `v, _ := m[k]`; a `!ok` branch that proceeds as though `ok`. The unchecked
  single-value `x.(T)` is errcheck's.
- **Nil traps on decoded input.** Optional fields of anything unmarshaled —
  JSON, YAML, protobuf — arrive as nil pointers; name a deref no path
  guarantees by a guard, a default, or a predicate. The Kubernetes form, and
  what to do about it, is `references/kubernetes.md`'s.
- **Aliasing and shared mutation.** A map or slice from a shared object, a
  cache, or another goroutine mutated in place; `append` to a slice the caller
  still holds; a returned map that is the receiver's own.
- **`defer` in a loop** — resources pile up until return; the body wants a
  function. **`time.After` in a select loop** — a timer per iteration until it
  fires; `time.NewTimer`/`NewTicker` with `Stop`.
- **`err :=` shadowing across error paths** — govet's `shadow`, via the template's
  `enable-all`, reports the declaration, the reviewer which check now tests a
  stale value. **Copies of sync types** are vet's `copylocks`; nothing here.
- **String-matched sentinels.** `strings.Contains(err.Error(), "…")` where a
  sentinel or `errors.As` target exists; errorlint covers `==` and type switches.
- **Read-modify-write on shared external state** — an API object, a file, a
  row — without conflict detection and retry; or a decision on a stale read.

## Silent-failure hunt

An error path that LOOKS handled but reports success. Hunt each shape, then
trace the caller's reaction: the consequence sets the severity, and a caller
that treats nil-error as done makes it a blocker. Tag `silent-failure`.

- **A wrap of a possibly-nil error.** `fmt.Errorf("…: %w", err)` with nil `err`
  yields a NON-nil error ending `%!w(<nil>)` — the symptom. The trap is the
  `err` shadowed or reassigned (the shadowing class above) so the real failure
  branch falls through to `return nil`; nilerr catches only the literal
  `return nil` under `err != nil`, the fall-through is proved by hand.
- **Success sentinels.** `(nil, nil)`, `(false, nil)`, an empty slice with nil
  error, on an error branch, fed to a caller for whom nil error means "done".
- **`_ =` on an error return**, `x, _ :=` included, is errcheck's
  (`check-blank`); not re-derived here.
- **Log-and-continue** where the caller assumes success: the loop that logs a
  failed item and reports the batch done; `slog.Error` then `return nil`.
- **`recover()` that eats a panic**: the call returns normally, the work is
  abandoned, the caller records success; it re-panics or returns an error.
- **The cause dropped.** errorlint reports `%v` on an error argument; the rest
  is the reviewer's: `err.Error()` formatted as a string, `errors.New(err.Error())`,
  a custom type without `Unwrap` — where `errors.Is`/`As` needs the chain.

## Modernization

The position and the idiom inventory are `references/errors-and-style.md`'s;
here, the review application:

- An archaic construct on an added or updated line is a `style` finding at
  changes-requested, not a nit; if also wrong, report the correctness finding
  it is. `ptr.To(v)` or `tmp := T(v); &tmp` on an added line is this finding —
  `new(expr)` is the form; a `ptr.To` already standing is not.
- Oracle: `go fix -diff ./<changed-pkgs>/...` (`-diff` prints, never writes:
  safe on a read-only target, where a writing `go fix` never runs), its hunks
  filtered to added and updated lines; `go tool fix help` lists the fixers.
- A fixer's NAME is not its coverage: `newexpr` sounds like every `new(expr)`
  opportunity but rewrites only `varOf`-shaped wrapper functions and their call
  sites; it and golangci-lint's modernize are both silent on an inline
  `tmp := v; return &tmp`, which is why that residue is the reviewer's. A clean
  run means the fixers found nothing THEY look for, never that the added lines
  are modern — read `go tool fix help <fixer>` before trusting silence and judge
  the rest by hand.

## Test quality

Owned by `references/testing.md`, "Test quality"; apply it to every `_test.go` in the diff.

## Coverage adequacy

Lives here, not in `references/testing.md`, so it fires on the dangerous case:
a code diff with no test changes. Tag `test-coverage`.

- Enumerate the diff's new or changed branches, error paths, and boundary
  conditions; for each meaningful one, name the spec that exercises it or
  report the gap, citing the unexercised path.
- A bugfix without a regression spec is changes-requested; a mechanical
  refactor without new specs is not. Red-before/green-after — the spec fails on
  the parent commit, passes on this one — is the standard.
- Integration coverage (`Label("integration")`, `references/testing.md`) is
  expected only for behavior observable solely against the real dependency;
  pure logic wants unit specs.

## Structure and API judgment

Tag `api`; cite the `go.dev/wiki/CodeReviewComments` entry when flagging.

- Exported only if used outside the package. In a module others import,
  removing or changing exported API is a compatibility decision — a question to
  raise, never a `style` finding.
- Five or more positional parameters of one type is a mix-up hazard; suggest a
  parameter struct only where the package already uses that shape.
- A returned interface where `references/layout.md` says a struct
  ("Interfaces"); a naked return in a function longer than a screen ("Naked
  Returns"); named results that document nothing ("Named Result Parameters").

## Naming and comments

Census first (The package wins). Tags `naming`, `comment`.

Naming — the finding is "inconsistent with this package": verb-prefix synonym
drift (`fetchUser` where the package says `getUser`, `removeBucket` beside
`deleteBucket`, `ensureX` where siblings say `configureX`); a new file whose
name does not say what it holds. Initialism casing and receiver-name
consistency are revive's (`var-naming`, `receiver-naming`); package names are
`references/layout.md`'s. Changes-requested when a clear package convention
breaks on an exported or long-lived symbol; nit otherwise.

Godoc — the form is `references/errors-and-style.md`'s and its absence revive
`exported`'s; the reviewer judges the contract (behavior, the inputs' meaning,
error semantics, ownership) and its truth: named parameters exist; claimed
defaults, units, and error behavior match the implementation.

Comments — accuracy before concision; the rule is `references/errors-and-style.md`'s.

- A comment that has drifted from the code it describes is a `comment` finding
  at changes-requested: the diff changed behavior and left the old description,
  or removed what the comment explained and kept the comment. One restating an
  invariant the change just broke is evidence for a code finding — check which
  is right first.
- A restatement, process narration ("as requested" — commit-message material,
  github-conventions' `references/commits.md`), or prompt residue (`TODO(ai)`)
  is the deletion the rule mandates: nit each, changes-requested when pervasive.
- Proofread every changed comment word by word; misspell knows only its dictionary.

## Verification before reporting

Order: oracles (Scope) → the passes above produce candidates → pre-filter →
refutation → exclusions → report.

**Pre-filter**, before any code is read, on the candidate's anchor, its text,
and the diff alone: an anchor outside the diff is pre-existing — if serious,
one mention in the summary, never a finding; it survives only when the
candidate's own text claims the change worsens it or copies it into new code.
An anchor in `go generate` output (what `templates/Makefile`'s `generate-check`
covers) is output nobody edits — the fix is source plus regen — unless the
claim IS the generation: a hand-edited hunk there, or a stale regen the diff
should carry, is a `generated` finding. A linter-flagged form is dropped (Scope).

**Refutation**, for each survivor: attempt to REFUTE it, never to confirm it.
Re-read assuming the author was right and you misread, from the real code, not
the hunk: (1) does the defect exist on the claimed path, callers and callees
included; (2) is it new in this change; (3) does something else already handle
it — a guard upstream, a guaranteeing predicate, an oracle; (4) does the
concrete failure scenario hold end to end? Inline, or by a separate agent given
only the finding and the code, is `go-review`'s call; the report says which.

**Confidence (0–100)**: 0–24 false positive, misreading, pre-existing and
untouched, or speculative ("could", "might", no concrete path) · 25–49
plausible but the failure path not traced to completion · 50–74 real but
minor, rare, or partially mitigated, or real with one unverified link (name
it) · 75–99 verified real, matters, evidence traced, small residual doubt · 100
certain, evidence in hand. Report ≥ 80 as CONFIRMED at any severity, a `nit`
included; 50–79 as PLAUSIBLE only at blocker or changes-requested, naming the
unverified link — a `nit` at 50–79 drops; below 50, drop silently. One strong
finding outranks three weak ones.

**Exclusions**, at any confidence: pedantic nits a senior maintainer would not
raise (micro-style, hypothetical performance, "extract a helper" where none
exists — a helper that DOES exist and the diff re-implemented is not this
class); intentional behavior changes the commits or PR body document as the
point, reported as though accidental; speculative DoS and missing-hardening;
production-grade bars on test-only code, which may take shortcuts production
code may not. Suppressions and incumbent style are excluded above (Scope, The
package wins); missing tests IS reportable (Coverage adequacy).

**Evidence contract** — a reported finding is:

```text
<file:line> — <severity> <tag>: <one-line claim>
  evidence: <what survived refutation — the traced path, the call site, the
             spec that fails; an inference is labelled as one>
  fix: <one line, only where the claim does not already imply it>
```

Severity: `blocker` (merging ships a defect — silent failure, leak, race, data
loss, compatibility break), `changes-requested` (must change before merge),
`nit` (the author's call). Tag, the class that found it: `lint`, `generated`,
`correctness`, `silent-failure`, `style`, `test-quality`, `test-coverage`,
`api`, `naming`, `comment`. Evidence or silence: a finding that cannot show its
evidence is not reported; an inference is labelled as one and keeps its
confidence honest.
