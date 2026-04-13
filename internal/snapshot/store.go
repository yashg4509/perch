// Package snapshot persists named deploy-SHA sets for cross-stack rollback (spec slice; CLI wiring later).
package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Record is one saved cross-stack SHA set (spec: snapshot save / rollback).
type Record struct {
	Name      string            `json:"name"`
	CreatedAt time.Time         `json:"created_at"`
	SHAs      map[string]string `json:"shas"` // logical node name -> git/deploy SHA
}

// Store persists records as JSON (path is usually under the project or ~/.perch).
type Store struct {
	path string
}

// NewStore returns a store writing to path (file will be created).
func NewStore(path string) *Store {
	return &Store{path: path}
}

func (s *Store) load() ([]Record, error) {
	raw, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	var recs []Record
	if err := json.Unmarshal(raw, &recs); err != nil {
		return nil, fmt.Errorf("snapshot: %w", err)
	}
	return recs, nil
}

func (s *Store) saveAll(recs []Record) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(recs, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

// Save appends a new record (does not dedupe names).
func (s *Store) Save(rec Record) error {
	recs, err := s.load()
	if err != nil {
		return err
	}
	if rec.SHAs == nil {
		rec.SHAs = map[string]string{}
	}
	recs = append(recs, rec)
	return s.saveAll(recs)
}

// List returns all records in creation order.
func (s *Store) List() ([]Record, error) {
	return s.load()
}
