package index

import (
	"context"
	"log/slog"
)

// Index is the store's key lookup.
type Index struct{ entries map[string]string }

// New returns an empty index.
func New() *Index { return &Index{entries: map[string]string{}} }

// Lookup returns the value stored under key. The library logs through whatever
// default handler its caller installed; it installs none of its own.
func (i *Index) Lookup(key string) (string, bool) {
	slog.DebugContext(context.Background(), "index lookup", slog.String("key", key))
	value, ok := i.entries[key]

	return value, ok
}
