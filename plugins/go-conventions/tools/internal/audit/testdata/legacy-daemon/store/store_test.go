package store

import "testing"

func TestOpen(t *testing.T) {
	if _, err := Open(""); err == nil {
		t.Fatal("want an error for the empty path")
	}
}
