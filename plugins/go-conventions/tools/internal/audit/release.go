package audit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Inputs are the values the two release templates are rendered with. An empty
// Binary, Owner or Repo is derived from the repository where it can be, and
// [Goreleaser] and [ReleaseWorkflow] fail naming the flag where it cannot.
type Inputs struct {
	// Binary fills {{BINARY}}; empty means the one cmd/<name> directory.
	Binary string
	// Owner and Repo fill {{OWNER}} and {{REPO}}; empty means the GitHub
	// origin remote's, else the second and third elements of a github.com
	// module path.
	Owner string
	Repo  string
	// Image opts the project into the container image. An existing
	// .goreleaser.yaml carrying a kos or docker_signs block has opted in
	// already, so this is only how a project with neither asks for one.
	Image bool
}

// Goreleaser renders templates/.goreleaser.yaml for the repository at
// opts.Dir: the placeholders filled from in, the {{IMAGE}} block kept when the
// project publishes an image, and an existing .goreleaser.yaml's goos and
// goarch lists carried into the first build in place of the template's.
func Goreleaser(ctx context.Context, opts Options, in Inputs) (string, error) {
	repo, err := scan(opts)
	if err != nil {
		return "", err
	}

	return repo.emitRelease(ctx, ".goreleaser.yaml", in, true)
}

// ReleaseWorkflow renders templates/release.yml for the repository at
// opts.Dir the same way, minus the targets: the workflow has none.
func ReleaseWorkflow(ctx context.Context, opts Options, in Inputs) (string, error) {
	repo, err := scan(opts)
	if err != nil {
		return "", err
	}

	return repo.emitRelease(ctx, "release.yml", in, false)
}

func (r *repo) emitRelease(ctx context.Context, name string, in Inputs, targets bool) (string, error) {
	// A shell redirect onto .goreleaser.yaml truncates it before this process
	// reads it, and an empty file would silently render as a project with
	// template targets and no image. No goreleaser config is legitimately
	// empty, so refuse rather than guess.
	if r.goreleaser.found && strings.TrimSpace(r.goreleaser.text) == "" {
		return "", fmt.Errorf("%s is empty: render to another path, then move it over", r.goreleaser.path)
	}
	text, err := r.template(name)
	if err != nil {
		return "", err
	}
	in, err = r.resolveInputs(ctx, in)
	if err != nil {
		return "", err
	}

	text = strings.NewReplacer(
		"{{BINARY}}", in.Binary, "{{OWNER}}", in.Owner, "{{REPO}}", in.Repo,
	).Replace(text)
	text, err = renderImage(text, r.publishesImage(in))
	if err != nil {
		return "", fmt.Errorf("template %s: %w", name, err)
	}
	if targets {
		goos, goarch, err := r.existingTargets()
		if err != nil {
			return "", err
		}
		text = carryTargets(text, goos, goarch)
	}

	return text, nil
}

// nameClass is what a binary name, an owner, a repository, or a build target
// may be spelled from: the characters goreleaser ids, ghcr.io paths, and
// GOOS/GOARCH values take. Every value lands unquoted in YAML the workflow
// runs, so anything outside it fails here rather than rendering a file that
// parses differently from how it reads.
var nameClass = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

// validate rejects a resolved input outside nameClass, naming its flag.
func (in Inputs) validate() error {
	for _, v := range []struct{ flag, value string }{
		{"--binary", in.Binary}, {"--owner", in.Owner}, {"--repo", in.Repo},
	} {
		if !nameClass.MatchString(v.value) {
			return fmt.Errorf("%s %q: not a name of [A-Za-z0-9._-]", v.flag, v.value)
		}
	}

	return nil
}

// resolveInputs fills what the caller left empty from the tree, or says which
// flag it needs, then checks every value is a name. The binary is the
// cmd/<name> directory when there is one; the owner and repository are
// [repo.slug]'s.
func (r *repo) resolveInputs(ctx context.Context, in Inputs) (Inputs, error) {
	in, err := r.deriveInputs(ctx, in)
	if err != nil {
		return in, err
	}

	return in, in.validate()
}

