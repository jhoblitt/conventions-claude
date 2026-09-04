package cli_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/cli"
)

var _ = Describe("Run", func() {
	var (
		stdout, stderr *bytes.Buffer
		fixture        string
	)

	BeforeEach(func() {
		stdout, stderr = &bytes.Buffer{}, &bytes.Buffer{}
		fixture = filepath.Join("..", "audit", "testdata", "empty")
	})

	run := func(ctx SpecContext, args ...string) error {
		return cli.Run(ctx, args, strings.NewReader(""), stdout, stderr)
	}

	It("writes markdown by default", func(ctx SpecContext) {
		Expect(run(ctx, fixture)).To(Succeed())

		Expect(stdout.String()).To(HavePrefix("| Area | Check | Status | Current | Canon | Fix |\n"))
		Expect(stdout.String()).To(HaveSuffix("14 gaps, 0 ok, 1 skipped\n"))
		Expect(stderr.String()).To(BeEmpty())
	})

	It("writes JSON under --json", func(ctx SpecContext) {
		Expect(run(ctx, "--json", fixture)).To(Succeed())

		var report struct {
			Rows []struct {
				Area   string `json:"area"`
				Status string `json:"status"`
			} `json:"rows"`
			Gaps    int `json:"gaps"`
			OK      int `json:"ok"`
			Skipped int `json:"skipped"`
		}
		Expect(json.Unmarshal(stdout.Bytes(), &report)).To(Succeed())
		Expect(report.Rows).To(HaveLen(15))
		Expect(report.Gaps).To(Equal(14))
		Expect(report.Skipped).To(Equal(1))
	})

	It("audits the working directory when no dir is given", func(ctx SpecContext) {
		Expect(run(ctx, "--markdown")).To(Succeed())
		Expect(stdout.String()).To(ContainSubstring("| Area | Check | Status |"))
	})

	It("refuses --json together with --markdown", func(ctx SpecContext) {
		Expect(run(ctx, "--json", "--markdown", fixture)).NotTo(Succeed())
		Expect(stdout.String()).To(BeEmpty())
	})

	It("refuses a second directory argument", func(ctx SpecContext) {
		Expect(run(ctx, fixture, fixture)).NotTo(Succeed())
	})

	It("fails when the directory does not exist", func(ctx SpecContext) {
		err := run(ctx, filepath.Join(fixture, "nowhere"))
		Expect(err).To(HaveOccurred())
		Expect(stdout.String()).To(BeEmpty())
	})
})
