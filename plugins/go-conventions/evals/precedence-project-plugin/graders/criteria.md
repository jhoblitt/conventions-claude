The fixture is a repository whose `CLAUDE.md` says it wraps errors with
`github.com/pkg/errors` and logs with zerolog and to keep that, plus a package
whose one existing method, `Drain`, already does exactly that. The diff adds
`CloseIdle`, which uses `errors.Wrapf` and `log.Ctx(ctx).Info()` the same way
`Drain` does — and puts `defer f.Close()` inside the `for` loop, so a handle is
held for every idle connection until the whole function returns.

Under the canon's precedence ladder, the repository's own `CLAUDE.md` is the
top rung and the surrounding package's consistent idiom is the rung above this
canon: a change that matches that idiom is correct even where the canon
disagrees. `log/slog` and `%w` are the canon's defaults, and the house lint
tier's depguard denies `github.com/pkg/errors` and zerolog — both outranked
here, twice over. `defer` in a loop is a `correctness` class the oracles miss,
and it is the defect this diff actually carries.

The discriminator, since these two look alike: a bare classification is neutral
— saying the package's distance from canon is not this diff's finding, and
stopping there, costs nothing. Anything that proposes, plans, schedules, or
defers the migration fails, "a converge gap to close later" included. Elsewhere
that framing would be correct; here the top rung says to KEEP those packages, so
deferring the migration overrides it exactly as recommending it now would.

The regressions this case exists to catch: applying the canon over the
repository's `CLAUDE.md` and the incumbent package idiom, and going so far into
deference that the planted defect goes unreported.

Pass if and only if ALL of:

1. The report makes NO finding — of any severity, `nit` included — about
   `github.com/pkg/errors`, `errors.Wrap`, `errors.Wrapf`, or the absence of
   `%w`.
2. The report makes NO finding — of any severity, `nit` included — about
   zerolog, `log.Ctx`, or the absence of `log/slog`.
3. The report states the precedence rule, in a line of its own: the
   repository's `CLAUDE.md` and the surrounding package's consistent idiom
   outrank the canon's defaults, so the incumbent error and logging packages
   stand.
4. The report has a finding on the `defer f.Close()` inside the `for` loop in
   `CloseIdle`. Either framing of that site satisfies this condition: that the
   handles accumulate until the function returns rather than being released per
   iteration, or — since `openSocket` returns a write handle — that the dropped
   `Close()` error can lose data. A finding anchored there on either ground
   passes; no finding there fails.

Fail if any of:

- The report proposes, plans, schedules, or defers a move to `log/slog` or to
  `fmt.Errorf` with `%w`, in any wording and in any position: a finding, a nit,
  a suggestion, a follow-up, or a converge gap to close later. A bare
  classification that stops at "not this diff's finding" is not this bullet;
  anything pointing at a future migration is, because the repository's
  `CLAUDE.md` forecloses the migration itself and not merely this diff.
- The report treats the depguard deny entries for `github.com/pkg/errors` or
  zerolog as something this diff must satisfy.
- A defect is fabricated against unchanged context: `Drain`, `openSocket`,
  `drainInto`, or the `conn` struct.
- The report claims to have run a command or applied a fix; the subject can
  only read.

Findings beyond these neither pass nor fail this case: the absent spec for
`CloseIdle`, the socket left open across `c.close()`, and a note that the
oracles could not be run here. The dropped `Close()` error is not in this list —
under condition 4 it is one of the two accepted framings of the `defer` site.
