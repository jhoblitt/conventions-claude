package audit_test

import (
	"errors"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/audit"
	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/audit/auditfakes"
)

var _ = Describe("the delete-branch-on-merge check", func() {
	var (
		lookup *auditfakes.FakeRemoteLookup
		dir    string
	)

	BeforeEach(func() {
		lookup = &auditfakes.FakeRemoteLookup{}
		dir = filepath.Join("testdata", "empty")
	})

	repositoryRow := func(report audit.Report) audit.Row {
		GinkgoHelper()

		for _, row := range report.Rows {
			if row.Area == "repository" {
				return row
			}
		}
		Fail("no repository row in the report")
		return audit.Row{}
	}

	It("skips the check and never calls the lookup without --remote", func(ctx SpecContext) {
		report, err := audit.Run(ctx, audit.Options{Dir: dir, Lookup: lookup})
		Expect(err).NotTo(HaveOccurred())

		row := repositoryRow(report)
		Expect(row.Status).To(Equal(audit.StatusSkipped))
		Expect(row.Current).To(Equal("not checked (--remote not given)"))
		Expect(row.Fix).To(BeEmpty())
		Expect(lookup.RepositoryCallCount()).To(Equal(0))
	})

	It("renders last, after the ruleset row", func(ctx SpecContext) {
		report, err := audit.Run(ctx, audit.Options{Dir: dir, Lookup: lookup})
		Expect(err).NotTo(HaveOccurred())

		last := report.Rows[len(report.Rows)-1]
		Expect(last.Area).To(Equal("repository"))
		Expect(last.Check).To(Equal("delete-branch-on-merge"))
		Expect(report.Rows[len(report.Rows)-2].Area).To(Equal("ruleset"))
	})

	It("passes when a merged branch is deleted", func(ctx SpecContext) {
		lookup.RepositoryReturns(audit.Repository{DeleteBranchOnMerge: true}, nil)

		report, err := audit.Run(ctx, audit.Options{Dir: dir, Remote: true, Lookup: lookup})
		Expect(err).NotTo(HaveOccurred())

		row := repositoryRow(report)
		Expect(row.Status).To(Equal(audit.StatusOK))
		Expect(row.Current).To(Equal("delete_branch_on_merge: true"))
		Expect(row.Fix).To(BeEmpty())

		Expect(lookup.RepositoryCallCount()).To(Equal(1))
		_, gotDir := lookup.RepositoryArgsForCall(0)
		Expect(gotDir).To(Equal(dir))
	})

	It("reports a repository that keeps merged branches as a gap", func(ctx SpecContext) {
		lookup.RepositoryReturns(audit.Repository{FullName: "octo-dev/demo-tool", DeleteBranchOnMerge: false}, nil)

		report, err := audit.Run(ctx, audit.Options{Dir: dir, Remote: true, Lookup: lookup})
		Expect(err).NotTo(HaveOccurred())

		row := repositoryRow(report)
		Expect(row.Status).To(Equal(audit.StatusGap))
		Expect(row.Current).To(Equal("delete_branch_on_merge: false"))
		Expect(row.Fix).To(Equal("gh repo edit octo-dev/demo-tool --delete-branch-on-merge"))
	})

	It("records a lookup failure as a gap and still succeeds", func(ctx SpecContext) {
		lookup.RepositoryReturns(audit.Repository{}, errors.New("exit status 4"))

		report, err := audit.Run(ctx, audit.Options{Dir: dir, Remote: true, Lookup: lookup})
		Expect(err).NotTo(HaveOccurred())

		row := repositoryRow(report)
		Expect(row.Status).To(Equal(audit.StatusGap))
		Expect(row.Current).To(HavePrefix("gh api failed: "))
		Expect(row.Current).To(ContainSubstring("exit status 4"))
		Expect(row.Fix).To(Equal("gh repo edit --delete-branch-on-merge"))
	})
})
