package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	. "github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/samber/lo"
)

type OutlineStore struct {
	cachedOutlines []CourseOutline
	lastLoaded     time.Time
	mu             sync.RWMutex
	filePath       string
}

func NewOutlineStore() (*OutlineStore, error) {
	store := &OutlineStore{
		filePath: "./internal/store/json/outlines.json",
	}

	if err := store.loadOutlines(); err != nil {
		return nil, fmt.Errorf("error loading outlines: %v", err)
	}

	return store, nil
}

func (s *OutlineStore) ForceReload() error {
	return s.loadOutlines()
}

func (s *OutlineStore) loadOutlines() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", s.filePath, err)
	}

	var outlines []CourseOutline
	if err := json.Unmarshal(data, &outlines); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	s.mu.Lock()
	s.cachedOutlines = outlines
	s.lastLoaded = time.Now()
	s.mu.Unlock()

	return nil
}

func (s *OutlineStore) reloadIfNeeded() error {
	s.mu.RLock()
	shouldReload := time.Since(s.lastLoaded) > time.Hour
	s.mu.RUnlock()

	if shouldReload {
		if err := s.loadOutlines(); err != nil {
			return err
		}
	}
	return nil
}

func (s *OutlineStore) Get(ctx context.Context, dept, number string) ([]CourseOutline, error) {
	if err := s.reloadIfNeeded(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if dept == "" && number == "" {
		return s.cachedOutlines, nil
	}

	dept = strings.ToUpper(dept)
	number = strings.ToUpper(number)

	if dept != "" && number == "" {
		outlines := lo.Filter(s.cachedOutlines, func(outline CourseOutline, _ int) bool {
			return outline.Dept == dept
		})
		return outlines, nil
	}

	if dept != "" && number != "" {
		outlines := lo.Filter(s.cachedOutlines, func(outline CourseOutline, _ int) bool {
			return outline.Dept == dept && outline.Number == number
		})
		return outlines, nil
	}

	return nil, ErrNotFound
}
