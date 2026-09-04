package main

import (
	"flag"
	"log/slog"
	"os"

	"github.com/pkg/errors"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if err := serve(*addr); err != nil {
		slog.Error("serve failed", "err", errors.Wrap(err, "serve"))
		os.Exit(1)
	}
}
