// Package store is the widget store other modules import.
package store

import (
	"errors"

	"example.com/widget/internal/index"
)

// ErrMissing reports a key the store does not hold.
var ErrMissing = errors.New("no such key")

// Store maps keys to widgets.
type Store struct{ idx *index.Index }

// New returns an empty store.
func New() *Store { return &Store{idx: index.New()} }

// Get returns the widget stored under key.
func (s *Store) Get(key string) (string, error) {
	value, ok := s.idx.Lookup(key)
	if !ok {
		return "", ErrMissing
	}

	return value, nil
}
