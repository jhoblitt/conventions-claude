// Command goconv-audit reports where a Go repository diverges from the
// go-conventions canon, emits the house lint config for its current imports,
// and lists the files an area's migration touches.
//
// Spec: skills/go-converge/references/goconv-audit.md
// Callers: skills/go-converge/SKILL.md, skills/go-review/SKILL.md,
// skills/go-conventions/SKILL.md
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jhoblitt/conventions-claude/plugins/go-conventions/tools/internal/cli"
)

func main() { os.Exit(start()) }

func start() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cli.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "goconv-audit: %v\n", err)

		return 1
	}

	return 0
}
