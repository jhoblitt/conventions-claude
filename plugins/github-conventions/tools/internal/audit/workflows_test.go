package audit_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/audit"
)

func repoWith(files map[string]string) string {
	GinkgoHelper()

	dir := GinkgoT().TempDir()
	for name, body := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		Expect(os.MkdirAll(filepath.Dir(path), 0o750)).To(Succeed())
		Expect(os.WriteFile(path, []byte(body), 0o600)).To(Succeed())
	}

	return dir
}

func rowFor(report audit.Report, area, check string) audit.Row {
	GinkgoHelper()

	for _, row := range report.Rows {
		if row.Area == area && row.Check == check {
			return row
		}
	}
	Fail("no " + area + "/" + check + " row in the report")

	return audit.Row{}
}

var _ = Describe("the workflow checks", func() {
	audited := func(ctx SpecContext, name, body string) audit.Report {
		GinkgoHelper()

		report, err := audit.Run(ctx, audit.Options{Dir: repoWith(map[string]string{".github/workflows/" + name: body})})
		Expect(err).NotTo(HaveOccurred())

		return report
	}

	It("does not demand timeout-minutes from a job that calls a reusable workflow", func(ctx SpecContext) {
		report := audited(ctx, "call.yml", `
name: call
on: [push]
jobs:
  reusable:
    uses: ./.github/workflows/lib.yml
  build:
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - run: echo hi
`)

		row := rowFor(report, "workflows", "timeout:call.yml")
		Expect(row.Status).To(Equal(audit.StatusOK))
		Expect(row.Current).To(Equal("1 job(s), all with timeout-minutes"))
	})

	It("treats a workflow with no jobs as vacuously ok", func(ctx SpecContext) {
		report := audited(ctx, "empty.yml", "name: empty\non: [push]\n")

		row := rowFor(report, "workflows", "timeout:empty.yml")
		Expect(row.Status).To(Equal(audit.StatusOK))
		Expect(row.Current).To(Equal("no jobs"))
		Expect(row.Fix).To(BeEmpty())
	})

	It("reads a quoted uses: value", func(ctx SpecContext) {
		report := audited(ctx, "quoted.yml", `
name: quoted
on: [push]
jobs:
  scan:
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - uses: "actions/checkout@0000000000000000000000000000000000000001" # v7
        with:
          persist-credentials: false
      - uses: 'github/codeql-action/init@0000000000000000000000000000000000000002' # v4
`)

		Expect(rowFor(report, "workflows", "pinned:quoted.yml").Status).To(Equal(audit.StatusOK))

		checkout := rowFor(report, "workflows", "checkout-credentials:quoted.yml")
		Expect(checkout.Status).To(Equal(audit.StatusOK))
		Expect(checkout.Current).To(Equal("persist-credentials: false on all 1"))

		Expect(rowFor(report, "security", "codeql").Status).To(Equal(audit.StatusOK))
	})

	It("does not count a uses: line inside a run: block as an action", func(ctx SpecContext) {
		report := audited(ctx, "heredoc.yml", `
name: heredoc
on: [push]
jobs:
  docs:
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - run: |
          cat <<'YAML'
          uses: github/codeql-action/init@v4
          uses: actions/checkout@v3
          YAML
`)

		Expect(rowFor(report, "security", "codeql").Status).To(Equal(audit.StatusGap))

		checkout := rowFor(report, "workflows", "checkout-credentials:heredoc.yml")
		Expect(checkout.Status).To(Equal(audit.StatusOK))
		Expect(checkout.Current).To(Equal("no actions/checkout steps"))
	})

	DescribeTable("reads the top-level permissions key",
		func(ctx SpecContext, permissions string, want audit.Status, wantCurrent string) {
			report := audited(ctx, "perms.yml", `
name: perms
on: [push]
`+permissions+`
jobs:
  build:
    runs-on: ubuntu-latest
    timeout-minutes: 5
    steps:
      - run: echo hi
`)

			row := rowFor(report, "workflows", "permissions:perms.yml")
			Expect(row.Status).To(Equal(want))
			Expect(row.Current).To(Equal(wantCurrent))
		},
		Entry("a map with a write scope is a gap", "permissions:\n  contents: read\n  packages: write",
			audit.StatusGap, "contents: read, packages: write"),
		Entry("a read-only map is ok", "permissions:\n  contents: read",
			audit.StatusOK, "contents: read"),
		Entry("read-all is ok", "permissions: read-all", audit.StatusOK, "permissions: read-all"),
		Entry("write-all is a gap", "permissions: write-all", audit.StatusGap, "permissions: write-all"),
		Entry("an empty map grants nothing", "permissions: {}", audit.StatusOK, "no scopes granted"),
	)
})
