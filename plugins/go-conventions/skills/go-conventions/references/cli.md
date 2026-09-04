# CLI and configuration

Owns the command-line contract — cobra for every binary — the viper
configuration contract, and the stdlib-only carve-out for hook and launcher
binaries. Runs under `SKILL.md`'s precedence and routing. `templates/root.go`
is the wiring for both contracts; the `main` that calls it is
`references/layout.md`, "main"; the logger it installs is
`references/logging.md`.

## cobra

- Every binary is a cobra command tree, even a single-command one:
  `templates/root.go`. No stdlib `flag`, kong, or urfave/cli — the one
  exception is a scaffolded kubebuilder manager
  (`references/kubernetes.md`, "Layout").
- `Run(ctx, args, stdin, stdout, stderr) error` builds the tree and runs
  `ExecuteContext`; the root sets `SilenceUsage` and `SilenceErrors`, so
  the error reaches the user once, printed by `main`
  (`references/layout.md`, "main").
- Streams are the caller's: `cmd.SetIn`, `SetOut`, and `SetErr` on the
  root, from the writers `Run` was handed. A command prints through
  `cmd.OutOrStdout()`, or through the writer that was passed down —
  `templates/root.go` hands `stderr` to the logger setup — never through
  `os.Stdout` or `os.Stderr`, so a spec can capture what it prints.
- `Version: version.String()` wires `--version`
  (`references/layout.md`, "Version").
- The root's persistent flags are `--config`, `--log-level`, and
  `--log-format` (`references/logging.md`); a command's own flags live on
  that command.

## Configuration

- `viper.New()` per command tree, never the package-level singleton.
- `PersistentPreRunE` on the root: `SetEnvPrefix(<BINARY>)`,
  `SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))`,
  `AutomaticEnv()`, `BindPFlags(cmd.Flags())`, then `ReadInConfig` when
  `--config` names a file. It runs before every command, so a subcommand's
  own flags are bound as well.
- Precedence: flags, then environment, then the config file, then
  defaults. The environment form of a key is `<PREFIX>_<KEY>` with `-`
  and `.` mapped to `_`: `--log-level` reads `<PREFIX>_LOG_LEVEL`.
- Values are read back through `v.GetString`, `v.GetInt`, `v.GetBool`, …
  — never through the Go variable a flag was bound to.
- Required-ness is checked in `PreRunE` —
  `if v.GetString("x") == "" { return errors.New("--x is required") }` —
  never `MarkFlagRequired`, which knows nothing about the environment or
  the file.

Pitfalls, each the bug a rule above prevents:

- A flag bound to a Go variable (`StringVar`) never sees an environment or
  file value; only viper resolves those.
- Keys are case-insensitive.
- `AutomaticEnv` resolves only keys viper already knows — a default, a
  config-file key, or a bound flag; an environment variable with no such
  key is invisible.

## Hook and launcher binaries

The one exemption from the framework rules above and from
`references/testing.md`: a hook or launcher binary the plugin builds
offline on first use — `hooks/goconv-hook` — is stdlib only, with no
`require` in its `go.mod`, and its tests use the standard `testing`
package, because a first `go build .` with no module cache must succeed.
The `go` directive is exempt too: such a binary declares the lowest version
its own code needs, not the version `references/toolchain.md` sets, because
a directive above the toolchain on the machine fails that first build before
it starts — and a launcher that fails open fails silently with it.
Nothing else is exempt.
