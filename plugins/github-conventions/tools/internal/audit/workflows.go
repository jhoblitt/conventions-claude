package audit

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

// pinPattern is the grep the pin-check job runs; the audit looks for it
// verbatim rather than reimplementing the job.
const pinPattern = "[0-9a-f]{40}"

var (
	usesLine    = regexp.MustCompile(`(?m)^[ \t]*(?:-[ \t]*)?uses:[ \t]*(\S+)(.*)$`)
	shaPinned   = regexp.MustCompile(`@[0-9a-f]{40}$`)
	versionNote = regexp.MustCompile(`^[ \t]+#[ \t]*v`)
)

// workflow is one file under .github/workflows, kept both as text (the pin
// check reads comments, which the parser drops) and as a loose tree.
type workflow struct {
	name  string
	text  string
	doc   map[string]any
	parse error
}

func loadWorkflows(dir string) ([]workflow, error) {
	root := filepath.Join(dir, ".github", "workflows")

	entries, err := os.ReadDir(root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", root, err)
	}

	workflows := make([]workflow, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !slices.Contains([]string{".yml", ".yaml"}, filepath.Ext(entry.Name())) {
			continue
		}

		text, found, err := readFile(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}

		w := workflow{name: entry.Name(), text: text}
		w.parse = yaml.Unmarshal([]byte(text), &w.doc)
		workflows = append(workflows, w)
	}

	return workflows, nil
}

func perWorkflowRows(workflows []workflow) []Row {
	rows := make([]Row, 0, 5*len(workflows))
	for _, w := range workflows {
		rows = append(rows, w.rows()...)
	}

	return rows
}

func (w workflow) rows() []Row {
	if w.parse != nil {
		current := "yaml parse error: " + w.parse.Error()
		rows := make([]Row, 0, 5)
		for _, check := range []string{"pinned", "permissions", "timeout", "concurrency", "checkout-credentials"} {
			rows = append(rows, gapRow("workflows", check+":"+w.name, current,
				"the workflow parses as YAML", "fix the YAML so the workflow can be read"))
		}

		return rows
	}

	return []Row{w.pinnedRow(), w.permissionsRow(), w.timeoutRow(), w.concurrencyRow(), w.checkoutRow()}
}

func (w workflow) check(name string) string { return name + ":" + w.name }

func (w workflow) pinnedRow() Row {
	const (
		canon = "every uses: ends in a 40-hex commit SHA with a # v… comment"
		fix   = "run pinact run over .github/workflows"
	)

	// The parser drops comments, and the canon is about the "# v…" comment as
	// much as the SHA, so this one check reads the file as text.
	var floating []string
	for _, m := range usesLine.FindAllStringSubmatch(w.text, -1) {
		ref, trailer := strings.Trim(m[1], "\"'"), m[2]
		if ref == "." || strings.HasPrefix(ref, "./") || strings.HasPrefix(ref, "docker://") {
			continue
		}
		if shaPinned.MatchString(ref) && versionNote.MatchString(trailer) {
			continue
		}
		floating = append(floating, ref)
	}
	if len(floating) == 0 {
		return okRow("workflows", w.check("pinned"), "every uses: is SHA-pinned", canon)
	}

	return gapRow("workflows", w.check("pinned"), listing("unpinned", floating, ""), canon, fix)
}

func (w workflow) permissionsRow() Row {
	const (
		canon     = "a top-level permissions key, with no write scope"
		fixAbsent = "add a top-level permissions: contents: read"
		fixWrite  = "move the write scopes onto the jobs that need them"
	)
	check := w.check("permissions")

	switch perms := w.doc["permissions"].(type) {
	case nil:
		return gapRow("workflows", check, "no top-level permissions key", canon, fixAbsent)
	case string:
		if strings.Contains(perms, "write") {
			return gapRow("workflows", check, "permissions: "+perms, canon, fixWrite)
		}

		return okRow("workflows", check, "permissions: "+perms, canon)
	case map[string]any:
		scopes := make([]string, 0, len(perms))
		writes := false
		for scope, level := range perms {
			scopes = append(scopes, fmt.Sprintf("%s: %v", scope, level))
			if fmt.Sprint(level) == "write" {
				writes = true
			}
		}
		slices.Sort(scopes)
		if writes {
			return gapRow("workflows", check, strings.Join(scopes, ", "), canon, fixWrite)
		}

		return okRow("workflows", check, join(scopes, "no scopes granted"), canon)
	default:
		return gapRow("workflows", check, fmt.Sprintf("unreadable permissions value %v", perms), canon, fixAbsent)
	}
}

