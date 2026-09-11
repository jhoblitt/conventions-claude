package audit_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.yaml.in/yaml/v3"

	"github.com/jhoblitt/conventions-claude/plugins/go-conventions/tools/internal/audit"
)

// tempRepo lays out a repository the release renderings read: a go.mod on
// module, one cmd/<name>/main.go per name, and a .goreleaser.yaml when
// goreleaser is not empty.
func tempRepo(module string, cmds []string, goreleaser string) audit.Options {
	GinkgoHelper()

	dir := GinkgoT().TempDir()
	Expect(os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module "+module+"\n\ngo 1.27\n"), 0o600)).To(Succeed())
	for _, name := range cmds {
		Expect(os.MkdirAll(filepath.Join(dir, "cmd", name), 0o750)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "cmd", name, "main.go"),
			[]byte("package main\n\nfunc main() {}\n"), 0o600)).To(Succeed())
	}
	if goreleaser != "" {
		Expect(os.WriteFile(filepath.Join(dir, ".goreleaser.yaml"), []byte(goreleaser), 0o600)).To(Succeed())
	}

	return audit.Options{Dir: dir, PluginRoot: pluginRoot}
}

const (
	templateTargets = "    goos:\n      - linux\n    goarch:\n      - amd64\n      - arm64\n"
	customTargets   = "    goos:\n      - linux\n      - darwin\n    goarch:\n      - arm64\n"

	// kosCustom is an existing config that opted in and widened its targets.
	kosCustom = "version: 2\nbuilds:\n  - id: x\n    goos: [linux, darwin]\n    goarch: [arm64]\nkos: [{id: x}]\n"
)

// imageLines are what an {{IMAGE}} block contributes to each rendering.
var imageLines = map[string][]string{
	".goreleaser.yaml": {"kos:", "docker_signs:", "ghcr.io/a/b"},
	"release.yml":      {"packages: write", "Log in to ghcr.io"},
}

// expectRendered checks what every rendering must satisfy whatever the image
// decision: no marker or placeholder left, no double blank line, and a parse
// as YAML.
func expectRendered(out string) map[string]any {
	GinkgoHelper()

	Expect(out).NotTo(ContainSubstring("IMAGE"))
	Expect(unfilled.FindString(out)).To(BeEmpty())
	Expect(out).NotTo(ContainSubstring("\n\n\n"))

	var doc map[string]any
	Expect(yaml.Unmarshal([]byte(out), &doc)).To(Succeed())

	return doc
}

