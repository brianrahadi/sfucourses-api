package store

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	. "github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/samber/lo"
	"github.com/samber/mo"
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

func (s *OutlineStore) GetAll(ctx context.Context, limitOpt mo.Option[int], offsetOpt mo.Option[int]) ([]CourseOutline, int, error) {
	if err := s.reloadIfNeeded(); err != nil {
		return nil, 0, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	totalCount := len(s.cachedOutlines)
	if limitOpt.IsAbsent() && offsetOpt.IsAbsent() {
		return s.cachedOutlines, totalCount, nil
	}
	limit := limitOpt.OrElse(totalCount)
	offset := offsetOpt.OrElse(0)

	if offset >= totalCount {
		return []CourseOutline{}, 0, nil
	}

	end := offset + limit
	if end > totalCount {
		end = totalCount
	}

	return s.cachedOutlines[offset:end], totalCount, nil
}

func (s *OutlineStore) GetByDept(ctx context.Context, dept string) ([]CourseOutline, error) {
	if err := s.reloadIfNeeded(); err != nil {
		return nil, err
	}

	dept = strings.ToUpper(dept)
	outlines := lo.Filter(s.cachedOutlines, func(outline CourseOutline, _ int) bool {
		return outline.Dept == dept
	})

	if len(outlines) == 0 {
		return outlines, ErrNotFound
	}

	return outlines, nil
}

func (s *OutlineStore) GetByDeptAndNumber(ctx context.Context, dept string, number string) (CourseOutline, error) {
	if err := s.reloadIfNeeded(); err != nil {
		return CourseOutline{}, err
	}

	dept = strings.ToUpper(dept)
	number = strings.ToUpper(number)

	index := sort.Search(len(s.cachedOutlines), func(i int) bool {
		if s.cachedOutlines[i].Dept > dept {
			return true
		}
		if s.cachedOutlines[i].Dept == dept {
			return s.cachedOutlines[i].Number >= number
		}
		return false
	})

	if index < len(s.cachedOutlines) && s.cachedOutlines[index].Dept == dept && s.cachedOutlines[index].Number == number {
		return s.cachedOutlines[index], nil
	}

	return CourseOutline{}, ErrNotFound
}
