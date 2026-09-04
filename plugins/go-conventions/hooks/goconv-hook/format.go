package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

// formatterName reports which formatter to run. gofumpt is the canon's
// formatter and a superset of gofmt; gofmt is always there, so a session
// without gofumpt installed still gets formatted files.
func formatterName() string {
	if _, err := exec.LookPath("gofumpt"); err == nil {
		return "gofumpt"
	}
	return "gofmt"
}

// formatContext formats path in place and returns the notice the model needs
// when the file on disk no longer matches what it just wrote. A file that was
// already formatted -- the common case -- produces nothing, and so does every
// failure: an unparseable file, a missing formatter, a path outside the
// session's reach.
func formatContext(ctx context.Context, path string) (string, bool) {
	if filepath.Ext(path) != ".go" {
		return "", false
	}

	before, err := os.ReadFile(path) //nolint:gosec // G304: the path the hooked tool call just wrote is the whole job.
	if err != nil {
		return "", false
	}

	tool := formatterName()
	// The path is one argv element handed to exec, never a shell word: this
	// starts no shell, so there is nothing for a crafted name to expand into.
	if err = exec.CommandContext(ctx, tool, "-w", path).Run(); err != nil { //nolint:gosec // G204: a fixed argv with no shell to expand it.
		return "", false
	}

	after, err := os.ReadFile(path) //nolint:gosec // G304: re-reading that same path.
	if err != nil || bytes.Equal(before, after) {
		return "", false
	}

	return tool + " rewrote " + path + " on disk; re-read it before the next Edit", true
}
