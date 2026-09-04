package audit_test

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.yaml.in/yaml/v3"

	"github.com/jhoblitt/conventions-claude/plugins/go-conventions/tools/internal/audit"
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

// fixtures are the golden repositories; every invariant spec runs over all of
// them, and each has an expected.md and expected.json beside it.
var fixtures = []string{
	"compliant", "legacy-daemon", "kubebuilder", "kubebuilder-stock",
	"stdlib-cli", "library", "quiet-cli", "slog-default",
}

func options(fixture string) audit.Options {
	return audit.Options{Dir: filepath.Join("testdata", fixture), PluginRoot: pluginRoot}
}

func run(fixture string) audit.Report {
	GinkgoHelper()

	report, err := audit.Run(options(fixture))
	Expect(err).NotTo(HaveOccurred())

	return report
}

func rowFor(report audit.Report, area, check string) audit.Row {
	GinkgoHelper()

	for _, row := range report.Rows {
		if row.Area == area && row.Check == check {
			return row
		}
	}
	Fail("no row " + area + "/" + check)

	return audit.Row{}
}

var _ = Describe("Run", func() {
	DescribeTable("matches the golden rendering of a fixture repository",
		func(fixture string) {
			dir := filepath.Join("testdata", fixture)
			report := run(fixture)

			golden(filepath.Join(dir, "expected.md"), report.Markdown())

			data, err := report.JSON()
			Expect(err).NotTo(HaveOccurred())
			golden(filepath.Join(dir, "expected.json"), string(data))
		},
		Entry("a repository rendered from the canon templates", "compliant"),
		Entry("a flat daemon that predates the canon", "legacy-daemon"),
		Entry("a controller with a standalone zap logger", "kubebuilder"),
		Entry("a stock kubebuilder scaffold", "kubebuilder-stock"),
		Entry("a stdlib-flag CLI with no tooling", "stdlib-cli"),
		Entry("a library with no binary", "library"),
		Entry("a command that logs nothing", "quiet-cli"),
		Entry("a command on slog with no handler of its own", "slog-default"),
	)

	It("finds nothing to change in a repository rendered from the canon templates", func() {
		report := run("compliant")
		Expect(report.Gaps).To(BeZero(), "compliant reports: %s", report.Markdown())
		Expect(report.Skipped).To(BeZero())
	})

	It("counts every row exactly once", func() {
		report := run("legacy-daemon")
		Expect(report.Gaps + report.OK + report.Skipped).To(Equal(len(report.Rows)))
		Expect(report.Tooling + report.Migration).To(Equal(report.Gaps))
	})

	It("leaves fix empty unless the row is a gap", func() {
		for _, fixture := range fixtures {
			for _, row := range run(fixture).Rows {
				if row.Status != audit.StatusGap {
					Expect(row.Fix).To(BeEmpty(), "%s row %s/%s", fixture, row.Area, row.Check)
				}
			}
		}
	})

	It("phases every gap and only gaps", func() {
		for _, fixture := range fixtures {
			for _, row := range run(fixture).Rows {
				if row.Status == audit.StatusGap {
					Expect(row.Phase).To(BeElementOf(audit.PhaseTooling, audit.PhaseMigration),
						"%s row %s/%s", fixture, row.Area, row.Check)

					continue
				}
				Expect(row.Phase).To(Equal(audit.PhaseNone), "%s row %s/%s", fixture, row.Area, row.Check)
			}
		}
	})

	It("reports a directory with no go.mod as an error", func() {
		_, err := audit.Run(audit.Options{Dir: "testdata", PluginRoot: pluginRoot})
		Expect(err).To(HaveOccurred())
	})

	It("reports an unreadable directory as an error", func() {
		_, err := audit.Run(options("no-such-repo"))
		Expect(err).To(HaveOccurred())
	})

	Describe("the toolchain row", func() {
		It("passes a minor-only directive with no toolchain line", func() {
			Expect(rowFor(run("compliant"), "toolchain", "go-directive").Status).To(Equal(audit.StatusOK))
		})

		It("fails a patch directive", func() {
			row := rowFor(run("stdlib-cli"), "toolchain", "go-directive")
			Expect(row.Status).To(Equal(audit.StatusGap))
			Expect(row.Current).To(ContainSubstring("1.26.1"))
		})

		It("fails a toolchain line even on the canon directive", func() {
			row := rowFor(run("kubebuilder"), "toolchain", "go-directive")
			Expect(row.Status).To(Equal(audit.StatusGap))
			Expect(row.Current).To(ContainSubstring("toolchain"))
		})
	})

	Describe("the lint rows", func() {
		It("skips the detail rows when the config is a v1 shape", func() {
			report := run("kubebuilder")
			Expect(rowFor(report, "lint", "config").Status).To(Equal(audit.StatusGap))
			Expect(rowFor(report, "lint", "config").Fix).To(ContainSubstring("--emit-golangci"))
			for _, check := range []string{"linters", "formatters", "max-issues"} {
				Expect(rowFor(report, "lint", check).Status).To(Equal(audit.StatusSkipped), check)
			}
		})

		It("names the missing linters when the config is v2 but thin", func() {
			report := run("legacy-daemon")
			Expect(rowFor(report, "lint", "config").Status).To(Equal(audit.StatusOK))
			row := rowFor(report, "lint", "linters")
			Expect(row.Status).To(Equal(audit.StatusGap))
			Expect(row.Current).To(ContainSubstring("depguard"))
			Expect(row.Current).NotTo(ContainSubstring("govet"))
		})
	})

	Describe("the layout rows", func() {
		It("accepts cmd/<name>", func() {
			Expect(rowFor(run("stdlib-cli"), "layout", "cmd").Status).To(Equal(audit.StatusOK))
		})

		It("names a main package outside cmd/<name>", func() {
			row := rowFor(run("legacy-daemon"), "layout", "cmd")
			Expect(row.Status).To(Equal(audit.StatusGap))
			Expect(row.Phase).To(Equal(audit.PhaseMigration))
			Expect(row.Current).To(HaveSuffix(": ."))
		})
	})

	Describe("the dependency rows", func() {
		It("defers the CLI and logging rows to controller-runtime", func() {
			report := run("kubebuilder")
			for _, check := range []string{"cli", "config", "logging"} {
				row := rowFor(report, "deps", check)
				Expect(row.Status).To(Equal(audit.StatusSkipped), check)
				Expect(row.Current).To(Equal("controller-runtime"), check)
			}
			Expect(rowFor(report, "logging", "stderr-json").Current).To(Equal("controller-runtime"))
		})

		It("names a standalone zap logger in a controller module", func() {
			Expect(rowFor(run("kubebuilder"), "deps", "go.uber.org/zap").Fix).To(Equal("migrate area: logging"))
		})

		// envtest in a _test.go is a harness, not the incumbent logger or manager
		// the kubernetes rules defer to; reading it as one drops rows this plain
		// stdlib CLI has earned.
		It("does not read envtest in a spec as a controller module", func() {
			report := run("stdlib-cli")
			Expect(rowFor(report, "deps", "flag").Fix).To(Equal("migrate area: cli"))
			for _, check := range []string{"cli", "config", "logging"} {
				Expect(rowFor(report, "deps", check).Current).NotTo(Equal("controller-runtime"), check)
			}
			Expect(rowFor(report, "layout", "cmd").Status).To(Equal(audit.StatusOK))
		})

		It("leaves the stock kubebuilder scaffold's own wiring alone", func() {
			for _, row := range run("kubebuilder-stock").Rows {
				Expect(row.Check).NotTo(BeElementOf("flag", "go.uber.org/zap", "go.uber.org/zap/zapcore"),
					"the scaffold's controller-runtime logger wiring is not a finding")
			}
		})

		It("skips the test rows when the module has no specs", func() {
			report := run("kubebuilder")
			for _, check := range []string{"testing", "fakes"} {
				Expect(rowFor(report, "deps", check).Status).To(Equal(audit.StatusSkipped), check)
			}
		})

		It("raises one migration row per forbidden import present", func() {
			report := run("legacy-daemon")
			Expect(rowFor(report, "deps", "github.com/pkg/errors").Fix).To(Equal("migrate area: errors"))
			Expect(rowFor(report, "deps", "flag").Fix).To(Equal("migrate area: cli"))
			for _, row := range report.Rows {
				Expect(row.Check).NotTo(Equal("go.uber.org/zap"))
			}
		})

		It("leaves stdlib flag outside a command alone", func() {
			for _, row := range run("stdlib-cli").Rows {
				if row.Check == "flag" {
					Expect(row.Current).To(ContainSubstring("cmd/x/main.go"))
				}
			}
		})
	})

	Describe("the logging rows", func() {
		It("skips both rows for a module that reaches for no logger at all", func() {
			report := run("quiet-cli")
			Expect(rowFor(report, "deps", "logging")).To(SatisfyAll(
				HaveField("Status", audit.StatusSkipped), HaveField("Current", "no logging")))
			Expect(rowFor(report, "logging", "stderr-json").Status).To(Equal(audit.StatusSkipped))
		})

		It("does not ask a library to install a handler", func() {
			report := run("library")
			Expect(rowFor(report, "deps", "logging").Status).To(Equal(audit.StatusOK))
			Expect(rowFor(report, "logging", "stderr-json")).To(SatisfyAll(
				HaveField("Status", audit.StatusSkipped), HaveField("Current", "library")))
		})

		It("still asks a binary that logs to install one", func() {
			Expect(rowFor(run("stdlib-cli"), "logging", "stderr-json").Status).To(Equal(audit.StatusSkipped))
			Expect(rowFor(run("legacy-daemon"), "logging", "stderr-json").Status).To(Equal(audit.StatusGap))
		})

		// A binary on slog's default handler is the case the row exists for, and
		// the one whose file list was empty: the command that has to install the
		// handler is what the migrator opens.
		It("hands the migrator the command when a binary rides the default handler", func() {
			Expect(rowFor(run("slog-default"), "logging", "stderr-json")).To(SatisfyAll(
				HaveField("Status", audit.StatusGap),
				HaveField("Current", "no non-test file builds a slog.NewJSONHandler")))

			paths, err := audit.Files(options("slog-default"), "logging")
			Expect(err).NotTo(HaveOccurred())
			Expect(paths).To(ConsistOf("cmd/svc/main.go"))
		})

		It("keeps a library module's pkg/ directory", func() {
			Expect(rowFor(run("library"), "layout", "pkg-dir")).To(SatisfyAll(
				HaveField("Status", audit.StatusOK), HaveField("Current", "pkg/ in a library module")))
		})

		// A gap converge cannot hand a migrator any file is a dispatch with an
		// empty list; the row and the file list have to agree.
		It("never reports a logging gap with no file for the migrator to open", func() {
			for _, fixture := range fixtures {
				report := run(fixture)
				if rowFor(report, "deps", "logging").Status != audit.StatusGap &&
					rowFor(report, "logging", "stderr-json").Status != audit.StatusGap {
					continue
				}

				paths, err := audit.Files(options(fixture), "logging")
				Expect(err).NotTo(HaveOccurred())
				Expect(paths).NotTo(BeEmpty(), "%s reports a logging gap, --files logging is empty", fixture)
			}
		})
	})

	Describe("the release rows", func() {
		It("finds -X in the Makefile and in a workflow run block", func() {
			row := rowFor(run("legacy-daemon"), "release", "no-ldflags-x")
			Expect(row.Status).To(Equal(audit.StatusGap))
			Expect(row.Current).To(ContainSubstring("Makefile"))
			Expect(row.Current).To(ContainSubstring("release.yml"))
		})

		It("does not read a -X out of a comment beside a run block", func() {
			Expect(rowFor(run("compliant"), "release", "no-ldflags-x").Status).To(Equal(audit.StatusOK))
		})

		DescribeTable("reads -X only where it takes importpath.name=value",
			func(line string, want bool) {
				Expect(audit.LooksLikeLdflagsX(line)).To(Equal(want), "line %q", line)
			},
			Entry("a Makefile stamp", `go build -ldflags "-s -w -X main.version=$(VERSION)" .`, true),
			Entry("a stamp opening the quoted flag list", `go build -ldflags "-X main.version=$TAG" .`, true),
			Entry("the -X= spelling", `go build -ldflags=-X=example.com/p/v.Version=1 .`, true),
			Entry("a make variable in front of the path", `LDFLAGS := -X $(PKG)/internal/version.Version=$(VERSION)`, true),
			Entry("a braced shell variable", `go build -ldflags "-X ${MODULE}/internal/version.Version=$VERSION"`, true),
			Entry("a bare shell variable", `go build -ldflags "-X $PKG.version=$VERSION"`, true),
			Entry("curl's request method", `curl -sS -X POST https://api.example.com/v1/things`, false),
			Entry("curl with a form field elsewhere on the line", `curl -X PUT -d name.first=ada https://x`, false),
			Entry("prose naming the flag", `# the binary carries no -X ldflags: go build stamps it`, false),
			Entry("curl with a query string", `curl -X GET https://x/a.b?c=1`, false),
			Entry("ssh forwarding", `ssh -X user@host.example.com`, false),
			Entry("tar's exclude file", `tar -X exclude.txt -cf out.tar src`, false),
		)
	})
})

