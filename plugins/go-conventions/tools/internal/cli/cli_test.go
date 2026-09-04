package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/jhoblitt/conventions-claude/plugins/go-conventions/tools/internal/cli"
)

const repo = "testdata/repo"

// pluginRoot is the go-conventions plugin directory, three levels above this
// package, the value the launcher exports as CLAUDE_PLUGIN_ROOT.
func pluginRoot() string {
	GinkgoHelper()

	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	Expect(err).NotTo(HaveOccurred())

	return root
}

func execute(ctx context.Context, args ...string) (string, string, error) {
	var stdout, stderr bytes.Buffer
	err := cli.Run(ctx, args, strings.NewReader(""), &stdout, &stderr)

	return stdout.String(), stderr.String(), err
}

var _ = Describe("Run", func() {
	It("renders markdown by default", func(ctx SpecContext) {
		stdout, _, err := execute(ctx, repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(stdout).To(HavePrefix("| Area | Check | Status | Phase | Current | Canon | Fix |\n"))
		Expect(stdout).To(ContainSubstring("tooling"))
	})

	It("renders JSON on --json", func(ctx SpecContext) {
		stdout, _, err := execute(ctx, "--json", repo)
		Expect(err).NotTo(HaveOccurred())

		var report struct {
			Rows      []map[string]any `json:"rows"`
			Gaps      int              `json:"gaps"`
			Tooling   int              `json:"tooling"`
			Migration int              `json:"migration"`
			OK        int              `json:"ok"`
			Skipped   int              `json:"skipped"`
		}
		Expect(json.Unmarshal([]byte(stdout), &report)).To(Succeed())
		Expect(report.Rows).NotTo(BeEmpty())
		Expect(report.Gaps).To(Equal(report.Tooling + report.Migration))
	})

	It("rejects two renderings at once", func(ctx SpecContext) {
		_, _, err := execute(ctx, "--json", "--markdown", repo)
		Expect(err).To(HaveOccurred())
	})

	It("rejects a directory that is not a Go repository root", func(ctx SpecContext) {
		_, _, err := execute(ctx, "testdata/not-a-repo")
		Expect(err).To(MatchError(ContainSubstring("go.mod")))
	})

	It("rejects an unknown flag", func(ctx SpecContext) {
		_, _, err := execute(ctx, "--nonsense", repo)
		Expect(err).To(HaveOccurred())
	})

	It("prints one path per line for --files", func(ctx SpecContext) {
		stdout, _, err := execute(ctx, "--files", "cli", repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(strings.Split(strings.TrimSuffix(stdout, "\n"), "\n")).To(Equal([]string{"cmd/tiny/main.go"}))
	})

	It("rejects an unknown --files area", func(ctx SpecContext) {
		_, _, err := execute(ctx, "--files", "nonsense", repo)
		Expect(err).To(MatchError(ContainSubstring("nonsense")))
	})

	It("emits the lint config from the plugin's template", func(ctx SpecContext) {
		GinkgoT().Setenv("CLAUDE_PLUGIN_ROOT", pluginRoot())

		stdout, _, err := execute(ctx, "--emit-golangci", repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(stdout).To(ContainSubstring(`version: "2"`))
		Expect(stdout).To(ContainSubstring("example.com/tiny"))
	})

	It("fails --emit-golangci when CLAUDE_PLUGIN_ROOT is unset", func(ctx SpecContext) {
		GinkgoT().Setenv("CLAUDE_PLUGIN_ROOT", "")

		_, _, err := execute(ctx, "--emit-golangci", repo)
		Expect(err).To(MatchError(ContainSubstring("CLAUDE_PLUGIN_ROOT")))
	})

	It("skips the template-backed rows when CLAUDE_PLUGIN_ROOT is unset", func(ctx SpecContext) {
		GinkgoT().Setenv("CLAUDE_PLUGIN_ROOT", "")

		stdout, _, err := execute(ctx, repo)
		Expect(err).NotTo(HaveOccurred())
		Expect(stdout).To(ContainSubstring("CLAUDE_PLUGIN_ROOT unset"))
	})
})
