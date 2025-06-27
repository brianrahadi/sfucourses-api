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
)

type SectionsStore struct {
	cachedSections map[string][]CourseWithSectionDetails
	lastLoaded     time.Time
	mu             sync.RWMutex
}

func NewSectionStore() (*SectionsStore, error) {
	store := &SectionsStore{
		cachedSections: make(map[string][]CourseWithSectionDetails),
	}

	if err := store.loadSections(); err != nil {
		return nil, fmt.Errorf("error loading sections: %v", err)
	}

	return store, nil
}

func (s *SectionsStore) ForceReload() error {
	return s.loadSections()
}

func (s *SectionsStore) loadSections() error {
	// Define the schedule files
	scheduleFiles := map[string]string{
		"2025-fall":   "./internal/store/json/sections/2025-fall.json",
		"2025-summer": "./internal/store/json/sections/2025-summer.json",
		"2025-spring": "./internal/store/json/sections/2025-spring.json",
		"2024-fall":   "./internal/store/json/sections/2024-fall.json",
		"2024-summer": "./internal/store/json/sections/2024-summer.json",
		"2024-spring": "./internal/store/json/sections/2024-spring.json",
	}

	newSections := make(map[string][]CourseWithSectionDetails)

	// Load each schedule file
	for term, filePath := range scheduleFiles {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", filePath, err)
		}

		var courses []CourseWithSectionDetails
		if err := json.Unmarshal(data, &courses); err != nil {
			return fmt.Errorf("error parsing JSON for term %s: %v", term, err)
		}
		newSections[term] = courses
	}

	s.mu.Lock()
	s.cachedSections = newSections
	s.lastLoaded = time.Now()
	s.mu.Unlock()

	return nil
}

func (s *SectionsStore) reloadIfNeeded() error {
	s.mu.RLock()
	shouldReload := time.Since(s.lastLoaded) > 5*time.Minute // Check every 5 minutes
	s.mu.RUnlock()

	if shouldReload {
		if err := s.loadSections(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SectionsStore) GetByTerm(ctx context.Context, year string, term string) ([]CourseWithSectionDetails, error) {
	if err := s.reloadIfNeeded(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s-%s", year, strings.ToLower(term))
	courses, found := s.cachedSections[key]
	if !found {
		return nil, ErrNotFound
	}
	return courses, nil
}

func (s *SectionsStore) GetByTermAndDept(ctx context.Context, year string, term string, dept string) ([]CourseWithSectionDetails, error) {
	courses, err := s.GetByTerm(ctx, year, term)
	if err != nil {
		return nil, err
	}

	dept = strings.ToUpper(dept)
	courses = lo.Filter(courses, func(course CourseWithSectionDetails, _ int) bool {
		return course.Dept == dept
	})

	if len(courses) == 0 {
		return nil, ErrNotFound
	}

	return courses, nil
}

func (s *SectionsStore) GetByTermAndDeptAndNumber(ctx context.Context, year string, term string, dept string, number string) (CourseWithSectionDetails, error) {
	courses, err := s.GetByTerm(ctx, year, term)
	if err != nil {
		return CourseWithSectionDetails{}, err
	}

	dept = strings.ToUpper(dept)
	number = strings.ToUpper(number)

	index := sort.Search(len(courses), func(i int) bool {
		if courses[i].Dept > dept {
			return true
		}
		if courses[i].Dept == dept {
			return courses[i].Number >= number
		}
		return false
	})

	if index < len(courses) && courses[index].Dept == dept && courses[index].Number == number {
		return courses[index], nil
	}

	return CourseWithSectionDetails{}, ErrNotFound
}
