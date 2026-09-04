There is no Go toolchain in this environment and no network: `go`,
`gofmt`, and `golangci-lint` are not on PATH, nothing can be installed, and
nothing you write can be built, formatted, or run. Write the code out in your
answer instead, and do not describe any of it as compiled, tested, or tidied.

I want a small Go CLI called `sumit`. It takes numbers as its arguments and
prints their sum; a `--json` flag makes it print the result as JSON instead of
a plain number. It is a brand-new project — an empty directory, nothing
written yet — and the module path should be `github.com/octo-dev/sumit`.

Your entire final answer is that project: the file layout, then the full
contents of every file it needs (at minimum the command, `go.mod`, and the
test).
