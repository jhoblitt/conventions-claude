package audit_test

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/audit"
	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/audit/auditfakes"
)

func protective() audit.Ruleset {
	return audit.Ruleset{
		Name:        "protect-default-branch",
		Target:      "branch",
		Enforcement: "active",
		Conditions: audit.Conditions{
			RefName: audit.RefName{Include: []string{"~DEFAULT_BRANCH"}},
		},
		Rules: []audit.Rule{{Type: "deletion"}, {Type: "non_fast_forward"}},
	}
}

var _ = Describe("the ruleset check", func() {
	var (
		lookup *auditfakes.FakeRulesetLookup
		dir    string
	)

	BeforeEach(func() {
		lookup = &auditfakes.FakeRulesetLookup{}
		dir = filepath.Join("testdata", "empty")
	})

	rulesetRow := func(report audit.Report) audit.Row {
		GinkgoHelper()

		for _, row := range report.Rows {
			if row.Area == "ruleset" {
				return row
			}
		}
		Fail("no ruleset row in the report")
		return audit.Row{}
	}

	It("skips the check and never calls the lookup without --remote", func(ctx SpecContext) {
		report, err := audit.Run(ctx, audit.Options{Dir: dir, Lookup: lookup})
		Expect(err).NotTo(HaveOccurred())

		Expect(rulesetRow(report).Status).To(Equal(audit.StatusSkipped))
		Expect(lookup.RulesetsCallCount()).To(Equal(0))
	})

	It("passes when a protective ruleset is active on the default branch", func(ctx SpecContext) {
		lookup.RulesetsReturns([]audit.Ruleset{protective()}, nil)

		report, err := audit.Run(ctx, audit.Options{Dir: dir, Remote: true, Lookup: lookup})
		Expect(err).NotTo(HaveOccurred())

		row := rulesetRow(report)
		Expect(row.Status).To(Equal(audit.StatusOK))
		Expect(row.Current).To(ContainSubstring("protect-default-branch"))
		Expect(row.Fix).To(BeEmpty())

		Expect(lookup.RulesetsCallCount()).To(Equal(1))
		_, gotDir := lookup.RulesetsArgsForCall(0)
		Expect(gotDir).To(Equal(dir))
	})

	DescribeTable("reports a gap when the ruleset falls short",
		func(ctx SpecContext, rs audit.Ruleset, wantCurrent string) {
			lookup.RulesetsReturns([]audit.Ruleset{rs}, nil)

			report, err := audit.Run(ctx, audit.Options{Dir: dir, Remote: true, Lookup: lookup})
			Expect(err).NotTo(HaveOccurred())

			row := rulesetRow(report)
			Expect(row.Status).To(Equal(audit.StatusGap))
			Expect(row.Current).To(ContainSubstring(wantCurrent))
			Expect(row.Fix).NotTo(BeEmpty())
		},
		Entry("not enforced", func() audit.Ruleset {
			rs := protective()
			rs.Enforcement = "evaluate"
			return rs
		}(), "evaluate"),
		Entry("aimed at tags", func() audit.Ruleset {
			rs := protective()
			rs.Target = "tag"
			return rs
		}(), "tag"),
		Entry("missing non_fast_forward", func() audit.Ruleset {
			rs := protective()
			rs.Rules = []audit.Rule{{Type: "deletion"}}
			return rs
		}(), "deletion"),
		Entry("aimed at another branch", func() audit.Ruleset {
			rs := protective()
			rs.Conditions = audit.Conditions{RefName: audit.RefName{Include: []string{"refs/heads/legacy"}}}
			return rs
		}(), "refs/heads/legacy"),
	)

	It("reports no rulesets at all as a gap", func(ctx SpecContext) {
		lookup.RulesetsReturns(nil, nil)

		report, err := audit.Run(ctx, audit.Options{Dir: dir, Remote: true, Lookup: lookup})
		Expect(err).NotTo(HaveOccurred())

		row := rulesetRow(report)
		Expect(row.Status).To(Equal(audit.StatusGap))
		Expect(row.Current).To(Equal("no rulesets on the repository"))
	})

	It("cannot pass a ruleset whose conditions and rules are missing", func(ctx SpecContext) {
		bare := protective()
		bare.Conditions = audit.Conditions{}
		bare.Rules = nil
		lookup.RulesetsReturns([]audit.Ruleset{bare}, nil)

		report, err := audit.Run(ctx, audit.Options{Dir: dir, Remote: true, Lookup: lookup})
		Expect(err).NotTo(HaveOccurred())

		row := rulesetRow(report)
		Expect(row.Status).To(Equal(audit.StatusGap))
		Expect(row.Current).To(ContainSubstring("no refs"))
		Expect(row.Current).To(ContainSubstring("no rules"))
	})

	It("records a lookup failure as a gap and still succeeds", func(ctx SpecContext) {
		lookup.RulesetsReturns(nil, errors.New("exit status 4"))

		report, err := audit.Run(ctx, audit.Options{Dir: dir, Remote: true, Lookup: lookup})
		Expect(err).NotTo(HaveOccurred())

		row := rulesetRow(report)
		Expect(row.Status).To(Equal(audit.StatusGap))
		Expect(row.Current).To(HavePrefix("gh api failed: "))
		Expect(row.Current).To(ContainSubstring("exit status 4"))
	})
})

var _ = Describe("the gh ruleset lookup", func() {
	It("keeps the list entry when one ruleset's detail cannot be read", func(ctx SpecContext) {
		list := []audit.Ruleset{
			{ID: 1, Name: "org-wide", Target: "branch", Enforcement: "active"},
			{ID: 2, Name: "protect-default-branch", Target: "branch", Enforcement: "active"},
		}

		calls := 0
		api := func(_ context.Context, _ string, into any, args ...string) error {
			calls++
			if strings.HasSuffix(args[len(args)-1], "/1") {
				return errors.New("HTTP 403: Resource not accessible by personal access token")
			}
			full, ok := into.(*audit.Ruleset)
			Expect(ok).To(BeTrue())
			*full = protective()

			return nil
		}

		got := audit.RulesetDetail(ctx, slog.New(slog.DiscardHandler), api, list)

		Expect(calls).To(Equal(2))
		Expect(got).To(HaveLen(2))
		Expect(got[0].Name).To(Equal("org-wide"))
		Expect(got[0].Rules).To(BeEmpty())
		Expect(got[1].Rules).To(HaveLen(2))
	})
})
