package audit

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
)

const scorecardBadgeURL = "api.scorecard.dev/projects/github.com/"

func licenseRows(dir string) ([]Row, error) {
	const (
		canonPresent = "LICENSE exists at the repository root"
		canonKind    = "Apache-2.0, or the license the repository already carries"
		fix          = "copy templates/LICENSE"
	)

	text, found, err := readFile(filepath.Join(dir, "LICENSE"))
	if err != nil {
		return nil, err
	}
	if !found {
		return []Row{
			gapRow("license", "present", "no LICENSE file", canonPresent, fix),
			gapRow("license", "apache-2.0", "no LICENSE file", canonKind, fix),
		}, nil
	}

	return []Row{
		okRow("license", "present", "LICENSE", canonPresent),
		okRow("license", "apache-2.0", licenseName(text), canonKind),
	}, nil
}

func licenseName(text string) string {
	switch {
	case strings.Contains(text, "Apache License") && strings.Contains(text, "Version 2.0"):
		return "Apache-2.0"
	case strings.Contains(text, "MIT License"):
		return "MIT"
	case strings.Contains(text, "GNU GENERAL PUBLIC LICENSE"):
		return "GPL"
	case strings.Contains(text, "Mozilla Public License"):
		return "MPL-2.0"
	case strings.Contains(text, "ISC License"):
		return "ISC"
	case strings.Contains(text, "BSD"):
		return "BSD"
	default:
		return "an unrecognized license"
	}
}

func readmeRows(dir string) ([]Row, error) {
	const (
		canonPresent = "README.md exists at the repository root"
		canonBadge   = "README.md carries the OpenSSF Scorecard badge"
		fixPresent   = "copy templates/README.md"
		fixBadge     = "add the badge line from templates/README.md"
	)

	text, found, err := readFile(filepath.Join(dir, "README.md"))
	if err != nil {
		return nil, err
	}
	if !found {
		return []Row{
			gapRow("readme", "present", "no README.md", canonPresent, fixPresent),
			gapRow("readme", "scorecard-badge", "no README.md", canonBadge, fixPresent),
		}, nil
	}

	badge := gapRow("readme", "scorecard-badge", "no scorecard badge", canonBadge, fixBadge)
	if strings.Contains(text, scorecardBadgeURL) {
		badge = okRow("readme", "scorecard-badge", "scorecard badge present", canonBadge)
	}

	return []Row{okRow("readme", "present", "README.md", canonPresent), badge}, nil
}

const (
	canonDependabotPresent = ".github/dependabot.yml exists"
	canonDependabotActions = "an updates[] entry sets package-ecosystem: github-actions"
	fixDependabot          = "copy templates/dependabot.yml to .github/"
)

func dependabotRows(dir string) ([]Row, error) {
	text, found, err := readFile(filepath.Join(dir, ".github", "dependabot.yml"))
	if err != nil {
		return nil, err
	}
	if !found {
		return []Row{
			gapRow("dependabot", "present", "no .github/dependabot.yml", canonDependabotPresent, fixDependabot),
			gapRow("dependabot", "github-actions", "no .github/dependabot.yml", canonDependabotActions, fixDependabot),
		}, nil
	}

	return []Row{
		okRow("dependabot", "present", ".github/dependabot.yml", canonDependabotPresent),
		dependabotActionsRow(text),
	}, nil
}

func dependabotActionsRow(text string) Row {
	var doc struct {
		Updates []struct {
			Ecosystem string `yaml:"package-ecosystem"`
		} `yaml:"updates"`
	}
	if err := yaml.Unmarshal([]byte(text), &doc); err != nil {
		return gapRow("dependabot", "github-actions", "yaml parse error: "+err.Error(),
			canonDependabotActions, fixDependabot)
	}

	ecosystems := make([]string, 0, len(doc.Updates))
	for _, update := range doc.Updates {
		ecosystems = append(ecosystems, update.Ecosystem)
	}
	if !slices.Contains(ecosystems, "github-actions") {
		return gapRow("dependabot", "github-actions",
			listing("ecosystems", ecosystems, "no updates[] entries"), canonDependabotActions, fixDependabot)
	}

	return okRow("dependabot", "github-actions", listing("ecosystems", ecosystems, ""), canonDependabotActions)
}

