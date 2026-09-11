package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

//go:generate go tool counterfeiter -generate

// RemoteLookup answers the checks that read the GitHub repository the
// checkout at dir pushes to, rather than the tree itself.
//
//counterfeiter:generate . RemoteLookup
type RemoteLookup interface {
	// Rulesets reports the rulesets configured on the repository.
	Rulesets(ctx context.Context, dir string) ([]Ruleset, error)
	// Repository reports the repository's own settings.
	Repository(ctx context.Context, dir string) (Repository, error)
}

// remoteLookup is the RemoteLookup a remote check uses: the one the caller
// supplied, else the gh CLI.
func (o Options) remoteLookup() RemoteLookup {
	if o.Lookup != nil {
		return o.Lookup
	}

	return ghRemoteLookup{logger: o.Logger, api: ghAPI}
}

func skippedRow(area, check, canon string) Row {
	return Row{
		Area:    area,
		Check:   check,
		Status:  StatusSkipped,
		Current: "not checked (--remote not given)",
		Canon:   canon,
	}
}

// apiFunc is one gh api call, decoding its output into into.
type apiFunc func(ctx context.Context, dir string, into any, args ...string) error

// ghRemoteLookup reads the repository through the gh CLI.
type ghRemoteLookup struct {
	logger *slog.Logger
	api    apiFunc
}

func ghAPI(ctx context.Context, dir string, into any, args ...string) error {
	out, err := run(ctx, dir, "gh", append([]string{"api"}, args...)...)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(out, into); err != nil {
		return fmt.Errorf("decode gh api %v: %w", args, err)
	}

	return nil
}

func originSlug(ctx context.Context, dir string) (string, error) {
	out, err := run(ctx, dir, "git", "remote", "get-url", "origin")
	if err != nil {
		return "", err
	}

	return parseSlug(strings.TrimSpace(string(out)))
}

func parseSlug(remote string) (string, error) {
	trimmed := strings.TrimSuffix(remote, ".git")
	if rest, ok := strings.CutPrefix(trimmed, "git@"); ok {
		if _, after, found := strings.Cut(rest, ":"); found {
			trimmed = after
		}
	}

	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) < 2 || parts[len(parts)-1] == "" || parts[len(parts)-2] == "" {
		return "", fmt.Errorf("cannot read owner/repo from origin %q", remote)
	}

	return parts[len(parts)-2] + "/" + parts[len(parts)-1], nil
}

func run(ctx context.Context, dir, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...) //nolint:gosec // argv, never a shell string; the arguments are built here
	cmd.Dir = dir

	out, err := cmd.Output()
	if err == nil {
		return out, nil
	}

	invocation := strings.Join(append([]string{name}, args...), " ")

	var exit *exec.ExitError
	if errors.As(err, &exit) && len(exit.Stderr) > 0 {
		return nil, fmt.Errorf("%s: %w: %s", invocation, err, strings.TrimSpace(string(exit.Stderr)))
	}

	return nil, fmt.Errorf("%s: %w", invocation, err)
}
