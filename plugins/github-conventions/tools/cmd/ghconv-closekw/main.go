// Command ghconv-closekw reports every GitHub closing keyword that sits
// directly before an issue or pull request reference, in a pull request's
// title, description, and commit messages as GitHub holds them (or a local
// draft and commit range), so each can be confirmed as a close that is meant
// before the PR merges.
//
// Spec: skills/github-conventions/references/pull-requests.md
// Callers: skills/github-conventions/SKILL.md
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/jhoblitt/conventions-claude/plugins/github-conventions/tools/internal/closekw"
)

func main() { os.Exit(start()) }

func start() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := closekw.Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "ghconv-closekw: %v\n", err)

		return 1
	}

	return 0
}
