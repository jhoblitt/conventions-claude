package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatterName(t *testing.T) {
	t.Run("gofumpt when it is on PATH", func(t *testing.T) {
		shim := t.TempDir()
		stub := filepath.Join(shim, "gofumpt")
		if err := os.WriteFile(stub, []byte("#!/bin/sh\nexit 0\n"), 0o700); err != nil {
			t.Fatal(err)
		}
		t.Setenv("PATH", shim)

		if got := formatterName(); got != "gofumpt" {
			t.Errorf("formatterName() = %q, want gofumpt", got)
		}
	})

	t.Run("gofmt when it is not", func(t *testing.T) {
		t.Setenv("PATH", t.TempDir())

		if got := formatterName(); got != "gofmt" {
			t.Errorf("formatterName() = %q, want gofmt", got)
		}
	})
}

func TestFormatContext(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name     string
		file     string
		content  string
		wantOK   bool
		wantKeep bool
	}{
		{
			name:    "rewrites a badly formatted file",
			file:    "bad.go",
			content: "package bad\nfunc  F( )  {}\n",
			wantOK:  true,
		},
		{
			name:     "silent on an already formatted file",
			file:     "clean.go",
			content:  "package clean\n",
			wantKeep: true,
		},
		{
			name:     "silent on a non-go path",
			file:     "notes.md",
			content:  "#  heading\n",
			wantKeep: true,
		},
		{
			name:     "silent on a file that does not parse",
			file:     "broken.go",
			content:  "package broken\nfunc (\n",
			wantKeep: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, tt.file)
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatal(err)
			}

			got, ok := formatContext(t.Context(), path)
			if ok != tt.wantOK {
				t.Fatalf("formatContext() ok = %v (%q), want %v", ok, got, tt.wantOK)
			}
			if ok {
				if !strings.Contains(got, path) {
					t.Errorf("formatContext() = %q, want it to name %q", got, path)
				}
				if !strings.HasSuffix(got, " on disk; re-read it before the next Edit") {
					t.Errorf("formatContext() = %q, want the re-read notice", got)
				}
				if !strings.HasPrefix(got, formatterName()+" rewrote ") {
					t.Errorf("formatContext() = %q, want it to name the formatter", got)
				}
			}
			if tt.wantKeep {
				after, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if string(after) != tt.content {
					t.Errorf("file changed to %q, want it untouched", after)
				}
			}
		})
	}
}

func TestFormatContextMissingFile(t *testing.T) {
	got, ok := formatContext(t.Context(), filepath.Join(t.TempDir(), "absent.go"))
	if ok {
		t.Errorf("formatContext() = %q, true; want silence for a missing file", got)
	}
}