func (r *repo) deriveInputs(ctx context.Context, in Inputs) (Inputs, error) {
	if in.Binary == "" {
		names, err := r.commandDirs()
		if err != nil {
			return in, err
		}
		switch len(names) {
		case 1:
			in.Binary = names[0]
		case 0:
			return in, errors.New("--binary is required: no cmd/<name> directory")
		default:
			return in, fmt.Errorf("--binary is required: cmd/ holds %s", strings.Join(names, ", "))
		}
	}

	if in.Owner != "" && in.Repo != "" {
		return in, nil
	}
	owner, repo, ok, err := r.slug(ctx)
	if err != nil {
		return in, err
	}
	if ok {
		if in.Owner == "" {
			in.Owner = owner
		}
		if in.Repo == "" {
			in.Repo = repo
		}

		return in, nil
	}

	var missing []string
	if in.Owner == "" {
		missing = append(missing, "--owner")
	}
	if in.Repo == "" {
		missing = append(missing, "--repo")
	}
	verb := "is"
	if len(missing) > 1 {
		verb = "are"
	}

	return in, fmt.Errorf("%s %s required: no GitHub origin, and module %s is not github.com/<owner>/<repo>",
		strings.Join(missing, " and "), verb, r.module)
}

// commandDirs lists the cmd/<name> directories, in name order, skipping the
// ones the go tool would.
func (r *repo) commandDirs() ([]string, error) {
	root := filepath.Join(r.dir, "cmd")

	entries, err := os.ReadDir(root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", root, err)
	}

	var names []string
	for _, entry := range entries {
		if entry.IsDir() && !skipDir(entry.Name()) {
			names = append(names, entry.Name())
		}
	}

	return names, nil
}

// slug is the owner and repository the release serves: origin's when it is a
// GitHub remote, else the module path's. Where both are known and disagree —
// a fork, or a repository moved without a module-path change — neither is
// picked: the guard would otherwise name a repository the release never runs
// in, so the caller is told to choose with the flags.
func (r *repo) slug(ctx context.Context) (owner, repo string, ok bool, err error) {
	modOwner, modRepo, modOK := githubSlug(r.module)
	originOwner, originRepo, originOK, err := originSlug(ctx, r.dir)
	switch {
	case err != nil:
		return "", "", false, err
	case originOK && modOK && (!strings.EqualFold(originOwner, modOwner) || !strings.EqualFold(originRepo, modRepo)):
		return "", "", false, fmt.Errorf("--owner and --repo are required: origin is %q but module %s names %q",
			originOwner+"/"+originRepo, r.module, modOwner+"/"+modRepo)
	case originOK:
		return originOwner, originRepo, true, nil
	default:
		return modOwner, modRepo, modOK, nil
	}
}

// originSlug reads the owner and repository out of the URL of the git remote
// origin of dir, when it has one on github.com. A tree outside git, without
// origin, or with origin elsewhere has none; any other git failure is an
// error, since falling back would render the module path's guard unasked.
func originSlug(ctx context.Context, dir string) (owner, repo string, ok bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, gitTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "remote", "get-url", "origin")
	cmd.Dir = dir
	// noOrigin reads git's stderr, which a translated locale would reword.
	cmd.Env = append(cmd.Environ(), "LC_ALL=C")
	out, err := cmd.Output()
	var exit *exec.ExitError
	switch {
	case err == nil:
		owner, repo, ok = parseRemote(strings.TrimSpace(string(out)))

		return owner, repo, ok, nil
	case errors.As(err, &exit) && noOrigin(exit):
		return "", "", false, nil
	case errors.As(err, &exit):
		return "", "", false, fmt.Errorf("git remote get-url origin in %s: %w: %s",
			dir, err, strings.TrimSpace(string(exit.Stderr)))
	default:
		return "", "", false, fmt.Errorf("git remote get-url origin in %s: %w", dir, err)
	}
}

// gitTimeout bounds the one git call, which reads local config only.
const gitTimeout = 10 * time.Second

// noOrigin is git's answer for a repository without origin (exit 2) or a
// directory outside any repository (exit 128 saying so). Exit 128 also
// covers refusals such as dubious ownership, which are not an absent remote.
func noOrigin(exit *exec.ExitError) bool {
	switch exit.ExitCode() {
	case 2:
		return true
	case 128:
		return bytes.Contains(exit.Stderr, []byte("not a git repository"))
	default:
		return false
	}
}

// remoteHost matches the scheme, user, host, and port of a GitHub remote URL
// (any scheme://, any case, port allowed) or the scp-like
// git@github.com: form, where a colon separates the path and no port exists.
var remoteHost = regexp.MustCompile(
	`(?i)^(?:[a-z][a-z0-9+.-]*://(?:[^@/]+@)?github\.com(?::[0-9]+)?/|(?:[^@/:]+@)?github\.com:)`)