// placeholders are the values testdata/compliant was rendered with; the spec
// below re-renders every templated file and diffs it, so a template edit fails
// here instead of leaving the fixture quietly stale.
var unfilled = regexp.MustCompile(`\{\{[A-Z][A-Z_]*\}\}`)

var placeholders = map[string]string{
	"{{MODULE}}":        "example.com/hello",
	"{{MODULES}}":       ".",
	"{{BINARY}}":        "hello",
	"{{ENV_PREFIX}}":    "HELLO",
	"{{OWNER}}":         "example",
	"{{REPO}}":          "hello",
	"{{DESCRIPTION}}":   "greet the world",
	"{{PACKAGE}}":       "cli",
	"{{PACKAGE_TITLE}}": "CLI",
}

func canonTemplate(name string) string {
	GinkgoHelper()

	data, err := os.ReadFile(filepath.Join(pluginRoot, "skills", "go-conventions", "templates", name))
	Expect(err).NotTo(HaveOccurred())

	rendered := string(data)
	for token, value := range placeholders {
		rendered = strings.ReplaceAll(rendered, token, value)
	}
	// Only a {{NAME}} token is filled at copy time; goreleaser's {{ .Version }}
	// and a workflow's ${{ … }} survive rendering verbatim
	// (references/layout.md, "Template placeholders").
	Expect(unfilled.FindString(rendered)).To(BeEmpty(), "template %s has an unfilled placeholder", name)

	return rendered
}

