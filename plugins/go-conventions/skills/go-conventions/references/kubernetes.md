# Kubernetes

Owns what changes when a module imports `k8s.io/*` or
`sigs.k8s.io/controller-runtime`: the kubebuilder layout, custom-resource
fields, informer-cache objects, conflict retry, reconcile cancellation, the
controller logger, and the lint config. Runs under `SKILL.md`'s precedence
and routing, and is read only when `go.mod` pulls those modules. The generic
forms of the nil and aliasing rules are `references/review-checks.md`,
"Correctness classes the oracles miss".

## Layout

The kubebuilder/controller-runtime scaffold is kept as generated:
`cmd/main.go`, `api/<group>/<version>/`, `internal/controller/`.
`references/layout.md`'s `cmd/<bin>/main.go` rule yields to it. So does
`references/cli.md`'s cobra rule: the scaffolded manager keeps the stdlib
`flag` wiring it is generated with, which is what binds controller-runtime's
zap logger flags — the same rung that keeps that logger ("Logging" below).
A new binary in the module that is not the manager is a cobra command tree
like any other.

## Custom resources

Write every access defensively; the review check on the same ground is
`references/review-checks.md`, "Correctness classes the oracles miss".

- Optional CR fields arrive as nil pointers, so a field is dereferenced only
  behind a guard, after a default has been applied, or under a predicate
  that guarantees it — on every path that reaches the deref.
- Objects from the informer cache are shared with every other reader:
  `DeepCopy()` before any mutation; never modify what `Get` or `List`
  returned in place.
- A read-modify-write on an API object needs conflict retry —
  `retry.RetryOnConflict` around a fresh `Get` and the `Update` — or a
  `Patch`; a stale `resourceVersion` is the expected case, not an error path.

## Reconcile

A reconcile path never blocks without honoring `ctx`: API calls, waits, and
external clients take the reconcile's context.

## Logging

controller-runtime's zap logger (`sigs.k8s.io/controller-runtime/pkg/log/zap`,
`log.FromContext(ctx)`) is the incumbent for a controller:
`references/logging.md`'s slog rule yields to it under the precedence rung
(`SKILL.md`, "Precedence"), and `references/lint.md`'s zap denial does not
apply to it.

## Lint

A stock kubebuilder `.golangci.yml` is v1 schema — no `version`,
`run.deadline`, `linters.disable-all` — and hard-fails on current
golangci-lint; it is replaced from `templates/.golangci.yml`
(`references/lint.md`), never edited into shape.
