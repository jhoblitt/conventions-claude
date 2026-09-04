package main

import (
	"testing"
)

func TestValidModulePath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"example.com/x", true},
		{"github.com/jhoblitt/conventions-claude", true},
		{"example.com/x/v2", true},
		{"a-b_c.d~e/f", true},
		{"Example.com/X", true},
		{"", false},
		{"example.com/x y", false},
		{"example.com/x\ty", false},
		{"example.com/x\ny", false},
		{`"example.com/x"`, false},
		{"example.com/x;rm -rf /", false},
		{"example.com/x\\y", false},
		{"exam ple", false},
		{"example.com//x", false},
		{"/", false},
		{"/example.com/x", false},
		{"example.com/x/", false},
		{".", false},
		{"..", false},
		{"../../etc/passwd", false},
		{"example.com/./x", false},
		{"example.com/../x", false},
		{"-example.com/x", false},
		{"example.com/-x", false},
		{"example.com/..x", true},
		{"example.com/x-", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := validModulePath(tt.path); got != tt.want {
				t.Errorf("validModulePath(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestParseGoMod(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantOK     bool
		wantModule string
		wantGo     string
	}{
		{
			name:       "module and go directive",
			content:    "module example.com/x\n\ngo 1.27\n",
			wantOK:     true,
			wantModule: "example.com/x",
			wantGo:     "1.27",
		},
		{
			name:       "patch version and trailing content",
			content:    "module example.com/x\n\ngo 1.27.1\n\nrequire (\n\tfoo v1.0.0\n)\n",
			wantOK:     true,
			wantModule: "example.com/x",
			wantGo:     "1.27.1",
		},
		{
			name:       "tabs and trailing spaces",
			content:    "module\texample.com/x   \ngo\t1.27  \n",
			wantOK:     true,
			wantModule: "example.com/x",
			wantGo:     "1.27",
		},
		{name: "module line has a space in it", content: "module example.com/x y\n\ngo 1.27\n"},
		{name: "no module line", content: "go 1.27\n"},
		{name: "no go directive", content: "module example.com/x\n"},
		{name: "go directive is not a version", content: "module example.com/x\n\ngo tip\n"},
		{name: "module directive indented is not a directive", content: "  module example.com/x\n\ngo 1.27\n"},
		{name: "empty", content: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mod, ver, ok := parseGoMod([]byte(tt.content))
			if ok != tt.wantOK {
				t.Fatalf("parseGoMod() ok = %v (%q, %q), want %v", ok, mod, ver, tt.wantOK)
			}
			if !ok {
				return
			}
			if mod != tt.wantModule {
				t.Errorf("module = %q, want %q", mod, tt.wantModule)
			}
			if ver != tt.wantGo {
				t.Errorf("go = %q, want %q", ver, tt.wantGo)
			}
		})
	}
}

func TestSessionContext(t *testing.T) {
	t.Run("go module", func(t *testing.T) {
		dir := t.TempDir()
		writeGoMod(t, dir, "module example.com/x\n\ngo 1.27\n")

		got, ok := sessionContext(dir)
		if !ok {
			t.Fatal("sessionContext() ok = false, want true")
		}
		const want = "Go module example.com/x (go 1.27): load go-conventions:go-conventions before writing Go"
		if got != want {
			t.Errorf("sessionContext() = %q, want %q", got, want)
		}
	})

	t.Run("silent cases", func(t *testing.T) {
		noMod := t.TempDir()
		badMod := t.TempDir()
		writeGoMod(t, badMod, "module example.com/x y\n\ngo 1.27\n")

		for _, dir := range []string{"", noMod, badMod} {
			if got, ok := sessionContext(dir); ok {
				t.Errorf("sessionContext(%q) = %q, true; want silence", dir, got)
			}
		}
	})
}
