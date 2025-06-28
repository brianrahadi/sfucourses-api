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

type InstructorStore struct {
	cachedInstructors []InstructorResponse
	lastLoaded        time.Time
	mu                sync.RWMutex
	filePath          string
}

func NewInstructorStore() (*InstructorStore, error) {
	store := &InstructorStore{
		filePath: "./internal/store/json/instructors.json",
	}

	if err := store.loadInstructors(); err != nil {
		return nil, fmt.Errorf("error loading instructors: %v", err)
	}

	return store, nil
}

func (s *InstructorStore) ForceReload() error {
	return s.loadInstructors()
}

func (s *InstructorStore) loadInstructors() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return fmt.Errorf("error reading file %s: %v", s.filePath, err)
	}

	var instructors []InstructorResponse
	if err := json.Unmarshal(data, &instructors); err != nil {
		return fmt.Errorf("error parsing JSON: %v", err)
	}

	// Sort instructors by name for consistent ordering and binary search
	sort.Slice(instructors, func(i, j int) bool {
		return strings.ToLower(instructors[i].Name) < strings.ToLower(instructors[j].Name)
	})

	s.mu.Lock()
	s.cachedInstructors = instructors
	s.lastLoaded = time.Now()
	s.mu.Unlock()

	return nil
}

func (s *InstructorStore) reloadIfNeeded() error {
	s.mu.RLock()
	shouldReload := time.Since(s.lastLoaded) > time.Hour
	s.mu.RUnlock()

	if shouldReload {
		if err := s.loadInstructors(); err != nil {
			return err
		}
	}
	return nil
}

func (s *InstructorStore) GetAll(ctx context.Context) ([]InstructorResponse, error) {
	if err := s.reloadIfNeeded(); err != nil {
		return nil, err
	}

	return s.cachedInstructors, nil
}

func (s *InstructorStore) GetByDept(ctx context.Context, dept string) ([]InstructorResponse, error) {
	if err := s.reloadIfNeeded(); err != nil {
		return nil, err
	}

	dept = strings.ToUpper(dept)
	instructors := lo.Filter(s.cachedInstructors, func(instructor InstructorResponse, _ int) bool {
		return lo.SomeBy(instructor.Offerings, func(offering InstructorOffering) bool {
			return offering.Dept == dept
		})
	})

	if len(instructors) == 0 {
		return instructors, ErrNotFound
	}

	return instructors, nil
}

func (s *InstructorStore) GetByDeptAndNumber(ctx context.Context, dept string, number string) ([]InstructorResponse, error) {
	if err := s.reloadIfNeeded(); err != nil {
		return nil, err
	}

	dept = strings.ToUpper(dept)
	number = strings.ToUpper(number)

	instructors := lo.Filter(s.cachedInstructors, func(instructor InstructorResponse, _ int) bool {
		return lo.SomeBy(instructor.Offerings, func(offering InstructorOffering) bool {
			return offering.Dept == dept && offering.Number == number
		})
	})

	if len(instructors) == 0 {
		return instructors, ErrNotFound
	}

	return instructors, nil
}

func (s *InstructorStore) GetByName(ctx context.Context, name string) ([]InstructorResponse, error) {
	if err := s.reloadIfNeeded(); err != nil {
		return nil, err
	}

	// Use binary search since instructors are sorted by name
	index := sort.Search(len(s.cachedInstructors), func(i int) bool {
		return strings.ToLower(s.cachedInstructors[i].Name) >= strings.ToLower(name)
	})

	instructors := make([]InstructorResponse, 0)
	if index < len(s.cachedInstructors) && strings.EqualFold(s.cachedInstructors[index].Name, name) {
		instructors = append(instructors, s.cachedInstructors[index])
	}

	if len(instructors) == 0 {
		return nil, ErrNotFound
	}

	return instructors, nil
}
