// Package cli implements the goconv-audit command line.
package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jhoblitt/conventions-claude/plugins/go-conventions/tools/internal/audit"
)

// mode is which of the mutually exclusive renderings a run performs; all
// false with an empty filesArea is the markdown audit.
type mode struct {
	asJSON, golangci, goreleaser, workflow bool
	filesArea                              string
}

// releaseFlags are the value flags only the two release renderings read.
var releaseFlags = []string{"binary", "owner", "repo", "image"}

// Run executes goconv-audit with args, writing the audit to stdout and
// diagnostics to stderr. It returns an error only on a usage or I/O failure; a
// repository full of gaps is a successful run.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	var (
		m          mode
		asMarkdown bool
		inputs     audit.Inputs
	)

	root := &cobra.Command{
		Use:           "goconv-audit [dir]",
		Short:         "Report where a Go repository diverges from the go-conventions canon",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, cmdArgs []string) error {
			if err := releaseOnly(cmd, m); err != nil {
				return err
			}

			opts := audit.Options{
				Dir:        ".",
				PluginRoot: os.Getenv("CLAUDE_PLUGIN_ROOT"),
				Logger:     slog.New(slog.NewJSONHandler(stderr, &slog.HandlerOptions{Level: slog.LevelWarn})),
			}
			if len(cmdArgs) == 1 {
				opts.Dir = cmdArgs[0]
			}

			out, err := render(opts, m, inputs)
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
	flags.BoolVar(&m.asJSON, "json", false, "render the audit as JSON")
	flags.BoolVar(&asMarkdown, "markdown", false, "render the audit as a markdown table (the default)")
	flags.BoolVar(&m.golangci, "emit-golangci", false, "print the house lint config for this repository")
	flags.BoolVar(&m.goreleaser, "emit-goreleaser", false, "print .goreleaser.yaml rendered for this repository")
	flags.BoolVar(&m.workflow, "emit-release-workflow", false, "print the release workflow rendered for this repository")
	flags.StringVar(&m.filesArea, "files", "", "print the files a migration of this area touches")
	flags.StringVar(&inputs.Binary, "binary", "", "the binary the release files are rendered for (default: the one cmd/<name>)")
	flags.StringVar(&inputs.Owner, "owner", "", "the GitHub owner the release files name (default: from a github.com module path)")
	flags.StringVar(&inputs.Repo, "repo", "", "the GitHub repository the release files name (default: from a github.com module path)")
	flags.BoolVar(&inputs.Image, "image", false, "render the release files with the container image")
	root.MarkFlagsMutuallyExclusive("json", "markdown", "emit-golangci", "emit-goreleaser", "emit-release-workflow", "files")

	return root.ExecuteContext(ctx)
}

// releaseOnly rejects a release value flag given with any other mode. cobra
// expresses "never together" but not "only with", so this is the check.
func releaseOnly(cmd *cobra.Command, m mode) error {
	if m.goreleaser || m.workflow {
		return nil
	}

	var given []string
	for _, name := range releaseFlags {
		if cmd.Flags().Changed(name) {
			given = append(given, "--"+name)
		}
	}
	if len(given) == 0 {
		return nil
	}

	return fmt.Errorf("%s: only with --emit-goreleaser or --emit-release-workflow", strings.Join(given, ", "))
}

func render(opts audit.Options, m mode, inputs audit.Inputs) (string, error) {
	switch {
	case m.golangci:
		return audit.Golangci(opts)
	case m.goreleaser:
		return audit.Goreleaser(opts, inputs)
	case m.workflow:
		return audit.ReleaseWorkflow(opts, inputs)
	case m.filesArea != "":
		paths, err := audit.Files(opts, m.filesArea)
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
	if !m.asJSON {
		return report.Markdown(), nil
	}

	data, err := report.JSON()
	if err != nil {
		return "", err
	}

	return string(data), nil
}
