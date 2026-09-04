package run

import (
	"fmt"
	"io"
)

func Greet(w io.Writer, name string) error {
	_, err := fmt.Fprintf(w, "hello, %s\n", name)

	return err
}
