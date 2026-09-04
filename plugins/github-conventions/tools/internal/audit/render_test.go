package audit_test

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/audit"
)

var _ = Describe("Report.Markdown", func() {
	It("escapes a pipe and folds a newline so one cell cannot break the table", func() {
		report := audit.Report{
			Rows: []audit.Row{{
				Area:    "workflows",
				Check:   "pinned:ci.yml",
				Status:  audit.StatusGap,
				Current: "unpinned: a|b\nand c",
				Canon:   "canon",
				Fix:     "fix",
			}},
			Gaps: 1,
		}

		got := report.Markdown()
		Expect(got).To(ContainSubstring(`| unpinned: a\|b and c |`))
		Expect(strings.Count(got, "\n")).To(Equal(5))
		Expect(got).To(HaveSuffix("1 gaps, 0 ok, 0 skipped\n"))
	})
})
