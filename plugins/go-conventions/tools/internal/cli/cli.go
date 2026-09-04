// Package cli implements the goconv-audit command line.
package cli

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jhoblitt/conventions-claude/plugins/go-conventions/tools/internal/audit"
)

// Run executes goconv-audit with args, writing the audit to stdout and
// diagnostics to stderr. It returns an error only on a usage or I/O failure; a
// repository full of gaps is a successful run.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	var asJSON, asMarkdown, emitGolangci bool
	var filesArea string

	root := &cobra.Command{
		Use:           "goconv-audit [dir]",
		Short:         "Report where a Go repository diverges from the go-conventions canon",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, cmdArgs []string) error {
			opts := audit.Options{
				Dir:        ".",
				PluginRoot: os.Getenv("CLAUDE_PLUGIN_ROOT"),
				Logger:     slog.New(slog.NewJSONHandler(stderr, &slog.HandlerOptions{Level: slog.LevelWarn})),
			}
			if len(cmdArgs) == 1 {
				opts.Dir = cmdArgs[0]
			}

			out, err := render(opts, asJSON, emitGolangci, filesArea)
			if err != nil {
				return err
			}
			_, err = io.WriteString(stdout, out)

			return err
		},
	}

	root.SetArgs(args)
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)

	flags := root.Flags()
	flags.BoolVar(&asJSON, "json", false, "render the audit as JSON")
	flags.BoolVar(&asMarkdown, "markdown", false, "render the audit as a markdown table (the default)")
	flags.BoolVar(&emitGolangci, "emit-golangci", false, "print the house lint config for this repository")
	flags.StringVar(&filesArea, "files", "", "print the files a migration of this area touches")
	root.MarkFlagsMutuallyExclusive("json", "markdown", "emit-golangci", "files")

	return root.ExecuteContext(ctx)
}

func render(opts audit.Options, asJSON, emitGolangci bool, filesArea string) (string, error) {
	switch {
	case emitGolangci:
		return audit.Golangci(opts)
	case filesArea != "":
		paths, err := audit.Files(opts, filesArea)
		if err != nil {
			return "", err
		}
		if len(paths) == 0 {
			return "", nil
		}

		return strings.Join(paths, "\n") + "\n", nil
	}

	report, err := audit.Run(opts)
	if err != nil {
		return "", err
	}
	if !asJSON {
		return report.Markdown(), nil
	}

	data, err := report.JSON()
	if err != nil {
		return "", err
	}

	return string(data), nil
}
