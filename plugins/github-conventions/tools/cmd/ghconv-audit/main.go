// Command ghconv-audit reports where a repository diverges from the
// github-conventions canon, as a markdown table or as JSON.
//
// Spec: skills/github-converge/references/ghconv-audit.md
// Callers: skills/github-converge/SKILL.md, skills/github-conventions/SKILL.md
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/cli"
)

func main() { os.Exit(start()) }

func start() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cli.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "ghconv-audit: %v\n", err)

		return 1
	}

	return 0
}
