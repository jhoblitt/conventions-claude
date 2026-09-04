package audit

import (
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// rows runs every check in the order the report renders them.
func (r *repo) rows() []Row {
	rows := make([]Row, 0, 48)
	rows = append(rows, r.toolchainRow())
	rows = append(rows, r.lintRows()...)
	rows = append(rows, r.makefileRows()...)
	rows = append(rows, r.ciRows()...)
	rows = append(rows, r.releaseRows()...)
	rows = append(rows, r.dependabotRow(), r.gitignoreRow(), r.claudeRow(), r.versionRow())
	rows = append(rows, r.layoutRows()...)
	rows = append(rows, r.depsRows()...)

	return append(rows, r.forbiddenRows()...)
}

// --- the module itself ---

func (r *repo) toolchainRow() Row {
	const (
		canon = "go 1.27, the minor only, and no toolchain line"
		fix   = "set go 1.27 in go.mod and drop any toolchain line"
	)

	current := "no go directive"
	if r.goVersion != "" {
		current = "go " + r.goVersion
	}
	if r.toolchain != "" {
		current += ", toolchain " + r.toolchain
	}
	if r.goVersion == canonGoVersion && r.toolchain == "" {
		return okRow("toolchain", "go-directive", current, canon)
	}

	return gapRow(PhaseTooling, "toolchain", "go-directive", current, canon, fix)
}

const canonGoVersion = "1.27"

// --- lint ---

const (
	canonLintConfig = `a .golangci.yml at the root, golangci-lint v2 (version: "2")`
	fixLint         = "regenerate: goconv-audit --emit-golangci"
)

func (r *repo) lintRows() []Row {
	config := r.lintConfigRow()
	if config.Status != StatusOK {
		return []Row{
			config,
			skipRow("lint", "linters", "no v2 lint config", canonLintEnable),
			skipRow("lint", "formatters", "no v2 lint config", canonLintFormatters),
			skipRow("lint", "max-issues", "no v2 lint config", canonLintIssues),
		}
	}

	return []Row{config, r.lintersRow(), r.formattersRow(), r.maxIssuesRow()}
}

func (r *repo) lintConfigRow() Row {
	switch {
	case !r.golangci.found:
		return gapRow(PhaseTooling, "lint", "config", "no .golangci.yml", canonLintConfig, fixLint)
	case r.golangci.parse != nil:
		return gapRow(PhaseTooling, "lint", "config",
			"yaml parse error: "+r.golangci.parse.Error(), canonLintConfig, fixLint)
	}

	version, hasVersion := scalarAt(r.golangci.doc, "version")

	var v1 []string
	if !hasVersion {
		v1 = append(v1, "no version key")
	}
	if _, ok := scalarAt(r.golangci.doc, "run", "deadline"); ok {
		v1 = append(v1, "run.deadline")
	}
	if _, ok := scalarAt(r.golangci.doc, "linters", "disable-all"); ok {
		v1 = append(v1, "linters.disable-all")
	}
	if len(v1) > 0 {
		return gapRow(PhaseTooling, "lint", "config",
			listing("v1 schema", v1, ""), canonLintConfig, fixLint)
	}
	if version != "2" {
		return gapRow(PhaseTooling, "lint", "config",
			r.golangci.path+", version: "+version, canonLintConfig, fixLint)
	}

	return okRow("lint", "config", r.golangci.path+", version: 2", canonLintConfig)
}

const (
	canonLintEnable     = "every linter templates/.golangci.yml enables"
	canonLintFormatters = "the formatters templates/.golangci.yml enables"
	canonLintIssues     = "issues.max-issues-per-linter and max-same-issues are both 0"
)

func (r *repo) lintersRow() Row {
	return r.enabledRow("linters", canonLintEnable, "linters", "enable")
}

func (r *repo) formattersRow() Row {
	return r.enabledRow("formatters", canonLintFormatters, "formatters", "enable")
}

// enabledRow compares one enable list against the template's.
func (r *repo) enabledRow(check, canon string, keys ...string) Row {
	template, err := r.templateDocument(".golangci.yml")
	if err != nil {
		return templateMissing("lint", check, canon, err.Error())
	}

	want := stringsAt(template, keys...)
	got := stringsAt(r.golangci.doc, keys...)

	var missing []string
	for _, name := range want {
		if !slices.Contains(got, name) {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return gapRow(PhaseTooling, "lint", check, listing("missing", missing, ""), canon, fixLint)
	}

	return okRow("lint", check, listing("enabled", got, "none enabled"), canon)
}

func (r *repo) maxIssuesRow() Row {
	perLinter, hasPerLinter := scalarAt(r.golangci.doc, "issues", "max-issues-per-linter")
	same, hasSame := scalarAt(r.golangci.doc, "issues", "max-same-issues")

	current := "max-issues-per-linter: " + or(perLinter, hasPerLinter) +
		", max-same-issues: " + or(same, hasSame)
	if perLinter == "0" && same == "0" {
		return okRow("lint", "max-issues", current, canonLintIssues)
	}

	return gapRow(PhaseTooling, "lint", "max-issues", current, canonLintIssues, fixLint)
}

func or(value string, present bool) string {
	if !present {
		return "unset"
	}

	return value
}

// --- Makefile ---

const fixMakefile = "copy templates/Makefile"

func (r *repo) makefileRows() []Row {
	const (
		canonPresent = "a Makefile at the repository root"
		canonTargets = "every target templates/Makefile defines"
		canonVersion = "GOLANGCI_LINT_VERSION is defined in the Makefile"
	)

	present := okRow("makefile", "present", "Makefile", canonPresent)
	if !r.hasMakefile {
		present = gapRow(PhaseTooling, "makefile", "present", "no Makefile", canonPresent, fixMakefile)
	}

	version := gapRow(PhaseTooling, "makefile", "golangci-version",
		"GOLANGCI_LINT_VERSION is not defined", canonVersion, fixMakefile)
	if strings.Contains(r.makefile, "GOLANGCI_LINT_VERSION") {
		version = okRow("makefile", "golangci-version", "GOLANGCI_LINT_VERSION defined", canonVersion)
	}

	return []Row{present, r.targetsRow(canonTargets), version}
}

// makeTarget is a target definition: the template is the one home of the list,
// so the names are read from it at runtime rather than copied into this file.
var makeTarget = regexp.MustCompile(`(?m)^([a-z-]+):`)

func (r *repo) targetsRow(canon string) Row {
	template, err := r.template("Makefile")
	if err != nil {
		return templateMissing("makefile", "targets", canon, err.Error())
	}

	want := targets(template)
	got := targets(r.makefile)

	var missing []string
	for _, name := range want {
		if !slices.Contains(got, name) {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		current := listing("missing", missing, "")
		if !r.hasMakefile {
			current = "no Makefile"
		}

		return gapRow(PhaseTooling, "makefile", "targets", current, canon, fixMakefile)
	}

	return okRow("makefile", "targets", listing("defined", got, "none defined"), canon)
}

func targets(text string) []string {
	var names []string
	for _, m := range makeTarget.FindAllStringSubmatch(text, -1) {
		if !slices.Contains(names, m[1]) {
			names = append(names, m[1])
		}
	}

	return names
}

// --- CI ---

const fixCI = "copy templates/ci.yml to .github/workflows/"

func (r *repo) ciRows() []Row {
	hasGoVersionFile := func(d document) bool {
		found := false
		walkMaps(d.doc, func(node map[string]any) {
			if file, ok := node["go-version-file"].(string); ok && file == "go.mod" {
				found = true
			}
		})

		return found
	}

	var setupGo []string
	for _, w := range r.workflows {
		if hasGoVersionFile(w) {
			setupGo = append(setupGo, w.path)
		}
	}

	race := r.runsMatching(func(s string) bool {
		return (strings.Contains(s, "go test") && strings.Contains(s, "-race")) ||
			strings.Contains(s, "make test") || strings.Contains(s, "make check")
	})
	checks := r.runsMatching(func(s string) bool {
		return (strings.Contains(s, "fix-check") && strings.Contains(s, "tidy-check")) ||
			strings.Contains(s, "make check")
	})
	vuln := append(r.usesAction("golang/govulncheck-action"),
		r.runsMatching(func(s string) bool { return strings.Contains(s, "govulncheck") })...)

	return []Row{
		r.workflowRow("ci", "setup-go", setupGo,
			"a workflow sets Go up with go-version-file: go.mod",
			"no workflow sets go-version-file: go.mod", fixCI),
		r.workflowRow("ci", "race", race,
			"a workflow runs the suite with the race detector",
			"no run: block runs go test -race or make test", fixCI),
		r.workflowRow("ci", "lint", r.usesAction("golangci/golangci-lint-action"),
			"a workflow runs golangci/golangci-lint-action",
			"no step uses golangci/golangci-lint-action", fixCI),
		r.workflowRow("ci", "checks", checks,
			"a workflow runs the go fix and go mod tidy checks",
			"no run: block runs fix-check and tidy-check", fixCI),
		r.workflowRow("ci", "govulncheck", vuln,
			"a workflow runs govulncheck over ./...",
			"no step or run: block runs govulncheck", fixCI),
	}
}

// workflowRow reports the workflows that satisfied a check, or the gap plus
// the workflows that could not be read at all.
func (r *repo) workflowRow(area, check string, names []string, canon, missing, fix string) Row {
	if len(names) > 0 {
		slices.Sort(names)

		return okRow(area, check, strings.Join(slices.Compact(names), ", "), canon)
	}
	if unreadable := r.unreadable(); len(unreadable) > 0 {
		missing += "; " + listing("unreadable", unreadable, "")
	}

	return gapRow(PhaseTooling, area, check, missing, canon, fix)
}

// --- release ---

const fixGoreleaser = "copy templates/.goreleaser.yaml"

func (r *repo) releaseRows() []Row {
	return []Row{
		r.goreleaserRow(),
		r.goreleaserKeyRow("kos", []string{"kos"},
			"a kos block builds the image from the same checkout"),
		r.goreleaserKeyRow("sign-sbom", []string{"signs", "sboms"},
			"signs and sboms blocks: cosign over the checksums, syft over the archives"),
		r.releaseWorkflowRow(),
		r.ldflagsRow(),
	}
}

func (r *repo) goreleaserRow() Row {
	const canon = ".goreleaser.yaml at the root, version: 2"

	switch {
	case !r.goreleaser.found:
		return gapRow(PhaseTooling, "release", "goreleaser", "no .goreleaser.yaml", canon, fixGoreleaser)
	case r.goreleaser.parse != nil:
		return gapRow(PhaseTooling, "release", "goreleaser",
			"yaml parse error: "+r.goreleaser.parse.Error(), canon, fixGoreleaser)
	}

	version, _ := scalarAt(r.goreleaser.doc, "version")
	if version != "2" {
		return gapRow(PhaseTooling, "release", "goreleaser",
			r.goreleaser.path+", version: "+or(version, version != ""), canon, fixGoreleaser)
	}

	return okRow("release", "goreleaser", r.goreleaser.path+", version: 2", canon)
}

func (r *repo) goreleaserKeyRow(check string, keys []string, canon string) Row {
	if !r.goreleaser.found || r.goreleaser.parse != nil {
		return gapRow(PhaseTooling, "release", check, "no readable .goreleaser.yaml", canon, fixGoreleaser)
	}

	var missing []string
	for _, key := range keys {
		if !r.goreleaser.hasKey(key) {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		return gapRow(PhaseTooling, "release", check, listing("missing", missing, ""), canon, fixGoreleaser)
	}

	return okRow("release", check, listing("present", keys, ""), canon)
}

func (r *repo) releaseWorkflowRow() Row {
	const (
		canon = "a workflow on v* tags runs goreleaser/goreleaser-action"
		fix   = "copy templates/release.yml to .github/workflows/"
	)

	var names []string
	for _, w := range r.workflows {
		if !tagTriggered(w) {
			continue
		}
		for _, ref := range w.usesRefs() {
			if refUses(ref, "goreleaser/goreleaser-action") {
				names = append(names, w.path)

				break
			}
		}
	}

	return r.workflowRow("release", "workflow", names, canon,
		"no workflow runs goreleaser on a v* tag", fix)
}

// tagTriggered reports whether the workflow's push trigger carries a v* tag
// pattern.
func tagTriggered(w document) bool {
	for _, pattern := range stringsAt(w.doc, "on", "push", "tags") {
		if strings.HasPrefix(pattern, "v") && strings.Contains(pattern, "*") {
			return true
		}
	}

	return false
}

// ldflagsX is the linker flag as it appears on a build command line. -X always
// takes importpath.name=value, and requiring that shape is what separates it
// from ssh -X, tar -X, curl -X POST, and prose naming the flag.
//
// The import path is matched as anything up to the separators, not as a Go path:
// a Makefile writes the stamp through a variable far more often than literally —
// -X $(PKG)/internal/version.Version=$(VERSION) — and a path class that excluded
// $ ( ) { } read every one of those as clean.
//
// The goreleaser ldflags values, the Makefile, and workflow run: blocks are
// searched; a workflow's YAML comments are not, so the templates' own note that
// they carry no -X is never read back as one.
var ldflagsX = regexp.MustCompile(`(?:^|[\s"'=])-X[ \t=]+["']?[^\s"'=]*\.[A-Za-z_][A-Za-z0-9_]*=`)

func (r *repo) ldflagsRow() Row {
	const (
		canon = "no -X ldflags anywhere; the version is the build stamp"
		fix   = "drop the -X ldflags; internal/version reads debug.ReadBuildInfo"
	)

	var found []string
	if r.goreleaser.found && r.goreleaser.parse == nil {
		walkMaps(r.goreleaser.doc, func(node map[string]any) {
			for _, flag := range stringsAt(node, "ldflags") {
				if ldflagsX.MatchString(flag) && !slices.Contains(found, r.goreleaser.path) {
					found = append(found, r.goreleaser.path)
				}
			}
		})
	}
	if ldflagsX.MatchString(r.makefile) {
		found = append(found, "Makefile")
	}
	found = append(found, r.runsMatching(ldflagsX.MatchString)...)

	if len(found) > 0 {
		slices.Sort(found)

		return gapRow(PhaseTooling, "release", "no-ldflags-x",
			listing("-X in", slices.Compact(found), ""), canon, fix)
	}

	return okRow("release", "no-ldflags-x", "no -X ldflags", canon)
}

// --- repository files ---

func (r *repo) dependabotRow() Row {
	const (
		canon = ".github/dependabot.yml carries a gomod updates entry"
		fix   = "append templates/dependabot-gomod.yml under the existing updates: list"
	)

	if !r.dependabot.found {
		return gapRow(PhaseTooling, "dependabot", "gomod", "no .github/dependabot.yml", canon, fix)
	}
	if r.dependabot.parse != nil {
		return gapRow(PhaseTooling, "dependabot", "gomod",
			"yaml parse error: "+r.dependabot.parse.Error(), canon, fix)
	}

	var ecosystems []string
	if updates, isList := r.dependabot.doc["updates"].([]any); isList {
		for _, update := range updates {
			if name, ok := scalarAt(update, "package-ecosystem"); ok {
				ecosystems = append(ecosystems, name)
			}
		}
	}
	if !slices.Contains(ecosystems, "gomod") {
		return gapRow(PhaseTooling, "dependabot", "gomod",
			listing("ecosystems", ecosystems, "no updates[] entries"), canon, fix)
	}

	return okRow("dependabot", "gomod", listing("ecosystems", ecosystems, ""), canon)
}

func (r *repo) gitignoreRow() Row {
	const (
		canon = ".gitignore at the repository root"
		fix   = "merge templates/.gitignore into the repository's"
	)

	if !r.gitignore {
		return gapRow(PhaseTooling, "gitignore", "present", "no .gitignore", canon, fix)
	}

	return okRow("gitignore", "present", ".gitignore", canon)
}

// claudePointer opens the block templates/CLAUDE-pointer.md installs.
const claudePointer = "<!-- go-conventions:begin -->"

func (r *repo) claudeRow() Row {
	const (
		canon = "CLAUDE.md carries the go-conventions pointer block"
		fix   = "insert templates/CLAUDE-pointer.md between its markers"
	)

	if !strings.Contains(r.claudeMD, claudePointer) {
		current := "no pointer block"
		if r.claudeMD == "" {
			current = "no CLAUDE.md"
		}

		return gapRow(PhaseTooling, "claude", "pointer", current, canon, fix)
	}

	return okRow("claude", "pointer", "pointer block present", canon)
}

func (r *repo) versionRow() Row {
	const (
		canon = "internal/version reads debug.ReadBuildInfo"
		fix   = "copy templates/version.go to internal/version/"
	)

	if !r.hasMain() {
		return skipRow("version", "package", "library", canon)
	}
	if !r.hasVersion {
		return gapRow(PhaseTooling, "version", "package", "no internal/version", canon, fix)
	}
	for _, f := range r.sources {
		if f.dir == "internal/version" && strings.Contains(f.text, "debug.ReadBuildInfo") {
			return okRow("version", "package", "internal/version", canon)
		}
	}

	return gapRow(PhaseTooling, "version", "package",
		"internal/version does not read debug.ReadBuildInfo", canon, fix)
}

// --- layout ---

func (r *repo) layoutRows() []Row {
	const (
		canonCmd = "every package main lives in cmd/<name>"
		canonPkg = "no pkg/ in a module that ships a binary"
		fix      = "migrate area: layout"
	)

	cmd := okRow("layout", "cmd", "every main package is under cmd/<name>", canonCmd)
	switch mains := r.mainDirs(); {
	case len(mains) == 0:
		cmd = skipRow("layout", "cmd", "library", canonCmd)
	// The kubebuilder scaffold's cmd/main.go is kept as generated, and
	// references/layout.md's cmd/<bin> rule yields to it.
	case r.controllerRuntime():
		cmd = skipRow("layout", "cmd", "controller-runtime", canonCmd)
	default:
		var offenders []string
		for _, dir := range mains {
			if !strings.HasPrefix(dir, "cmd/") || strings.Count(dir, "/") != 1 {
				offenders = append(offenders, dir)
			}
		}
		if len(offenders) > 0 {
			cmd = gapRow(PhaseMigration, "layout", "cmd",
				listing("main packages outside cmd/<name>", offenders, ""), canonCmd, fix)
		}
	}

	hasPkg := slices.ContainsFunc(r.sources, func(f goFile) bool {
		return f.dir == "pkg" || strings.HasPrefix(f.dir, "pkg/")
	})
	switch {
	case hasPkg && r.hasMain():
		return []Row{cmd, gapRow(PhaseMigration, "layout", "pkg-dir",
			"pkg/ beside a main package", canonPkg, fix)}
	case hasPkg:
		return []Row{cmd, okRow("layout", "pkg-dir", "pkg/ in a library module", canonPkg)}
	}

	return []Row{cmd, okRow("layout", "pkg-dir", "no pkg/", canonPkg)}
}

// --- dependencies ---

func (r *repo) depsRows() []Row {
	return []Row{
		r.frameworkRow("cli", "github.com/spf13/cobra",
			"cobra builds every command tree", "migrate area: cli"),
		r.frameworkRow("config", "github.com/spf13/viper",
			"viper resolves flags, environment, and config file", "migrate area: cli"),
		r.testingRow(),
		r.fakesRow(),
		r.loggingRow(),
		r.stderrJSONRow(),
	}
}

// frameworkRow reports a dependency a binary must carry. A module with no main
// package has no command tree to build, and a controller-runtime module keeps
// the flag wiring kubebuilder scaffolded (references/kubernetes.md).
func (r *repo) frameworkRow(check, pkg, canon, fix string) Row {
	switch {
	case !r.hasMain():
		return skipRow("deps", check, "library", canon)
	case r.controllerRuntime():
		return skipRow("deps", check, "controller-runtime", canon)
	case !r.imports(pkg):
		return gapRow(PhaseMigration, "deps", check, "not imported", canon, fix)
	}

	return okRow("deps", check, pkg, canon)
}

func (r *repo) testingRow() Row {
	const (
		canon = "Ginkgo v2 and Gomega in every _test.go"
		fix   = "migrate area: testing"
	)

	if !r.hasTests() {
		return skipRow("deps", "testing", "no test files", canon)
	}

	var missing []string
	for _, pkg := range []string{"github.com/onsi/ginkgo/v2", "github.com/onsi/gomega"} {
		if !slices.ContainsFunc(r.sources, func(f goFile) bool {
			return f.test && slices.Contains(f.imports, pkg)
		}) {
			missing = append(missing, pkg)
		}
	}
	if len(missing) > 0 {
		return gapRow(PhaseMigration, "deps", "testing", listing("missing", missing, ""), canon, fix)
	}

	return okRow("deps", "testing", "ginkgo/v2, gomega", canon)
}

func (r *repo) fakesRow() Row {
	const (
		canon = "a tool directive for counterfeiter generates the fakes"
		fix   = "migrate area: testing"
		pkg   = "github.com/maxbrunsfeld/counterfeiter/v6"
	)

	if !r.hasTests() {
		return skipRow("deps", "fakes", "no test files", canon)
	}
	if !slices.Contains(r.tools, pkg) {
		return gapRow(PhaseMigration, "deps", "fakes", "no tool directive", canon, fix)
	}

	return okRow("deps", "fakes", "tool "+pkg, canon)
}

// loggers are the log packages the canon replaces with log/slog.
var loggers = []string{"log", "github.com/sirupsen/logrus", "go.uber.org/zap", "github.com/rs/zerolog"}

func (r *repo) loggingRow() Row {
	const (
		canon = "log/slog only, in every non-test file that logs"
		fix   = "migrate area: logging"
	)

	if r.controllerRuntime() {
		return skipRow("deps", "logging", "controller-runtime", canon)
	}

	var present []string
	for _, pkg := range loggers {
		if slices.ContainsFunc(r.sources, func(f goFile) bool {
			return !f.test && importMatches(f.imports, pkg, pkg == "log")
		}) {
			present = append(present, pkg)
		}
	}
	if len(present) > 0 {
		return gapRow(PhaseMigration, "deps", "logging", listing("imports", present, ""), canon, fix)
	}
	// references/logging.md governs which logger a module uses when it logs, not
	// whether it logs: a module that reaches for none has nothing to migrate.
	if !r.importsNonTest("log/slog") {
		return skipRow("deps", "logging", "no logging", canon)
	}

	return okRow("deps", "logging", "log/slog", canon)
}

// jsonHandler is the handler the canon installs by default; stdoutHandler is
// any slog handler built over the stream reserved for program output.
//
// The canon's own templates/root.go builds the handler over the writer main
// passed in, so a literal slog.NewJSONHandler(os.Stderr is not the test: what
// the row reports is a handler pointed at stdout, and the absence of a JSON
// handler altogether.
var (
	jsonHandler   = regexp.MustCompile(`slog\.NewJSONHandler\(`)
	stdoutHandler = regexp.MustCompile(`slog\.New\w*Handler\(os\.Stdout`)
)

func (r *repo) stderrJSONRow() Row {
	const (
		canon = "the JSON handler writes to stderr; stdout is program output"
		fix   = "migrate area: logging"
	)

	if r.controllerRuntime() {
		return skipRow("logging", "stderr-json", "controller-runtime", canon)
	}
	if !r.importsNonTest("log/slog") {
		return skipRow("logging", "stderr-json", "log/slog is not imported", canon)
	}

	var stdout []string
	for _, f := range r.sources {
		if !f.test && stdoutHandler.MatchString(f.text) {
			stdout = append(stdout, f.path)
		}
	}
	if len(stdout) > 0 {
		slices.Sort(stdout)

		return gapRow(PhaseMigration, "logging", "stderr-json",
			listing("handler over os.Stdout in", stdout, ""), canon, fix)
	}
	if slices.ContainsFunc(r.sources, func(f goFile) bool {
		return !f.test && jsonHandler.MatchString(f.text)
	}) {
		return okRow("logging", "stderr-json", "slog.NewJSONHandler, never over os.Stdout", canon)
	}
	// Installing the handler is the binary's job; a library logs through
	// whatever default its caller installed, so it has nothing to migrate here.
	if !r.hasMain() {
		return skipRow("logging", "stderr-json", "library", canon)
	}

	return gapRow(PhaseMigration, "logging", "stderr-json",
		"no non-test file builds a slog.NewJSONHandler", canon, fix)
}

// --- forbidden imports ---

// forbiddenImport is one denied package and the migration area that removes
// it. Matching is by prefix, so a subpackage counts, unless exact is set.
type forbiddenImport struct {
	pkg   string
	exact bool
	// commandOnly restricts the match to a main package or internal/cli, where
	// the stdlib flag package is the CLI the canon replaces.
	commandOnly bool
	// scaffolded lists the matching import paths a controller-runtime module
	// carries only to wire controller-runtime's own logger and its flags — the
	// wiring references/kubernetes.md keeps as generated, beside rows already
	// skipped for it. Under controller-runtime those paths raise no row; a path
	// outside the set, such as a standalone go.uber.org/zap logger, still does.
	scaffolded []string
	area       string
	canon      string
}

var forbiddenImports = []forbiddenImport{
	{
		pkg: "flag", exact: true, commandOnly: true, scaffolded: []string{"flag"},
		area: "cli", canon: "cobra, never the stdlib flag package",
	},
	{pkg: "github.com/alecthomas/kong", area: "cli", canon: "cobra builds every command tree"},
	{pkg: "github.com/urfave/cli", area: "cli", canon: "cobra builds every command tree"},
	{pkg: "github.com/pkg/errors", area: "errors", canon: "errors and fmt.Errorf with %w"},
	{pkg: "github.com/stretchr/testify", area: "testing", canon: "Ginkgo v2 and Gomega"},
	{pkg: "github.com/sirupsen/logrus", area: "logging", canon: "log/slog"},
	{
		pkg: "go.uber.org/zap", scaffolded: []string{"go.uber.org/zap/zapcore"},
		area: "logging", canon: "log/slog",
	},
	{pkg: "github.com/rs/zerolog", area: "logging", canon: "log/slog"},
	{pkg: "io/ioutil", exact: true, area: "errors", canon: "io and os"},
	{pkg: "golang.org/x/exp/slices", area: "errors", canon: "slices from the standard library"},
	{pkg: "golang.org/x/exp/maps", area: "errors", canon: "maps from the standard library"},
	{pkg: "github.com/golang/mock", area: "testing", canon: "counterfeiter fakes"},
	{pkg: "go.uber.org/mock", area: "testing", canon: "counterfeiter fakes"},
}

func (r *repo) forbiddenRows() []Row {
	controller := r.controllerRuntime()

	var rows []Row
	for _, denied := range forbiddenImports {
		importers := map[string][]string{}
		for _, f := range r.sources {
			if denied.commandOnly && !r.isCommandFile(f) {
				continue
			}
			for _, path := range f.imports {
				if importMatches([]string{path}, denied.pkg, denied.exact) {
					importers[path] = append(importers[path], f.path)
				}
			}
		}

		for _, path := range slices.Sorted(maps.Keys(importers)) {
			if controller && slices.Contains(denied.scaffolded, path) {
				continue
			}
			files := importers[path]
			slices.Sort(files)
			rows = append(rows, gapRow(PhaseMigration, "deps", path,
				listing("imported by", slices.Compact(files), ""),
				denied.canon, "migrate area: "+denied.area))
		}
	}

	return rows
}

// --- shared facts about the tree ---

func (r *repo) mainDirs() []string {
	var dirs []string
	for _, f := range r.sources {
		if !f.test && f.pkg == "main" && !slices.Contains(dirs, f.dir) {
			dirs = append(dirs, f.dir)
		}
	}
	slices.Sort(dirs)

	return dirs
}

func (r *repo) hasMain() bool { return len(r.mainDirs()) > 0 }

func (r *repo) hasTests() bool {
	return slices.ContainsFunc(r.sources, func(f goFile) bool { return f.test })
}

func (r *repo) imports(pkg string) bool {
	return slices.ContainsFunc(r.sources, func(f goFile) bool {
		return importMatches(f.imports, pkg, false)
	})
}

func (r *repo) importsNonTest(pkg string) bool {
	return slices.ContainsFunc(r.sources, func(f goFile) bool {
		return !f.test && importMatches(f.imports, pkg, false)
	})
}

// controllerRuntime reports a module whose own code is built on
// controller-runtime, and so carries the incumbents references/kubernetes.md
// keeps. A _test.go import alone is envtest, which changes nothing the checks
// measure: reading it as the incumbent would silently drop rows a plain stdlib
// CLI has earned.
func (r *repo) controllerRuntime() bool {
	return r.importsNonTest("sigs.k8s.io/controller-runtime")
}

// isCommandFile reports whether the file belongs to the command surface: a
// main package, or the internal/cli tree the canon puts the command tree in.
func (r *repo) isCommandFile(f goFile) bool {
	return f.pkg == "main" || f.dir == "internal/cli" || strings.HasPrefix(f.dir, "internal/cli/")
}

// importMatches reports whether any of paths is pkg or, unless exact, a
// package under it.
func importMatches(paths []string, pkg string, exact bool) bool {
	for _, path := range paths {
		if path == pkg || (!exact && strings.HasPrefix(path, pkg+"/")) {
			return true
		}
	}

	return false
}

// importPaths is every distinct import in the tree, tests included.
func (r *repo) importPaths() []string {
	var paths []string
	for _, f := range r.sources {
		paths = append(paths, f.imports...)
	}
	slices.Sort(paths)

	return slices.Compact(paths)
}

func (r *repo) template(name string) (string, error) {
	if r.opts.PluginRoot == "" {
		return "", errTemplateRoot
	}

	path := filepath.Join(r.opts.PluginRoot, "skills", "go-conventions", "templates", name)
	text, found, err := readFile(path)
	if err != nil {
		return "", err
	}
	if !found {
		return "", &missingTemplateError{path: path}
	}

	return text, nil
}

// templateDocument parses a template for the values a check compares against.
// {{MODULE}} is filled with a stand-in first so the parsed tree carries a module
// path wherever the repository's own config carries one, and the two compare
// like for like.
func (r *repo) templateDocument(name string) (map[string]any, error) {
	text, err := r.template(name)
	if err != nil {
		return nil, err
	}

	doc := newDocument(name, strings.ReplaceAll(text, "{{MODULE}}", "example.com/module"))
	if doc.parse != nil {
		return nil, doc.parse
	}

	return doc.doc, nil
}
