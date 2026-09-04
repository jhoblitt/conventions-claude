// Package audit checks a repository against the github-conventions canon and
// renders the result as markdown or JSON.
package audit

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
)

// Status is a check's verdict.
type Status string

// The verdicts a [Row] can carry. A gap is data, not a failure: an audit that
// ran to completion reports its gaps and still succeeds.
const (
	StatusOK      Status = "ok"
	StatusGap     Status = "gap"
	StatusSkipped Status = "skipped"
)

// Row is one check's verdict, the canon it was measured against, and the
// action that would close the gap.
type Row struct {
	Area    string `json:"area"`
	Check   string `json:"check"`
	Status  Status `json:"status"`
	Current string `json:"current"`
	Canon   string `json:"canon"`
	Fix     string `json:"fix"`
}

// Report is the whole audit: every row, plus a tally by status.
type Report struct {
	Rows    []Row `json:"rows"`
	Gaps    int   `json:"gaps"`
	OK      int   `json:"ok"`
	Skipped int   `json:"skipped"`
}

// Options selects what Run audits and how it reaches GitHub.
type Options struct {
	// Dir is the repository root to audit.
	Dir string
	// Remote enables the checks that need the GitHub API.
	Remote bool
	// Lookup fetches rulesets; nil means the gh CLI.
	Lookup RulesetLookup
	// Logger receives operational warnings; nil discards them.
	Logger *slog.Logger
}

// Run audits the repository at opts.Dir. It fails only when the repository
// cannot be read; a check that cannot be answered becomes a gap.
func Run(ctx context.Context, opts Options) (Report, error) {
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.DiscardHandler)
	}

	info, err := os.Stat(opts.Dir)
	if err != nil {
		return Report{}, fmt.Errorf("read repository: %w", err)
	}
	if !info.IsDir() {
		return Report{}, fmt.Errorf("read repository: %s is not a directory", opts.Dir)
	}

	workflows, err := loadWorkflows(opts.Dir)
	if err != nil {
		return Report{}, err
	}

	rows, err := staticRows(opts.Dir, workflows)
	if err != nil {
		return Report{}, err
	}
	rows = append(rows, rulesetRow(ctx, opts))

	report := Report{Rows: rows}
	for _, row := range rows {
		switch row.Status {
		case StatusGap:
			report.Gaps++
		case StatusOK:
			report.OK++
		case StatusSkipped:
			report.Skipped++
		}
	}

	return report, nil
}

func staticRows(dir string, workflows []workflow) ([]Row, error) {
	rows := make([]Row, 0, 16+5*len(workflows))
	for _, section := range []func(string) ([]Row, error){licenseRows, readmeRows, dependabotRows} {
		got, err := section(dir)
		if err != nil {
			return nil, err
		}
		rows = append(rows, got...)
	}
	rows = append(rows, perWorkflowRows(workflows)...)
	rows = append(rows, securityRows(workflows)...)
	rows = append(rows, workflowLintRows(workflows)...)

	commitlint, err := commitlintRows(dir, workflows)
	if err != nil {
		return nil, err
	}

	return append(rows, commitlint...), nil
}

func okRow(area, check, current, canon string) Row {
	return Row{Area: area, Check: check, Status: StatusOK, Current: current, Canon: canon}
}

func gapRow(area, check, current, canon, fix string) Row {
	return Row{Area: area, Check: check, Status: StatusGap, Current: current, Canon: canon, Fix: fix}
}

func readFile(path string) (text string, found bool, err error) {
	data, err := os.ReadFile(path) //nolint:gosec // the path is under the directory the caller asked to audit
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("read %s: %w", path, err)
	}

	return string(data), true, nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat %s: %w", path, err)
	}

	return true, nil
}
