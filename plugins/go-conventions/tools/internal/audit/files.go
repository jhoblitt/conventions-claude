package audit

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

// areas are the migrations go-converge dispatches, each naming the files a
// go-migrator run would rewrite.
var areas = []string{"cli", "logging", "testing", "errors", "layout", "version"}

func (r *repo) files(area string) ([]string, error) {
	var paths []string

	switch area {
	case "cli":
		paths = r.cliFiles()
	case "logging":
		paths = r.loggingFiles()
	case "testing":
		var err error
		if paths, err = r.testingFiles(); err != nil {
			return nil, err
		}
	case "errors":
		paths = r.importers(func(f goFile) bool {
			return importMatches(f.imports, "github.com/pkg/errors", false)
		})
	case "layout":
		paths = r.layoutFiles()
	case "version":
		var err error
		if paths, err = r.versionFiles(); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown area %q: one of %s", area, strings.Join(areas, ", "))
	}

	slices.Sort(paths)

	return slices.Compact(paths), nil
}

func (r *repo) cliFiles() []string {
	paths := r.importers(func(f goFile) bool {
		return importMatches(f.imports, "flag", true) ||
			importMatches(f.imports, "github.com/alecthomas/kong", false) ||
			importMatches(f.imports, "github.com/urfave/cli", false)
	})

	for _, f := range r.sources {
		if f.dir == "internal/cli" || strings.HasPrefix(f.dir, "internal/cli/") {
			paths = append(paths, f.path)

			continue
		}
		if strings.HasPrefix(f.dir, "cmd/") && strings.Count(f.dir, "/") == 1 &&
			filepath.Base(f.path) == "main.go" {
			paths = append(paths, f.path)
		}
	}

	return paths
}

func (r *repo) loggingFiles() []string {
	return r.importers(func(f goFile) bool {
		for _, pkg := range loggers {
			if importMatches(f.imports, pkg, pkg == "log") {
				return true
			}
		}
		// A binary on slog with no handler of its own logs through the default
		// text handler, which logging/stderr-json reports; the command that has
		// to install one is where that migration lands, and without this the row
		// could name a gap the list has no file for.
		if f.pkg == "main" && !f.test && importMatches(f.imports, "log/slog", true) {
			return true
		}

		return strings.Contains(f.text, "slog.New(") || strings.Contains(f.text, "slog.SetDefault(")
	})
}

// testingFiles are the specs a testing migration rewrites plus the generated
// double directories it replaces with counterfeiter fakes.
func (r *repo) testingFiles() ([]string, error) {
	paths := r.importers(func(f goFile) bool {
		if !f.test {
			return false
		}

		return importMatches(f.imports, "github.com/stretchr/testify", false) ||
			(importMatches(f.imports, "testing", true) &&
				!importMatches(f.imports, "github.com/onsi/ginkgo/v2", false))
	})

	err := filepath.WalkDir(r.dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() || path == r.dir {
			return nil
		}
		if skipDir(entry.Name()) {
			return fs.SkipDir
		}
		if strings.HasSuffix(entry.Name(), "fakes") || entry.Name() == "mocks" {
			paths = append(paths, rel(r.dir, path))
		}

		return nil
	})

	return paths, err
}

func (r *repo) layoutFiles() []string {
	return r.importers(func(f goFile) bool {
		return f.dir != "cmd" && !strings.HasPrefix(f.dir, "cmd/") &&
			f.dir != "internal" && !strings.HasPrefix(f.dir, "internal/")
	})
}

// stampedVars are the package-level variables an -X ldflag writes into.
var stampedVars = []string{"version", "Version", "commit", "date"}

func (r *repo) versionFiles() ([]string, error) {
	paths := r.importers(declaresStamp)

	for _, name := range []string{".goreleaser.yaml", ".goreleaser.yml", "Makefile"} {
		found, err := exists(filepath.Join(r.dir, name))
		if err != nil {
			return nil, err
		}
		if found {
			paths = append(paths, name)
		}
	}

	return append(paths, r.runsMatching(ldflagsX.MatchString)...), nil
}

// declaresStamp reports whether the file declares a package-level variable an
// -X ldflag could be stamping. parser.ImportsOnly stops at the import block,
// so the declarations are parsed here instead.
func declaresStamp(f goFile) bool {
	parsed, err := parser.ParseFile(token.NewFileSet(), f.path, f.text, parser.SkipObjectResolution)
	if err != nil {
		return false
	}

	for _, decl := range parsed.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			value, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, name := range value.Names {
				if slices.Contains(stampedVars, name.Name) {
					return true
				}
			}
		}
	}

	return false
}

// importers lists the repository-relative paths of the Go files matching keep.
func (r *repo) importers(keep func(goFile) bool) []string {
	var paths []string
	for _, f := range r.sources {
		if keep(f) {
			paths = append(paths, f.path)
		}
	}

	return paths
}
