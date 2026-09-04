package run

import (
	"bytes"
	"testing"
)

func TestGreet(t *testing.T) {
	var buf bytes.Buffer
	if err := Greet(&buf, "world"); err != nil {
		t.Fatalf("Greet: %v", err)
	}
	if got := buf.String(); got != "hello, world\n" {
		t.Fatalf("got %q", got)
	}
}
