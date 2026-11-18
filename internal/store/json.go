package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/yecharlot/NEXO/internal/domain"
)

type Store struct {
	mu   sync.RWMutex
	path string
	db   domain.Database
}

func Open(path string) (*Store, error) {
	s := &Store{path: path, db: domain.Database{Projects: map[string]domain.Project{}, Assets: map[string]domain.Asset{}}}
	b, err := os.ReadFile(path)
	if err == nil {
		if err = json.Unmarshal(b, &s.db); err != nil {
			return nil, fmt.Errorf("decode database: %w", err)
		}
	}
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	if s.db.Projects == nil {
		s.db.Projects = map[string]domain.Project{}
	}
	if s.db.Assets == nil {
		s.db.Assets = map[string]domain.Asset{}
	}
	return s, nil
}
func (s *Store) persistLocked() error {
	b, err := json.MarshalIndent(s.db, "", "  ")
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err = os.WriteFile(tmp, b, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
func (s *Store) Snapshot() domain.Database {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, _ := json.Marshal(s.db)
	var cp domain.Database
	_ = json.Unmarshal(b, &cp)
	return cp
}
func (s *Store) PutProject(p domain.Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db.Projects[p.ID] = p
	return s.persistLocked()
}
func (s *Store) PutAsset(a domain.Asset) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db.Assets[a.ID] = a
	return s.persistLocked()
}
func (s *Store) GetAsset(id string) (domain.Asset, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.db.Assets[id]
	return a, ok
}
func (s *Store) ActivatePro() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.db.License = domain.License{Plan: "pro"}
	return s.persistLocked()
}
func (s *Store) IsPro() bool { s.mu.RLock(); defer s.mu.RUnlock(); return s.db.License.Plan == "pro" }
