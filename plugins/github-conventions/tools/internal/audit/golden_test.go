package audit_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/audit"
)

func golden(path, got string) {
	GinkgoHelper()

	if *update {
		Expect(os.WriteFile(path, []byte(got), 0o644)).To(Succeed())
		return
	}
	want, err := os.ReadFile(path)
	Expect(err).NotTo(HaveOccurred())
	Expect(got).To(Equal(string(want)))
}

var _ = Describe("Run", func() {
	DescribeTable("matches the golden rendering of a fixture repository",
		func(ctx SpecContext, fixture string) {
			dir := filepath.Join("testdata", fixture)

			report, err := audit.Run(ctx, audit.Options{Dir: dir})
			Expect(err).NotTo(HaveOccurred())

			golden(filepath.Join(dir, "expected.md"), report.Markdown())

			data, err := report.JSON()
			Expect(err).NotTo(HaveOccurred())
			golden(filepath.Join(dir, "expected.json"), string(data))
		},
		Entry("a repository that already meets the canon", "compliant"),
		Entry("a repository that predates it", "unpinned-legacy"),
		Entry("a repository with nothing at all", "empty"),
	)

	It("counts every row exactly once", func(ctx SpecContext) {
		report, err := audit.Run(ctx, audit.Options{Dir: filepath.Join("testdata", "unpinned-legacy")})
		Expect(err).NotTo(HaveOccurred())
		Expect(report.Gaps + report.OK + report.Skipped).To(Equal(len(report.Rows)))
	})

	It("leaves fix empty on every ok row", func(ctx SpecContext) {
		report, err := audit.Run(ctx, audit.Options{Dir: filepath.Join("testdata", "compliant")})
		Expect(err).NotTo(HaveOccurred())
		for _, row := range report.Rows {
			if row.Status == audit.StatusOK {
				Expect(row.Fix).To(BeEmpty(), "row %s/%s", row.Area, row.Check)
			}
		}
	})

	It("reports an unreadable directory as an error", func(ctx SpecContext) {
		_, err := audit.Run(ctx, audit.Options{Dir: filepath.Join("testdata", "no-such-repo")})
		Expect(err).To(HaveOccurred())
	})
})
