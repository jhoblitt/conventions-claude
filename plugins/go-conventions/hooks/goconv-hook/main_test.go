package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecodeInput(t *testing.T) {
	tests := []struct {
		name     string
		stdin    string
		wantErr  bool
		wantPath string
		wantCwd  string
	}{
		{
			name:     "post tool use",
			stdin:    `{"tool_name":"Write","tool_input":{"file_path":"/w/x.go"},"cwd":"/w"}`,
			wantPath: "/w/x.go",
			wantCwd:  "/w",
		},
		{
			name:    "session start carries only cwd",
			stdin:   `{"session_id":"abc","cwd":"/w"}`,
			wantCwd: "/w",
		},
		{
			name:    "not json",
			stdin:   "not json at all",
			wantErr: true,
		},
		{
			name:    "empty",
			stdin:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := decodeInput(strings.NewReader(tt.stdin))
			if (err != nil) != tt.wantErr {
				t.Fatalf("decodeInput() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if got.ToolInput.FilePath != tt.wantPath {
				t.Errorf("FilePath = %q, want %q", got.ToolInput.FilePath, tt.wantPath)
			}
			if got.Cwd != tt.wantCwd {
				t.Errorf("Cwd = %q, want %q", got.Cwd, tt.wantCwd)
			}
		})
	}
}

func TestEmit(t *testing.T) {
	var buf bytes.Buffer
	if err := emit(&buf, eventSessionStart, "hello"); err != nil {
		t.Fatalf("emit() error = %v", err)
	}

	var got struct {
		HookSpecificOutput struct {
			HookEventName     string `json:"hookEventName"`
			AdditionalContext string `json:"additionalContext"`
		} `json:"hookSpecificOutput"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("emit() wrote invalid JSON %q: %v", buf.String(), err)
	}
	if got.HookSpecificOutput.HookEventName != eventSessionStart {
		t.Errorf("hookEventName = %q, want %q", got.HookSpecificOutput.HookEventName, eventSessionStart)
	}
	if got.HookSpecificOutput.AdditionalContext != "hello" {
		t.Errorf("additionalContext = %q, want %q", got.HookSpecificOutput.AdditionalContext, "hello")
	}
}

func TestRunFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.go")
	if err := os.WriteFile(path, []byte("package bad\nfunc  F( )  {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	stdin := `{"tool_name":"Write","tool_input":{"file_path":` + quote(t, path) + `},"cwd":` + quote(t, dir) + `}`
	var out bytes.Buffer
	run(t.Context(), modeFormat, strings.NewReader(stdin), &out)

	if !strings.Contains(out.String(), `"hookEventName":"PostToolUse"`) {
		t.Fatalf("run() stdout = %q, want a PostToolUse payload", out.String())
	}
	if !strings.Contains(out.String(), "re-read it before the next Edit") {
		t.Errorf("run() stdout = %q, want the re-read notice", out.String())
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != "package bad\n\nfunc F() {}\n" {
		t.Errorf("file after run() = %q, want it formatted", after)
	}
}

func TestResolvePath(t *testing.T) {
	tests := []struct {
		name string
		cwd  string
		path string
		want string
	}{
		{"relative path joins the event cwd", "/w", "pkg/x.go", "/w/pkg/x.go"},
		{"absolute path is left alone", "/w", "/other/x.go", "/other/x.go"},
		{"no cwd leaves the path alone", "", "pkg/x.go", "pkg/x.go"},
		{"empty path stays empty", "/w", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolvePath(tt.cwd, tt.path); got != tt.want {
				t.Errorf("resolvePath(%q, %q) = %q, want %q", tt.cwd, tt.path, got, tt.want)
			}
		})
	}
}

// A payload's file_path may be relative, and it is relative to the session's
// cwd -- never to wherever the hook process happens to be running.
func TestRunFormatResolvesARelativePath(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "pkg"), 0o750); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "pkg", "bad.go")
	if err := os.WriteFile(path, []byte("package bad\nfunc  F( )  {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	// A decoy at the same relative path under the process's own cwd: formatting
	// this one instead is the bug the resolution exists to prevent.
	decoy := filepath.Join(t.TempDir(), "pkg")
	if err := os.MkdirAll(decoy, 0o750); err != nil {
		t.Fatal(err)
	}
	const unformatted = "package bad\nfunc  D( )  {}\n"
	if err := os.WriteFile(filepath.Join(decoy, "bad.go"), []byte(unformatted), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Dir(decoy))

	stdin := `{"tool_name":"Write","tool_input":{"file_path":"pkg/bad.go"},"cwd":` + quote(t, dir) + `}`
	var out bytes.Buffer
	run(t.Context(), modeFormat, strings.NewReader(stdin), &out)

	if !strings.Contains(out.String(), path) {
		t.Errorf("run() stdout = %q, want it to name the resolved path %q", out.String(), path)
	}

	after, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != "package bad\n\nfunc F() {}\n" {
		t.Errorf("file under cwd = %q, want it formatted", after)
	}

	untouched, err := os.ReadFile(filepath.Join(decoy, "bad.go"))
	if err != nil {
		t.Fatal(err)
	}
	if string(untouched) != unformatted {
		t.Errorf("the file under the process cwd was formatted: %q", untouched)
	}
}

func TestRunSilent(t *testing.T) {
	dir := t.TempDir()
	clean := filepath.Join(dir, "clean.go")
	if err := os.WriteFile(clean, []byte("package clean\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		mode  string
		stdin string
	}{
		{"unknown mode", "explode", `{"cwd":"` + dir + `"}`},
		{"malformed json", modeFormat, "{{{"},
		{"non-go path", modeFormat, `{"tool_name":"Write","tool_input":{"file_path":"/w/README.md"}}`},
		{"missing file", modeFormat, `{"tool_name":"Edit","tool_input":{"file_path":"/w/nope.go"}}`},
		{"already formatted", modeFormat, `{"tool_name":"Write","tool_input":{"file_path":"` + clean + `"}}`},
		{"session start without go.mod", modeSessionStart, `{"cwd":"` + dir + `"}`},
		{"session start without cwd", modeSessionStart, `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CLAUDE_PROJECT_DIR", "")
			var out bytes.Buffer
			run(t.Context(), tt.mode, strings.NewReader(tt.stdin), &out)
			if out.Len() != 0 {
				t.Errorf("run() stdout = %q, want empty", out.String())
			}
		})
	}
}

func TestRunSessionStartFallsBackToProjectDir(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir, "module example.com/fallback\n\ngo 1.27\n")
	t.Setenv("CLAUDE_PROJECT_DIR", dir)

	var out bytes.Buffer
	run(t.Context(), modeSessionStart, strings.NewReader(`{}`), &out)

	if !strings.Contains(out.String(), "Go module example.com/fallback (go 1.27)") {
		t.Errorf("run() stdout = %q, want the module notice", out.String())
	}
}

func quote(t *testing.T, s string) string {
	t.Helper()
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func writeGoMod(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
