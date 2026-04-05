package credentials

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Store persists API keys and tokens outside the repo (e.g. $HOME/.perch/credentials as JSON).
type Store struct {
	path string
	mu   sync.Mutex
}

// NewStore returns a store that reads/writes path (caller supplies absolute path).
func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) readFile() (map[string]string, error) {
	raw, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return map[string]string{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return map[string]string{}, nil
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("credentials: %w", err)
	}
	if m == nil {
		m = map[string]string{}
	}
	return m, nil
}

func (s *Store) writeFile(m map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Set persists key=value (overwrites).
func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.readFile()
	if err != nil {
		return err
	}
	m[key] = value
	return s.writeFile(m)
}

// Get returns a stored secret.
func (s *Store) Get(key string) (string, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, err := s.readFile()
	if err != nil {
		return "", false, err
	}
	v, ok := m[key]
	return v, ok, nil
}
