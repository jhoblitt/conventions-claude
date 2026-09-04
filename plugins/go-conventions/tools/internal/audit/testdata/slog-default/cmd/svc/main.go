package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"example.com/svc/internal/serve"
)

// Logs through whatever slog's default handler is: text, to stderr, installed
// by nobody. The canon's handler is never built.
func main() {
	cmd := &cobra.Command{
		Use: "svc",
		RunE: func(cmd *cobra.Command, _ []string) error {
			slog.InfoContext(cmd.Context(), "starting", slog.String("addr", ":8080"))

			return serve.Listen(context.Background(), ":8080")
		},
	}
	if err := cmd.Execute(); err != nil {
		slog.Error("exiting", slog.Any("error", err))
		os.Exit(1)
	}
}
