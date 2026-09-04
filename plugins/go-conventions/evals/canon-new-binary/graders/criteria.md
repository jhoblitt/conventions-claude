The request is the ordinary one a user with no plugin knowledge makes: a small
new CLI, `sumit`, summing its numeric arguments, with a `--json` flag, module
`github.com/octo-dev/sumit`, in an empty directory. Nothing in the request
names a framework, a logger, a test library, a layout, or a Go version — the
canon supplies all of them, and this case checks that the subject applied it
instead of writing the flat `flag`-and-`fmt.Println` program the request
invites.

Under the canon: every binary is a cobra tree whose root sets `SilenceUsage`
and `SilenceErrors`, configured through a per-tree `viper.New()` under the
binary's environment prefix, with values read back through `v.Get*`;
`log/slog` with a JSON handler on stderr, stdout reserved for the program's
output; `cmd/<bin>/main.go` thin over `internal/`; the version from
`debug.ReadBuildInfo` and never from `-X` ldflags; `go 1.27`, minor only;
Ginkgo v2 and Gomega with the suite bootstrap's `RandomizeAllSpecs` and
`FailOnPending`; and `fmt.Errorf` with `%w` for wrapping.

The regressions this case exists to catch: reaching for the standard `flag`
package because the program is small, printing diagnostics to stdout where
they corrupt the program's output, testify or bare `testing` in the test, and
stamping the version with `-X` ldflags.

Pass if and only if ALL of:

1. The root command is a `github.com/spf13/cobra` command that sets both
   `SilenceUsage: true` and `SilenceErrors: true`.
2. Configuration goes through viper — a `viper.New()` instance, not the
   package-level singleton — with the environment prefix `SUMIT` (any
   spelling of `SetEnvPrefix("SUMIT")` or `SetEnvPrefix("sumit")`), and every
   configured value is read back through `v.GetString`/`v.GetBool`/`v.GetInt`
   and the like, never through a Go variable bound by `StringVar`/`BoolVar`.
3. Logging is `log/slog` through a JSON handler writing to stderr, and stdout
   carries only the program's output — the sum, or the JSON result. No
   diagnostic, log line, or error message is written to stdout.
4. The layout is `cmd/sumit/main.go` plus at least one `internal/` package
   holding the command tree (`internal/cli` or equivalent), and the version is
   read from `debug.ReadBuildInfo` — an `internal/version` package, or
   `--version`/cobra's `Version` field fed from `debug.ReadBuildInfo`.
5. `main` opens a `signal.NotifyContext` (`os.Interrupt`, `SIGTERM`) and
   returns or exits with status 1 when the run returns an error.
6. `go.mod` carries `go 1.27` — the minor only, no patch digit — and no
   `toolchain` line.
7. There is a Ginkgo suite bootstrap file that calls `RegisterFailHandler`
   and `RunSpecs` with `RandomizeAllSpecs = true` and `FailOnPending = true`,
   plus at least one real spec — a `Describe`/`Context`/`It` or a
   `DescribeTable` — asserting the summing behavior with Gomega.
8. Wherever an error carries a cause, it is wrapped with `fmt.Errorf` and
   `%w`. A `%v`, an `errors.Wrap`, or a concatenated `err.Error()` fails this;
   a bare `errors.New` for an originating error does not.
9. No `-X` linker flag appears anywhere in the answer — not in a Makefile, a
   workflow, a `.goreleaser.yaml`, or a `go build` line.

Fail if any of:

- The standard-library `flag` package, kong, or urfave/cli is imported or
  used for argument or flag parsing.
- The test uses testify, or the standard `testing` package for assertions
  (a `func TestX(t *testing.T)` that only bootstraps Ginkgo is the required
  form, not a violation).
- A diagnostic or log line is emitted with `log.Printf`, `log.Println`,
  `log.Fatal`, `fmt.Print*` to stdout, or `println`.
- `interface{}` appears where `any` is meant.
- A `pkg/` directory appears in the layout.
- `github.com/pkg/errors` is imported.
- The answer claims to have built, run, formatted, tidied, or tested the
  code, or reports a command's output; the subject has no toolchain.

Detail beyond these neither passes nor fails: whether the answer also
includes a Makefile, `.golangci.yml`, workflows, `.gitignore`, or `CLAUDE.md`;
whether the sum is integer or floating point; the JSON object's field names;
how invalid arguments are rejected; and the presence of `--config`,
`--log-level`, or `--log-format` beyond what condition 3 requires.
