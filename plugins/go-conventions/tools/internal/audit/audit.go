// Package audit checks a Go repository against the go-conventions canon and
// renders the result as markdown or JSON.
//
// Everything it reports is read statically from the tree: go.mod through
// [modfile], imports through go/parser, YAML through go.yaml.in/yaml/v3, and
// the Makefile and workflow run: blocks as text. Nothing is executed.
package audit

import (
	"errors"
	"log/slog"
	"path/filepath"
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

// Phase says who closes a gap.
type Phase string

// The phases a [Row] can carry: converge applies a tooling gap itself, a
// migration gap changes code and waits for the user to name the area, and a
// row that is not a gap has no phase.
const (
	PhaseTooling   Phase = "tooling"
	PhaseMigration Phase = "migration"
	PhaseNone      Phase = "none"
)

// Row is one check's verdict, the canon it was measured against, and the
// action that would close the gap.
type Row struct {
	Area    string `json:"area"`
	Check   string `json:"check"`
	Status  Status `json:"status"`
	Phase   Phase  `json:"phase"`
	Current string `json:"current"`
	Canon   string `json:"canon"`
	Fix     string `json:"fix"`
}

// Report is the whole audit: every row, plus a tally by status and by phase.
type Report struct {
	Rows      []Row `json:"rows"`
	Gaps      int   `json:"gaps"`
	Tooling   int   `json:"tooling"`
	Migration int   `json:"migration"`
	OK        int   `json:"ok"`
	Skipped   int   `json:"skipped"`
}

// Options selects what to audit and where the canon templates are.
type Options struct {
	// Dir is the Go repository root to audit; it must hold a go.mod.
	Dir string
	// PluginRoot is the go-conventions plugin directory, CLAUDE_PLUGIN_ROOT.
	// Empty means unset: the checks that read a template are skipped, and
	// [Golangci] fails.
	PluginRoot string
	// Logger receives operational warnings; nil discards them.
	Logger *slog.Logger
}

// ErrNoModule reports a directory that is not a Go repository root.
var ErrNoModule = errors.New("no go.mod")

// Run audits the repository at opts.Dir. It fails only when the repository
// cannot be read; a check that cannot be answered becomes a gap or a skip.
func Run(opts Options) (Report, error) {
	repo, err := scan(opts)
	if err != nil {
		return Report{}, err
	}

	report := Report{Rows: repo.rows()}
	for _, row := range report.Rows {
		switch row.Status {
		case StatusGap:
			report.Gaps++
		case StatusOK:
			report.OK++
		case StatusSkipped:
			report.Skipped++
		}
		switch row.Phase {
		case PhaseTooling:
			report.Tooling++
		case PhaseMigration:
			report.Migration++
		case PhaseNone:
		}
	}

	return report, nil
}

// Golangci renders the canon lint config for the repository at opts.Dir: the
// template with its module path filled in, minus every depguard deny entry the
// repository's current imports would trip on.
func Golangci(opts Options) (string, error) {
	repo, err := scan(opts)
	if err != nil {
		return "", err
	}

	return repo.emitGolangci()
}

// Files lists, sorted and unique, the paths a migration of area touches,
// relative to opts.Dir.
func Files(opts Options, area string) ([]string, error) {
	repo, err := scan(opts)
	if err != nil {
		return nil, err
	}

	return repo.files(area)
}

func okRow(area, check, current, canon string) Row {
	return Row{Area: area, Check: check, Status: StatusOK, Phase: PhaseNone, Current: current, Canon: canon}
}

func gapRow(phase Phase, area, check, current, canon, fix string) Row {
	return Row{Area: area, Check: check, Status: StatusGap, Phase: phase, Current: current, Canon: canon, Fix: fix}
}

func skipRow(area, check, current, canon string) Row {
	return Row{Area: area, Check: check, Status: StatusSkipped, Phase: PhaseNone, Current: current, Canon: canon}
}

func templateMissing(area, check, canon, reason string) Row {
	return skipRow(area, check, "templates unreadable: "+reason, canon)
}

func rel(dir, path string) string {
	got, err := filepath.Rel(dir, path)
	if err != nil {
		return filepath.ToSlash(path)
	}

	return filepath.ToSlash(got)
}
