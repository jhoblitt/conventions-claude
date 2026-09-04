# Logging

Owns the logger: `log/slog`, its handler and sink, the flags that control
it, and the shape of a log call. Runs under `SKILL.md`'s precedence and
routing. `templates/root.go`, `setupLogging`, is the form; the flags' binding
and environment names are `references/cli.md`, "Configuration"; sloglint's
settings live in `templates/.golangci.yml` (`references/lint.md`); a
controller-runtime manager's logger is `references/kubernetes.md`, "Logging".

## Sink and handler

- `log/slog` only. No `log`, logrus, zap, or zerolog — depguard denies the
  three (`references/lint.md`).
- The handler writes to stderr. stdout is reserved for program output, so
  the binary composes in a pipe.
- `--log-format` selects the handler — `json` or `text` — and defaults to
  `json`, so an unconfigured run gets `slog.NewJSONHandler`.
  `--log-level=debug|info|warn|error` (default `info`) sets its level.
  `<PREFIX>_LOG_FORMAT` and `<PREFIX>_LOG_LEVEL` are the environment forms.
  The level flag is the only verbosity switch: no `-v`, `--verbose`,
  `--debug`, or `--quiet`.
- `slog.SetDefault` installs the handler in `PersistentPreRunE`, before any
  command runs.
- Where a `context.Context` is in scope, the `…Context` variant:
  `slog.InfoContext(ctx, …)`, never `slog.Info(…)`.

## Messages and attributes

- The message is a static, lowercase string — `"reading config"` — never
  built with `Sprintf`, never capitalized.
- Everything variable is an attribute: `slog.String("path", p)`,
  `slog.Int`, `slog.Any` — never a bare key-value pair (`"path", p`).
- Keys are `snake_case`.
- sloglint enforces exactly this: `context: scope`, `static-msg`,
  `msg-style: lowercased`, `attr-only`, `key-naming-case: snake`.
- Never log a secret — a token, a password, a private key, a URL carrying
  credentials — at any level.
