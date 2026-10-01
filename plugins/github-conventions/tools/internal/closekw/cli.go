package closekw

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

// Run executes ghconv-closekw with args. It prints one line per match, then,
// under --pr, a head line naming the commit the scan saw, then a summary line;
// matches are a successful run, and only a usage, read, git, or gh failure, an
// incomplete commit list, a commit count that changed during the scan, a head
// that moved during the scan, a failed stdout write, or a canceled ctx returns
// an error.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	var (
		description, revRange, dir, repo string
		pr                               int
	)

	root := &cobra.Command{
		Use:           "ghconv-closekw --pr N [--repo OWNER/REPO] [--dir DIR] | [--description FILE|-] [--range REV-RANGE] [--dir DIR]",
		Short:         "Report closing keywords directly before an issue or PR reference",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			matches, head, err := collect(cmd.Context(), stdin, options{
				description: description, revRange: revRange, dir: dir,
				pr: pr, prSet: cmd.Flags().Changed("pr"), repo: repo,
			})
			if err != nil {
				return err
			}

			for _, m := range matches {
				if _, err := fmt.Fprintln(stdout, m); err != nil {
					return fmt.Errorf("write: %w", err)
				}
			}

			if head != "" {
				if _, err := fmt.Fprintf(stdout, "head %s\n", head); err != nil {
					return fmt.Errorf("write: %w", err)
				}
			}

			if _, err := fmt.Fprintf(stdout, "%d matches\n", len(matches)); err != nil {
				return fmt.Errorf("write: %w", err)
			}

			return nil
		},
	}

	root.SetArgs(args)
	root.SetIn(stdin)
	root.SetOut(stdout)
	root.SetErr(stderr)

	flags := root.Flags()
	flags.IntVar(&pr, "pr", 0, "pull request whose title, description, and commits to scan, read through gh")
	flags.StringVar(&repo, "repo", "", "repository of --pr, as gh --repo takes it (default: the checkout's)")
	flags.StringVar(&description, "description", "", "local PR description to scan: a file, or - for stdin")
	flags.StringVar(&revRange, "range", "", "local commits whose messages to scan, as git rev-list takes them (origin/main..HEAD)")
	flags.StringVar(&dir, "dir", ".", "repository that git and gh run in")
	root.MarkFlagsMutuallyExclusive("pr", "description")
	root.MarkFlagsMutuallyExclusive("pr", "range")

	return root.ExecuteContext(ctx)
}

type options struct {
	description, revRange, dir, repo string
	pr                               int
	prSet                            bool
}

// collect runs the scan o asks for. head is the PR head commit under --pr,
// and empty otherwise.
func collect(ctx context.Context, stdin io.Reader, o options) (matches []Match, head string, err error) {
	switch {
	case o.prSet:
		if o.pr <= 0 {
			return nil, "", fmt.Errorf("--pr must be a positive number, not %d", o.pr)
		}

		return scanPR(ctx, o.dir, o.pr, o.repo)
	case o.repo != "":
		return nil, "", errors.New("--repo needs --pr")
	case o.description == "" && o.revRange == "":
		return nil, "", errors.New("nothing to scan: give --pr, or --description and/or --range")
	}

	if o.description != "" {
		m, err := scanDescription(ctx, o.description, stdin)
		if err != nil {
			return nil, "", err
		}

		matches = append(matches, m...)
	}

	if o.revRange != "" {
		m, err := scanCommits(ctx, o.dir, o.revRange)
		if err != nil {
			return nil, "", err
		}

		matches = append(matches, m...)
	}

	return matches, "", nil
}

func scanDescription(ctx context.Context, path string, stdin io.Reader) ([]Match, error) {
	if path == "-" {
		return scanReader(ctx, stdin)
	}

	b, err := os.ReadFile(path) //nolint:gosec // G304: the path is the caller's own argument
	if err != nil {
		return nil, fmt.Errorf("open description: %w", err)
	}

	return Scan(descriptionSource, string(b)), nil
}