var _ = Describe("Goreleaser", func() {
	DescribeTable("keeps or drops the image block and carries the targets",
		func(goreleaser string, in audit.Inputs, image bool, targets string) {
			out, err := audit.Goreleaser(tempRepo("github.com/a/b", []string{"x"}, goreleaser), in)
			Expect(err).NotTo(HaveOccurred())

			doc := expectRendered(out)
			Expect(out).To(ContainSubstring("  - id: x\n    main: ./cmd/x\n    binary: x\n"))
			Expect(out).To(ContainSubstring(targets))
			Expect(doc).To(HaveKeyWithValue("version", 2))
			for _, line := range imageLines[".goreleaser.yaml"] {
				if image {
					Expect(out).To(ContainSubstring(line))
				} else {
					Expect(out).NotTo(ContainSubstring(line))
				}
			}
			if image {
				// goreleaser's own template survives; only {{NAME}} is a placeholder.
				Expect(out).To(ContainSubstring(`"{{ .Version }}"`))
			}
		},
		Entry("a new project, no image", "", audit.Inputs{}, false, templateTargets),
		Entry("a new project asking for an image", "", audit.Inputs{Image: true}, true, templateTargets),
		Entry("an existing kos block with custom targets", kosCustom, audit.Inputs{}, true, customTargets),
		Entry("an existing docker_signs block alone", "version: 2\ndocker_signs: [{cmd: cosign}]\n",
			audit.Inputs{}, true, templateTargets),
		Entry("an existing config with neither block", "version: 2\nbuilds:\n  - id: x\n",
			audit.Inputs{}, false, templateTargets),
		Entry("an existing config that does not parse", "version: [\n", audit.Inputs{}, false, templateTargets),
	)

	It("collapses the blank lines a dropped block leaves to one", func() {
		out, err := audit.Goreleaser(tempRepo("github.com/a/b", []string{"x"}, ""), audit.Inputs{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("checksum:\n  name_template: checksums.txt\n\nsboms:\n"))
	})

	It("carries one list and leaves the other to the template", func() {
		out, err := audit.Goreleaser(tempRepo("github.com/a/b", []string{"x"},
			"version: 2\nbuilds:\n  - id: x\n    goos: [freebsd]\n"), audit.Inputs{})
		Expect(err).NotTo(HaveOccurred())
		Expect(out).To(ContainSubstring("    goos:\n      - freebsd\n    goarch:\n      - amd64\n      - arm64\n"))
	})

	It("fails when the plugin root is unset", func() {
		opts := tempRepo("github.com/a/b", []string{"x"}, "")
		opts.PluginRoot = ""

		_, err := audit.Goreleaser(opts, audit.Inputs{})
		Expect(err).To(MatchError(ContainSubstring("CLAUDE_PLUGIN_ROOT")))
	})
})

var _ = Describe("ReleaseWorkflow", func() {
	DescribeTable("keeps or drops the packages scope and the registry login",
		func(goreleaser string, in audit.Inputs, image bool) {
			out, err := audit.ReleaseWorkflow(tempRepo("github.com/a/b", []string{"x"}, goreleaser), in)
			Expect(err).NotTo(HaveOccurred())

			doc := expectRendered(out)
			Expect(out).To(ContainSubstring("if: github.repository == 'a/b'"))
			Expect(out).To(ContainSubstring(`go build -o "$RUNNER_TEMP/x" ./cmd/x`))
			Expect(out).To(ContainSubstring("TAG: ${{ github.ref_name }}"))
			Expect(doc).To(HaveKey("jobs"))
			for _, line := range imageLines["release.yml"] {
				if image {
					Expect(out).To(ContainSubstring(line))
				} else {
					Expect(out).NotTo(ContainSubstring(line))
				}
			}
			if !image {
				Expect(out).To(ContainSubstring("      contents: write\n      id-token: write\n"))
				Expect(out).To(ContainSubstring("download-syft@v0\n\n      - name: goreleaser\n"))
			}
		},
		Entry("a new project, no image", "", audit.Inputs{}, false),
		Entry("a new project asking for an image", "", audit.Inputs{Image: true}, true),
		Entry("an existing kos block", kosCustom, audit.Inputs{}, true),
	)
})

var _ = Describe("the release inputs", func() {
	DescribeTable("derive the binary from cmd/ and the slug from the module path",
		func(module string, cmds []string, in audit.Inputs, want, wantErr string) {
			out, err := audit.Goreleaser(tempRepo(module, cmds, ""), in)
			if wantErr != "" {
				Expect(err).To(MatchError(ContainSubstring(wantErr)))

				return
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(out).To(ContainSubstring(want))
		},
		Entry("one cmd directory", "github.com/a/b", []string{"x"}, audit.Inputs{Image: true},
			"binary: x\n", ""),
		Entry("two cmd directories", "github.com/a/b", []string{"x", "y"}, audit.Inputs{},
			"", "--binary is required: cmd/ holds x, y"),
		Entry("no cmd directory and no --binary", "github.com/a/b", nil, audit.Inputs{},
			"", "--binary is required: no cmd/<name> directory"),
		Entry("no cmd directory with --binary", "github.com/a/b", nil, audit.Inputs{Binary: "z"},
			"binary: z\n", ""),
		Entry("a deeper github.com module path", "github.com/a/b/c", []string{"x"}, audit.Inputs{Image: true},
			"ghcr.io/a/b\n", ""),
		Entry("a module path off github.com without flags", "example.com/x", []string{"x"}, audit.Inputs{},
			"", "--owner and --repo are required"),
		Entry("a module path off github.com with only --owner", "example.com/x", []string{"x"},
			audit.Inputs{Owner: "o"}, "", "--repo is required"),
		Entry("a module path off github.com with both flags", "example.com/x", []string{"x"},
			audit.Inputs{Owner: "o", Repo: "r", Image: true}, "ghcr.io/o/r\n", ""),
		Entry("flags over a github.com module path", "github.com/a/b", []string{"x"},
			audit.Inputs{Owner: "o", Repo: "r", Image: true}, "ghcr.io/o/r\n", ""),
		Entry("a binary name outside the name class", "github.com/a/b", nil,
			audit.Inputs{Binary: "x y"}, "", `--binary "x y": not a name`),
		// A quoted module directive admits what an unquoted one cannot, and the
		// tool reads go.mod statically, never through the go tool.
		Entry("an owner that would close the fork guard's quote", `"github.com/a'b/c"`, []string{"x"},
			audit.Inputs{}, "", `--owner "a'b": not a name`),
		Entry("a cmd directory name with a newline", "github.com/a/b", []string{"x\ny"},
			audit.Inputs{}, "", `--binary "x\ny": not a name`),
	)

	It("refuses an empty existing config, the tell of a redirect onto it", func() {
		_, err := audit.Goreleaser(tempRepo("github.com/a/b", []string{"x"}, "\n"), audit.Inputs{})
		Expect(err).To(MatchError(ContainSubstring(".goreleaser.yaml is empty")))
	})

	It("rejects a carried target item outside the name class", func() {
		existing := "version: 2\nbuilds:\n  - id: x\n    goos: [\"linux\\n  evil: true\"]\n"
		_, err := audit.Goreleaser(tempRepo("github.com/a/b", []string{"x"}, existing), audit.Inputs{})
		Expect(err).To(MatchError(ContainSubstring("goos item")))
	})

	It("fails when the directory is not a Go repository root", func() {
		_, err := audit.Goreleaser(audit.Options{Dir: GinkgoT().TempDir(), PluginRoot: pluginRoot}, audit.Inputs{})
		Expect(err).To(MatchError(audit.ErrNoModule))
	})
})

// compliantInputs render testdata/compliant's release files: its module is
// not on github.com, so the slug is given; its binary is the one cmd/
// directory, and its image opt-in is the .goreleaser.yaml it already carries.
var compliantInputs = audit.Inputs{Owner: "example", Repo: "hello"}

var _ = Describe("the compliant fixture's release files", func() {
	DescribeTable("are the canon template, rendered by the tool",
		func(render func(audit.Options, audit.Inputs) (string, error), fixture string) {
			out, err := render(options("compliant"), compliantInputs)
			Expect(err).NotTo(HaveOccurred())
			Expect(unfilled.FindString(out)).To(BeEmpty())

			data, err := os.ReadFile(filepath.Join("testdata", "compliant", filepath.FromSlash(fixture)))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(out),
				"testdata/compliant/%s has drifted from its template; re-render it", fixture)
		},
		Entry("the goreleaser config", audit.Goreleaser, ".goreleaser.yaml"),
		Entry("the release workflow", audit.ReleaseWorkflow, ".github/workflows/release.yml"),
	)
})
