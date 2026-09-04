// Package cli implements the ghconv-audit command line.
package cli

import (
	"context"
	"io"
	"log/slog"

	"github.com/spf13/cobra"

	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/audit"
)

// Run executes ghconv-audit with args, writing the audit to stdout and
// diagnostics to stderr. It returns an error only on a usage or I/O failure;
// a repository full of gaps is a successful run.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	var asJSON, asMarkdown, remote bool

	root := &cobra.Command{
		Use:           "ghconv-audit [dir]",
		Short:         "Report where a repository diverges from the github-conventions canon",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, cmdArgs []string) error {
			dir := "."
			if len(cmdArgs) == 1 {
				dir = cmdArgs[0]
			}

			report, err := audit.Run(cmd.Context(), audit.Options{
				Dir:    dir,
				Remote: remote,
				Logger: slog.New(slog.NewJSONHandler(stderr, &slog.HandlerOptions{Level: slog.LevelWarn})),
			})
			if err != nil {
				return err
			}

			out := []byte(report.Markdown())
			if asJSON {
				out, err = report.JSON()
				if err != nil {
					return err
				}
			}

			_, err = stdout.Write(out)

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
	flags.BoolVar(&remote, "remote", false, "also check the branch ruleset, through gh")
	root.MarkFlagsMutuallyExclusive("json", "markdown")

	return root.ExecuteContext(ctx)
}