func commitlintRows(dir string, workflows []workflow) ([]Row, error) {
	const (
		canonConfig  = ".commitlintrc.yml exists at the repository root"
		canonAction  = "a workflow runs wagoid/commitlint-github-action on pull requests"
		canonFooter  = ".github/scripts/check-breaking-footer.sh exists"
		fixConfig    = "copy templates/.commitlintrc.yml"
		fixAction    = "copy templates/commitlint.yml to .github/workflows/"
		fixFooter    = "copy templates/check-breaking-footer.sh to .github/scripts/"
		footerScript = ".github/scripts/check-breaking-footer.sh"
	)

	rows := make([]Row, 0, 3)

	config, err := exists(filepath.Join(dir, ".commitlintrc.yml"))
	if err != nil {
		return nil, err
	}
	if config {
		rows = append(rows, okRow("commitlint", "config", ".commitlintrc.yml", canonConfig))
	} else {
		rows = append(rows, gapRow("commitlint", "config", "no .commitlintrc.yml", canonConfig, fixConfig))
	}

	rows = append(rows, actionRow("commitlint", "workflow", "wagoid/commitlint-github-action", canonAction, fixAction, workflows))

	footer, err := exists(filepath.Join(dir, footerScript))
	if err != nil {
		return nil, err
	}
	if footer {
		rows = append(rows, okRow("commitlint", "breaking-footer", footerScript, canonFooter))
	} else {
		rows = append(rows, gapRow("commitlint", "breaking-footer", "no "+footerScript, canonFooter, fixFooter))
	}

	return rows, nil
}

func securityRows(workflows []workflow) []Row {
	return []Row{
		actionRow("security", "codeql", "github/codeql-action/init",
			"a workflow runs github/codeql-action/init",
			"copy templates/codeql.yml to .github/workflows/", workflows),
		actionRow("security", "dependency-review", "actions/dependency-review-action",
			"a workflow runs actions/dependency-review-action",
			"copy templates/dependency-review.yml to .github/workflows/", workflows),
		actionRow("security", "scorecard", "ossf/scorecard-action",
			"a workflow runs ossf/scorecard-action",
			"copy templates/scorecard.yml to .github/workflows/", workflows),
	}
}

func workflowLintRows(workflows []workflow) []Row {
	const fix = "copy templates/workflow-lint.yml to .github/workflows/"

	return []Row{
		scriptRow("workflow-lint", "actionlint", "actionlint",
			"a workflow runs actionlint over .github/workflows",
			"no run: block invokes actionlint", fix, workflows),
		scriptRow("workflow-lint", "pin-check", pinPattern,
			"a workflow fails any uses: not pinned to a 40-hex SHA",
			"no run: block greps for "+pinPattern, fix, workflows),
	}
}

func actionRow(area, check, action, canon, fix string, workflows []workflow) Row {
	var names []string
	for _, w := range workflows {
		for _, ref := range w.usesRefs() {
			if refUses(ref, action) {
				names = append(names, w.name)

				break
			}
		}
	}
	if len(names) == 0 {
		return gapRow(area, check, "no step uses "+action, canon, fix)
	}

	return okRow(area, check, strings.Join(names, ", "), canon)
}

func scriptRow(area, check, needle, canon, missing, fix string, workflows []workflow) Row {
	var names []string
	for _, w := range workflows {
		for _, script := range w.runScripts() {
			if strings.Contains(script, needle) {
				names = append(names, w.name)

				break
			}
		}
	}
	if len(names) == 0 {
		return gapRow(area, check, missing, canon, fix)
	}

	return okRow(area, check, strings.Join(names, ", "), canon)
}

func listing(label string, items []string, empty string) string {
	if len(items) == 0 {
		return empty
	}

	return fmt.Sprintf("%s: %s", label, strings.Join(items, ", "))
}
