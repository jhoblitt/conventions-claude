package store

import "github.com/pkg/errors"

type Store struct{ path string }

func Open(path string) (*Store, error) {
	if path == "" {
		return nil, errors.New("empty path")
	}

	return &Store{path: path}, nil
}
