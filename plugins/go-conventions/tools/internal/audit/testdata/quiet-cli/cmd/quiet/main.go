package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// A command that only prints its result: it reaches for no logger at all.
func main() {
	cmd := &cobra.Command{
		Use: "quiet",
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "ok")

			return nil
		},
	}
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "quiet:", err)
		os.Exit(1)
	}
}
