package audit

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
	"golang.org/x/mod/modfile"
)

// goFile is one *.go file in the tree, kept as text as well as imports: some
// checks are about a call the file makes, which parser.ImportsOnly never sees.
type goFile struct {
	path    string // relative to the repository root, slash-separated
	dir     string // the file's directory, "." at the root
	pkg     string
	test    bool
	imports []string
	text    string
}

// document is a YAML file: its text, its loose tree, and the error that kept
// it from parsing. An unreadable document reports gaps, never a false pass.
type document struct {
	path  string
	found bool
	text  string
	doc   map[string]any
	parse error
}

// repo is everything the audit read, gathered once so Run, Files and Golangci
// all answer from the same scan.
type repo struct {
	opts Options
	dir  string

	module    string
	goVersion string
	toolchain string
	tools     []string

	sources []goFile

	golangci   document
	goreleaser document
	dependabot document
	workflows  []document

	makefile    string
	hasMakefile bool
	gitignore   bool
	claudeMD    string
	hasVersion  bool
}

func scan(opts Options) (*repo, error) {
	if opts.Logger == nil {
		opts.Logger = slog.New(slog.DiscardHandler)
	}

	dir := opts.Dir
	if dir == "" {
		dir = "."
	}

	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("read repository: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("read repository: %s is not a directory", dir)
	}

	r := &repo{opts: opts, dir: dir}
	if err := r.readModule(); err != nil {
		return nil, err
	}
	if err := r.walk(); err != nil {
		return nil, err
	}
	if err := r.readSupport(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *repo) readModule() error {
	path := filepath.Join(r.dir, "go.mod")

	text, found, err := readFile(path)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("%s: %w", r.dir, ErrNoModule)
	}

	parsed, err := modfile.Parse(path, []byte(text), nil)
	if err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	if parsed.Module != nil {
		r.module = parsed.Module.Mod.Path
	}
	if parsed.Go != nil {
		r.goVersion = parsed.Go.Version
	}
	if parsed.Toolchain != nil {
		r.toolchain = parsed.Toolchain.Name
	}
	for _, tool := range parsed.Tool {
		r.tools = append(r.tools, tool.Path)
	}

	return nil
}

// walk collects every *.go file outside vendor/, testdata/, and the dot- and
// underscore-led directories the go tool itself ignores.
func (r *repo) walk() error {
	fset := token.NewFileSet()

	return filepath.WalkDir(r.dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path == r.dir {
				return nil
			}
			if skipDir(entry.Name()) {
				return fs.SkipDir
			}

			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}

		text, found, err := readFile(path)
		if err != nil {
			return err
		}
		if !found {
			return nil
		}

		file := goFile{
			path: rel(r.dir, path),
			dir:  rel(r.dir, filepath.Dir(path)),
			test: strings.HasSuffix(path, "_test.go"),
			text: text,
		}
		parsed, err := parser.ParseFile(fset, path, text, parser.ImportsOnly|parser.SkipObjectResolution)
		if err != nil {
			r.opts.Logger.Warn("skipping unparsable file", slog.String("path", file.path), slog.Any("error", err))
			r.sources = append(r.sources, file)

			return nil
		}
		file.pkg = parsed.Name.Name
		for _, spec := range parsed.Imports {
			file.imports = append(file.imports, strings.Trim(spec.Path.Value, `"`))
		}
		r.sources = append(r.sources, file)

		return nil
	})
}

func skipDir(name string) bool {
	return name == "vendor" || name == "testdata" ||
		strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")
}

func (r *repo) readSupport() error {
	var err error
	if r.golangci, err = r.firstDocument(".golangci.yml", ".golangci.yaml"); err != nil {
		return err
	}
	if r.goreleaser, err = r.firstDocument(".goreleaser.yaml", ".goreleaser.yml"); err != nil {
		return err
	}
	if r.dependabot, err = r.firstDocument(filepath.Join(".github", "dependabot.yml")); err != nil {
		return err
	}
	if r.workflows, err = r.readWorkflows(); err != nil {
		return err
	}

	if r.makefile, r.hasMakefile, err = readFile(filepath.Join(r.dir, "Makefile")); err != nil {
		return err
	}
	if r.gitignore, err = exists(filepath.Join(r.dir, ".gitignore")); err != nil {
		return err
	}
	if r.claudeMD, _, err = readFile(filepath.Join(r.dir, "CLAUDE.md")); err != nil {
		return err
	}
	r.hasVersion, err = exists(filepath.Join(r.dir, "internal", "version"))

	return err
}

