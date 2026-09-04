package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// The whole line is captured and then charset-checked, rather than
	// captured as one non-space run: `module example.com/x y` is not a module
	// path, and a regexp that stops at the space would call it one.
	//
	// The cost of matching whole lines is that forms go.mod allows but this
	// does not read -- a trailing comment (`module example.com/x // why`) and
	// a pre-release go directive (`go 1.27rc1`) -- are treated as an absent
	// directive: the hook stays silent rather than guessing at them.
	moduleRE = regexp.MustCompile(`(?m)^module[ \t]+(.*?)[ \t\r]*$`)
	goRE     = regexp.MustCompile(`(?m)^go[ \t]+(\d+(?:\.\d+){1,2})[ \t\r]*$`)

	modulePathRE = regexp.MustCompile(`^[A-Za-z0-9\-._~/]+$`)
)

// sessionContext returns the pointer at the canon for a session rooted in a Go
// module, and nothing at all otherwise -- no go.mod, an unreadable one, or one
// whose module line does not survive validModulePath.
func sessionContext(dir string) (string, bool) {
	if dir == "" {
		return "", false
	}

	data, err := os.ReadFile(filepath.Join(dir, "go.mod")) //nolint:gosec // G304: go.mod under the session root the harness reported.
	if err != nil {
		return "", false
	}

	modulePath, goVersion, ok := parseGoMod(data)
	if !ok {
		return "", false
	}

	return "Go module " + modulePath + " (go " + goVersion +
		"): load go-conventions:go-conventions before writing Go", true
}

// parseGoMod pulls the module path and the go directive out of a go.mod. It
// reads the two lines it needs rather than the file's grammar: the caller
// wants a sentence for the model, and a go.mod this cannot make sense of is
// one the hook stays quiet about.
func parseGoMod(data []byte) (modulePath, goVersion string, ok bool) {
	m := moduleRE.FindSubmatch(data)
	g := goRE.FindSubmatch(data)
	if m == nil || g == nil {
		return "", "", false
	}

	modulePath = string(m[1])
	if !validModulePath(modulePath) {
		return "", "", false
	}

	return modulePath, string(g[1]), true
}

// validModulePath checks the shape of a module path, and only the shape: every
// character is an ASCII letter, digit, or one of -._~/, and every
// slash-separated element is non-empty (so no `//`, no leading or trailing
// slash), is neither `.` nor `..`, and does not start with `-`.
//
// This is NOT golang.org/x/mod/module.CheckPath -- x/mod is not stdlib and
// hook binaries are the stdlib-only carve-out -- and it does not claim that
// check's coverage: it accepts paths the real one rejects (reserved names,
// element-level dot rules, the case-encoding rules) and rejects some it
// accepts. The bias is deliberate. A path this rejects costs the session a
// pointer it can live without; a path it wrongly accepts is text going into
// the model's context.
func validModulePath(p string) bool {
	if p == "" || !modulePathRE.MatchString(p) {
		return false
	}
	for elem := range strings.SplitSeq(p, "/") {
		if elem == "" || elem == "." || elem == ".." || strings.HasPrefix(elem, "-") {
			return false
		}
	}
	return true
}
