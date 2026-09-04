package mocks

type Store struct{ OpenCalls int }

func (s *Store) Open(string) error {
	s.OpenCalls++

	return nil
}