func parseRemote(url string) (owner, repo string, ok bool) {
	loc := remoteHost.FindStringIndex(url)
	if loc == nil {
		return "", "", false
	}
	parts := strings.Split(strings.TrimSuffix(strings.TrimSuffix(url[loc[1]:], "/"), ".git"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}

	return parts[0], parts[1], true
}

// githubSlug reads the owner and repository out of a github.com module path;
// a deeper path (github.com/o/r/v2, github.com/o/r/cmd/x) still names the
// repository in its third element.
func githubSlug(module string) (owner, repo string, ok bool) {
	parts := strings.Split(module, "/")
	if len(parts) < 3 || parts[0] != "github.com" || parts[1] == "" || parts[2] == "" {
		return "", "", false
	}

	return parts[1], parts[2], true
}

// publishesImage is the opt-in: asked for, or already in the existing
// .goreleaser.yaml as either of its two blocks. Either one means the project
// opted in; the image row separately names the missing half.
func (r *repo) publishesImage(in Inputs) bool {
	if in.Image {
		return true
	}
	if !r.goreleaser.found || r.goreleaser.parse != nil {
		return false
	}

	return r.goreleaser.hasKey("kos") || r.goreleaser.hasKey("docker_signs")
}

// existingTargets reads the goos and goarch lists of the existing
// .goreleaser.yaml's first build; nil where there is no readable file, no
// build, or no list. An item outside nameClass is an error: it would be
// written back as a list item and could not be a target.
func (r *repo) existingTargets() (goos, goarch []string, err error) {
	readable := r.goreleaser.found && r.goreleaser.parse == nil
	if !readable {
		return nil, nil, nil
	}
	builds, ok := r.goreleaser.doc["builds"].([]any)
	if !ok || len(builds) == 0 {
		return nil, nil, nil
	}

	goos, goarch = stringsAt(builds[0], "goos"), stringsAt(builds[0], "goarch")
	for key, items := range map[string][]string{"goos": goos, "goarch": goarch} {
		for _, item := range items {
			if !nameClass.MatchString(item) {
				return nil, nil, fmt.Errorf("%s: %s item %q is not a target name", r.goreleaser.path, key, item)
			}
		}
	}

	return goos, goarch, nil
}

// imageMarker is either end of an {{IMAGE}} block; the submatch is the slash
// that makes it the closing one.
var imageMarker = regexp.MustCompile(`^\s*# \{\{(/?)IMAGE\}\}\s*$`)

// renderImage removes every {{IMAGE}} marker line and keeps the lines between
// a pair when the project publishes an image, dropping them otherwise. A
// dropped block that sat between two blank lines would leave both behind, so
// the run is collapsed to one. The markers are the template's to balance: an
// unpaired one is an error, never a silently half-rendered file.
func renderImage(text string, image bool) (string, error) {
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))

	open := false
	for i := 0; i < len(lines); i++ {
		match := imageMarker.FindStringSubmatch(lines[i])
		switch {
		case match == nil:
			if image || !open {
				out = append(out, lines[i])
			}
		case match[1] == "" && !open:
			open = true
		case match[1] == "/" && open:
			open = false
			if !image && len(out) > 0 && out[len(out)-1] == "" {
				for i+1 < len(lines) && lines[i+1] == "" {
					i++
				}
			}
		default:
			return "", fmt.Errorf("line %d: unpaired {{IMAGE}} marker", i+1)
		}
	}
	if open {
		return "", errors.New("unclosed {{IMAGE}} block")
	}

	return strings.Join(out, "\n"), nil
}

// targetKey opens a goos: or goarch: list; the submatches are its indent and
// which one it is.
var targetKey = regexp.MustCompile(`^(\s*)(goos|goarch):\s*$`)

// carryTargets replaces the items under the template's first goos: and
// goarch: keys with the lists given, keeping the template's own indentation
// so its comments and layout survive. The template's first build is its only
// one, so the first key of each name is that build's. An empty list leaves
// the template's default standing.
func carryTargets(text string, goos, goarch []string) string {
	lists := map[string][]string{"goos": goos, "goarch": goarch}
	lines := strings.Split(text, "\n")
	out := make([]string, 0, len(lines))

	for i := 0; i < len(lines); {
		out = append(out, lines[i])
		match := targetKey.FindStringSubmatch(lines[i])
		i++
		if match == nil {
			continue
		}
		key, items := match[2], lists[match[2]]
		if len(items) == 0 {
			continue
		}

		end := i
		for end < len(lines) && deeper(lines[end], len(match[1])) {
			end++
		}
		indent := match[1] + "  "
		if end > i {
			indent = lines[i][:len(lines[i])-len(strings.TrimLeft(lines[i], " \t"))]
		}
		for _, item := range items {
			out = append(out, indent+"- "+item)
		}
		delete(lists, key)
		i = end
	}

	return strings.Join(out, "\n")
}
