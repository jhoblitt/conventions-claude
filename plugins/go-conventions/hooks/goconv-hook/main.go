// Command goconv-hook backs the go-conventions plugin's Claude Code hooks. It
// is invoked by hooks/goconv-hook.sh with the event's JSON on stdin and one
// argument:
//
//	format         PostToolUse on Write|Edit -- format the written .go file
//	               and report a rewrite the model has to re-read.
//	session-start  SessionStart -- inside a Go module, point the session at
//	               the canon skill.
//
// It fails open: every error path is silent and the exit status is always 0,
// because a hook that cannot do its job must not disturb the tool call or the
// session it wraps. Nothing is logged either; a hook's stdout is a channel to
// the model, not a place to complain.
//
// The additionalContext it emits is shape-constrained rather than fenced: each
// value is a literal or a field validated before use -- a .go path that was
// formatted, a module path that passed the charset check, a go directive that
// matched a version pattern -- so untrusted file content has no route into the
// model's context through this hook.
//
// Spec: skills/go-conventions/references/toolchain.md
// Callers: hooks/hooks.json
package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"
)

const (
	modeFormat       = "format"
	modeSessionStart = "session-start"

	eventPostToolUse  = "PostToolUse"
	eventSessionStart = "SessionStart"

	// Under the 30s hooks.json timeout, so a formatter that hangs is the
	// binary's own failure to report rather than the harness killing it.
	deadline = 20 * time.Second

	// The Edit payload carries the old and new strings, so the input is not
	// small; the cap only keeps an unbounded stdin from becoming unbounded
	// memory.
	maxInput = 16 << 20
)

// hookInput is the subset of the hook payload this binary reads. The fields
// come from different events: PostToolUse fills tool_name and tool_input,
// SessionStart only cwd.
type hookInput struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		FilePath string `json:"file_path"`
	} `json:"tool_input"`
	Cwd string `json:"cwd"`
}

type hookOutput struct {
	HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
}

type hookSpecificOutput struct {
	HookEventName     string `json:"hookEventName"`
	AdditionalContext string `json:"additionalContext"`
}

func main() {
	if len(os.Args) != 2 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()

	run(ctx, os.Args[1], os.Stdin, os.Stdout)
}

func run(ctx context.Context, mode string, stdin io.Reader, stdout io.Writer) {
	in, err := decodeInput(stdin)
	if err != nil {
		return
	}

	var event, additional string
	var ok bool
	switch mode {
	case modeFormat:
		event = eventPostToolUse
		additional, ok = formatContext(ctx, resolvePath(in.Cwd, in.ToolInput.FilePath))
	case modeSessionStart:
		event = eventSessionStart
		additional, ok = sessionContext(projectDir(in))
	}
	if !ok {
		return
	}

	_ = emit(stdout, event, additional) //nolint:errcheck // fail open: a failed write to a hook's only channel has nowhere to be reported.
}

func decodeInput(r io.Reader) (hookInput, error) {
	var in hookInput
	if err := json.NewDecoder(io.LimitReader(r, maxInput)).Decode(&in); err != nil {
		return hookInput{}, err
	}
	return in, nil
}

// resolvePath resolves the payload's file_path against the event's cwd. A relative one is
// relative to the session's directory, which is not where the hook process
// runs; joining it against the event's cwd is what makes the notice name the
// file that was actually formatted.
func resolvePath(cwd, path string) string {
	if path == "" || cwd == "" || filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(cwd, path)
}

// projectDir prefers the event's cwd: after the session enters a worktree that
// is the tree in play, while CLAUDE_PROJECT_DIR still names the directory the
// session started in.
func projectDir(in hookInput) string {
	if in.Cwd != "" {
		return in.Cwd
	}
	return os.Getenv("CLAUDE_PROJECT_DIR")
}

// emit writes the one shape a hook's stdout is read in. Plain text there is
// discarded for both of these events.
func emit(w io.Writer, event, additional string) error {
	return json.NewEncoder(w).Encode(hookOutput{
		HookSpecificOutput: hookSpecificOutput{
			HookEventName:     event,
			AdditionalContext: additional,
		},
	})
}
