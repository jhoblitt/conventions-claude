package audit

import (
	"errors"
	"regexp"
	"strings"
)

// errTemplateRoot is what every template read returns when the plugin
// directory is unknown. The templates are the one home of the config the tool
// emits, so there is nothing to fall back to: --emit-golangci fails, and the
// rows measured against a template skip.
var errTemplateRoot = errors.New("CLAUDE_PLUGIN_ROOT unset: the canon templates cannot be read")

type missingTemplateError struct{ path string }

func (e *missingTemplateError) Error() string { return "no such template: " + e.path }

// denyEntry opens one depguard deny entry in templates/.golangci.yml. The
// entry runs to the next line indented no deeper than the dash.
var denyEntry = regexp.MustCompile(`^(\s*)- pkg: *"?([^"]*?)"? *$`)

// emitGolangci renders the canon lint config for this repository: the template
// with its module path filled in, minus every depguard deny entry an import
// already present would trip on. Dropping those keeps the config green on
// install; the area migration restores the entry by re-running this command
// once the imports are gone.
func (r *repo) emitGolangci() (string, error) {
	text, err := r.template(".golangci.yml")
	if err != nil {
		return "", err
	}

	return dropDenied(strings.ReplaceAll(text, "{{MODULE}}", r.module), r.importPaths()), nil
}

func dropDenied(text string, imports []string) string {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))

	denyLine, kept := -1, 0
	for i := 0; i < len(lines); {
		if strings.TrimSpace(lines[i]) == "deny:" {
			denyLine, kept = len(out), 0
		}

		match := denyEntry.FindStringSubmatch(lines[i])
		if match == nil {
			out = append(out, lines[i])
			i++

			continue
		}

		end := i + 1
		for end < len(lines) && deeper(lines[end], len(match[1])) {
			end++
		}
		if !denied(match[2], imports) {
			out = append(out, lines[i:end]...)
			kept++
		}
		i = end
	}

	// A deny list emptied of every entry would leave a dangling key.
	if denyLine >= 0 && kept == 0 {
		out = append(out[:denyLine], out[denyLine+1:]...)
	}

	return strings.Join(out, "\n")
}

// deeper reports whether line is a continuation of an entry opened at indent.
func deeper(line string, indent int) bool {
	trimmed := strings.TrimLeft(line, " \t")

	return trimmed != "" && len(line)-len(trimmed) > indent
}

// denied reports whether any import matches a depguard pkg pattern: a prefix,
// or an exact path when the pattern is anchored with $.
func denied(pattern string, imports []string) bool {
	exact := strings.HasSuffix(pattern, "$")

	return importMatches(imports, strings.TrimSuffix(pattern, "$"), exact)
}