var _ = Describe("the compliant fixture", func() {
	DescribeTable("is the canon template, rendered",
		func(template, fixture string) {
			data, err := os.ReadFile(filepath.Join("testdata", "compliant", filepath.FromSlash(fixture)))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(canonTemplate(template)),
				"testdata/compliant/%s has drifted from templates/%s; re-render it", fixture, template)
		},
		Entry("the lint config", ".golangci.yml", ".golangci.yml"),
		Entry("the Makefile", "Makefile", "Makefile"),
		Entry("the CI workflow", "ci.yml", ".github/workflows/ci.yml"),
		Entry("the release workflow", "release.yml", ".github/workflows/release.yml"),
		Entry("the goreleaser config", ".goreleaser.yaml", ".goreleaser.yaml"),
		Entry("the gitignore", ".gitignore", ".gitignore"),
		Entry("the CLAUDE.md pointer", "CLAUDE-pointer.md", "CLAUDE.md"),
		Entry("main", "main.go", "cmd/hello/main.go"),
		Entry("the command tree", "root.go", "internal/cli/root.go"),
		Entry("the version package", "version.go", "internal/version/version.go"),
		Entry("the suite bootstrap", "suite_test.go", "internal/cli/cli_suite_test.go"),
	)

	It("carries both halves of the dependabot config", func() {
		data, err := os.ReadFile(filepath.Join("testdata", "compliant", ".github", "dependabot.yml"))
		Expect(err).NotTo(HaveOccurred())

		sibling, err := os.ReadFile(filepath.Join(pluginRoot, "..", "github-conventions",
			"skills", "github-conventions", "templates", "dependabot.yml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(string(data)).To(Equal(string(sibling) + canonTemplate("dependabot-gomod.yml")))
	})
})

var _ = Describe("Files", func() {
	DescribeTable("matches the golden file list for an area",
		func(area string) {
			paths, err := audit.Files(options("legacy-daemon"), area)
			Expect(err).NotTo(HaveOccurred())
			golden(filepath.Join("testdata", "legacy-daemon", "files-"+area+".txt"),
				strings.Join(paths, "\n")+"\n")
		},
		Entry("logging", "logging"),
		Entry("testing", "testing"),
	)

	It("rejects an unknown area", func() {
		_, err := audit.Files(options("legacy-daemon"), "nonsense")
		Expect(err).To(HaveOccurred())
	})

	It("returns sorted, unique paths", func() {
		for _, area := range []string{"cli", "logging", "testing", "errors", "layout", "version"} {
			paths, err := audit.Files(options("legacy-daemon"), area)
			Expect(err).NotTo(HaveOccurred())

			want := slices.Clone(paths)
			slices.Sort(want)
			Expect(paths).To(Equal(want), "area %q is out of order", area)
			Expect(paths).To(Equal(slices.Compact(want)), "area %q repeats a path", area)
		}
	})

	It("lists every cmd main and the whole cli package for the cli area", func() {
		paths, err := audit.Files(options("compliant"), "cli")
		Expect(err).NotTo(HaveOccurred())
		Expect(paths).To(ContainElements("cmd/hello/main.go", "internal/cli/root.go"))
	})
})

var _ = Describe("Golangci", func() {
	It("fills the module path from go.mod", func() {
		out, err := audit.Golangci(options("legacy-daemon"))
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("example.com/daemon"))
		Expect(out).NotTo(ContainSubstring("{{MODULE}}"))
	})

	It("drops the deny entries the repository would trip on today", func() {
		out, err := audit.Golangci(options("legacy-daemon"))
		Expect(err).NotTo(HaveOccurred())
		Expect(out).NotTo(ContainSubstring("github.com/pkg/errors"))
		Expect(out).To(ContainSubstring("go.uber.org/zap"))
		golden(filepath.Join("testdata", "legacy-daemon", "emit-golangci.yml"), out)
	})

	It("reproduces the config a repository built from the templates already carries", func() {
		out, err := audit.Golangci(options("compliant"))
		Expect(err).NotTo(HaveOccurred())

		want, err := os.ReadFile(filepath.Join("testdata", "compliant", ".golangci.yml"))
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(Equal(string(want)))
	})

	It("emits a config that parses as YAML", func() {
		out, err := audit.Golangci(options("legacy-daemon"))
		Expect(err).NotTo(HaveOccurred())

		var doc map[string]any
		Expect(yaml.Unmarshal([]byte(out), &doc)).To(Succeed())
		Expect(doc).To(HaveKeyWithValue("version", "2"))
	})

	It("fails when the plugin root is unset", func() {
		_, err := audit.Golangci(audit.Options{Dir: filepath.Join("testdata", "compliant")})
		Expect(err).To(HaveOccurred())
	})

	It("fails when the template is missing", func() {
		_, err := audit.Golangci(audit.Options{
			Dir:        filepath.Join("testdata", "compliant"),
			PluginRoot: filepath.Join("testdata", "compliant"),
		})
		Expect(err).To(HaveOccurred())
	})
})
