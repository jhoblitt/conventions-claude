package audit

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Markdown renders the report as a table followed by the summary line.
func (r Report) Markdown() string {
	var b strings.Builder

	b.WriteString("| Area | Check | Status | Phase | Current | Canon | Fix |\n")
	b.WriteString("| --- | --- | --- | --- | --- | --- | --- |\n")
	for _, row := range r.Rows {
		fmt.Fprintf(&b, "| %s | %s | %s | %s | %s | %s | %s |\n",
			cell(row.Area), cell(row.Check), cell(string(row.Status)), cell(string(row.Phase)),
			cell(row.Current), cell(row.Canon), cell(row.Fix))
	}
	fmt.Fprintf(&b, "\n%s\n", r.summary())

	return b.String()
}

func (r Report) summary() string {
	return fmt.Sprintf("%d gaps (%d tooling, %d migration), %d ok, %d skipped",
		r.Gaps, r.Tooling, r.Migration, r.OK, r.Skipped)
}

// JSON renders the report as indented JSON with a trailing newline.
func (r Report) JSON() ([]byte, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal report: %w", err)
	}

	return append(data, '\n'), nil
}

var cellEscaper = strings.NewReplacer("|", `\|`, "\n", " ")

func cell(s string) string { return cellEscaper.Replace(s) }