// firstDocument reads the first of names that exists; a missing file yields a
// document that reports itself absent rather than an error.
func (r *repo) firstDocument(names ...string) (document, error) {
	for _, name := range names {
		text, found, err := readFile(filepath.Join(r.dir, name))
		if err != nil {
			return document{}, err
		}
		if !found {
			continue
		}

		return newDocument(filepath.ToSlash(name), text), nil
	}

	return document{path: filepath.ToSlash(names[0])}, nil
}

func newDocument(path, text string) document {
	d := document{path: path, found: true, text: text}
	d.parse = yaml.Unmarshal([]byte(text), &d.doc)

	return d
}

func (r *repo) readWorkflows() ([]document, error) {
	root := filepath.Join(r.dir, ".github", "workflows")

	entries, err := os.ReadDir(root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", root, err)
	}

	docs := make([]document, 0, len(entries))
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
		docs = append(docs, newDocument(filepath.ToSlash(filepath.Join(".github", "workflows", entry.Name())), text))
	}

	return docs, nil
}

// unreadable names the workflows that did not parse, so a check that found
// nothing says where it could not look.
func (r *repo) unreadable() []string {
	var names []string
	for _, w := range r.workflows {
		if w.parse != nil {
			names = append(names, w.path)
		}
	}

	return names
}

// usesAction reports the workflows with a step whose uses: names action. The
// tree is walked rather than the text: a uses: inside a run: script is a shell
// line, not a step, and a quoted value is the parser's to unquote.
func (r *repo) usesAction(action string) []string {
	var names []string
	for _, w := range r.workflows {
		for _, ref := range w.usesRefs() {
			if refUses(ref, action) {
				names = append(names, w.path)

				break
			}
		}
	}

	return names
}

// runsMatching reports the workflows with a run: block satisfying match.
func (r *repo) runsMatching(match func(string) bool) []string {
	var names []string
	for _, w := range r.workflows {
		if slices.ContainsFunc(w.runScripts(), match) {
			names = append(names, w.path)
		}
	}

	return names
}

func (d document) usesRefs() []string {
	var refs []string
	walkMaps(d.doc, func(node map[string]any) {
		if ref, ok := node["uses"].(string); ok {
			refs = append(refs, ref)
		}
	})

	return refs
}

func (d document) runScripts() []string {
	var scripts []string
	walkMaps(d.doc, func(node map[string]any) {
		if script, ok := node["run"].(string); ok {
			scripts = append(scripts, script)
		}
	})

	return scripts
}

// hasKey reports whether the document carries a non-nil top-level key.
func (d document) hasKey(key string) bool { return d.doc[key] != nil }

// refUses reports whether ref names action, allowing the subdirectory and
// version suffixes GitHub permits (golang/govulncheck-action/foo@<sha>).
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

// stringsAt walks a path of map keys and renders the sequence it ends on.
func stringsAt(node any, keys ...string) []string {
	for _, key := range keys {
		typed, ok := node.(map[string]any)
		if !ok {
			return nil
		}
		node = typed[key]
	}
	items, ok := node.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		out = append(out, fmt.Sprint(item))
	}

	return out
}

// scalarAt walks a path of map keys and renders the scalar it ends on; the
// second result is false when the key is absent.
func scalarAt(node any, keys ...string) (string, bool) {
	for _, key := range keys {
		typed, ok := node.(map[string]any)
		if !ok {
			return "", false
		}
		node = typed[key]
	}
	if node == nil {
		return "", false
	}

	return fmt.Sprint(node), true
}

func readFile(path string) (text string, found bool, err error) {
	data, err := os.ReadFile(path) //nolint:gosec // the path is under the directory the caller asked to audit
	if errors.Is(err, fs.ErrNotExist) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("read %s: %w", path, err)
	}

	return string(data), true, nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat %s: %w", path, err)
	}

	return true, nil
}

func listing(label string, items []string, empty string) string {
	if len(items) == 0 {
		return empty
	}

	return label + ": " + strings.Join(items, ", ")
}
