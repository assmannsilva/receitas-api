package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

var errNotFound = errors.New("receita não encontrada")

type Store struct {
	mu      sync.RWMutex
	path    string
	recipes []Recipe
}

func NewStore(path string) (*Store, error) {
	s := &Store{path: path, recipes: []Recipe{}}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal(data, &s.recipes); err != nil {
		return nil, err
	}
	if s.recipes == nil {
		s.recipes = []Recipe{}
	}
	return s, nil
}

func (s *Store) List() []Recipe {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Recipe, len(s.recipes))
	for i, r := range s.recipes {
		result[i] = r
		result[i].Ingredients = append([]string{}, r.Ingredients...)
	}
	return result
}

// Write a new snapshot before publishing it in memory. Rename is atomic on
// the same filesystem; failed writes leave the previous state untouched.
func (s *Store) save(next []Recipe) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(s.path), ".recipes-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), s.path); err != nil {
		return err
	}
	s.recipes = next
	return nil
}

func (s *Store) Create(r Recipe) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := append(append([]Recipe{}, s.recipes...), r)
	return s.save(next)
}

func (s *Store) Update(id string, r Recipe) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := append([]Recipe{}, s.recipes...)
	for i, old := range next {
		if old.ID == id {
			next[i] = r
			return s.save(next)
		}
	}
	return errNotFound
}

func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	next := make([]Recipe, 0, len(s.recipes))
	found := false
	for _, r := range s.recipes {
		if r.ID == id {
			found = true
		} else {
			next = append(next, r)
		}
	}
	if !found {
		return errNotFound
	}
	return s.save(next)
}