func (w workflow) timeoutRow() Row {
	const (
		canon = "every job sets timeout-minutes"
		fix   = "add timeout-minutes to each job"
	)
	check := w.check("timeout")

	jobs, isMap := w.doc["jobs"].(map[string]any)
	if !isMap || len(jobs) == 0 {
		return okRow("workflows", check, "no jobs", canon)
	}

	// GitHub rejects timeout-minutes on a job that calls a reusable workflow;
	// the timeout belongs to the jobs inside the called workflow.
	checked := 0
	var missing []string
	for name, job := range jobs {
		fields, isMap := job.(map[string]any)
		if isMap && fields["uses"] != nil {
			continue
		}
		checked++
		if !isMap || fields["timeout-minutes"] == nil {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		slices.Sort(missing)

		return gapRow("workflows", check, listing("no timeout-minutes on", missing, ""), canon, fix)
	}
	if checked == 0 {
		return okRow("workflows", check, strconv.Itoa(len(jobs))+" reusable-workflow job(s)", canon)
	}

	return okRow("workflows", check, strconv.Itoa(checked)+" job(s), all with timeout-minutes", canon)
}

func (w workflow) concurrencyRow() Row {
	const (
		canon = "a top-level concurrency key"
		fix   = "add a top-level concurrency group keyed on the workflow and the ref"
	)
	check := w.check("concurrency")

	if w.doc["concurrency"] == nil {
		return gapRow("workflows", check, "no top-level concurrency key", canon, fix)
	}

	return okRow("workflows", check, "concurrency group set", canon)
}

func (w workflow) checkoutRow() Row {
	const (
		canon = "every actions/checkout step sets with.persist-credentials: false"
		fix   = "add with.persist-credentials: false to every actions/checkout step"
	)
	check := w.check("checkout-credentials")

	steps := 0
	leaking := 0
	walkMaps(w.doc, func(node map[string]any) {
		ref, ok := node["uses"].(string)
		if !ok || !refUses(ref, "actions/checkout") {
			return
		}
		steps++
		with, hasWith := node["with"].(map[string]any)
		if !hasWith {
			leaking++

			return
		}
		if persist, isBool := with["persist-credentials"].(bool); !isBool || persist {
			leaking++
		}
	})

	switch {
	case steps == 0:
		return okRow("workflows", check, "no actions/checkout steps", canon)
	case leaking == 0:
		return okRow("workflows", check, fmt.Sprintf("persist-credentials: false on all %d", steps), canon)
	default:
		return gapRow("workflows", check, fmt.Sprintf("persist-credentials not false on %d of %d", leaking, steps), canon, fix)
	}
}

func (w workflow) usesRefs() []string {
	var refs []string
	walkMaps(w.doc, func(node map[string]any) {
		if ref, ok := node["uses"].(string); ok {
			refs = append(refs, ref)
		}
	})

	return refs
}

func (w workflow) runScripts() []string {
	var scripts []string
	walkMaps(w.doc, func(node map[string]any) {
		if script, ok := node["run"].(string); ok {
			scripts = append(scripts, script)
		}
	})

	return scripts
}

// refUses reports whether ref names action, allowing the subdirectory and
// version suffixes GitHub permits (github/codeql-action/init@<sha>).
func refUses(ref, action string) bool {
	rest, ok := strings.CutPrefix(ref, action)
	if !ok {
		return false
	}

	return rest == "" || rest[0] == '@' || rest[0] == '/'
}

func walkMaps(node any, visit func(map[string]any)) {
	switch typed := node.(type) {
	case map[string]any:
		visit(typed)
		for _, child := range typed {
			walkMaps(child, visit)
		}
	case []any:
		for _, child := range typed {
			walkMaps(child, visit)
		}
	}
}
