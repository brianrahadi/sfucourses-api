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

func (s *InstructorStore) Get(ctx context.Context, dept, number, name string) ([]InstructorResponse, error) {
	if err := s.reloadIfNeeded(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if dept == "" && number == "" && name == "" {
		return s.cachedInstructors, nil
	}

	dept = strings.ToUpper(dept)
	number = strings.ToUpper(number)

	if name != "" {
		instructors := lo.Filter(s.cachedInstructors, func(instructor InstructorResponse, _ int) bool {
			return strings.Contains(strings.ToLower(instructor.Name), strings.ToLower(name))
		})
		return instructors, nil
	}

	if dept != "" && number == "" {
		instructors := lo.Filter(s.cachedInstructors, func(instructor InstructorResponse, _ int) bool {
			return lo.SomeBy(instructor.Offerings, func(offering InstructorOffering) bool {
				return offering.Dept == dept
			})
		})
		return instructors, nil
	}

	if dept != "" && number != "" {
		instructors := lo.Filter(s.cachedInstructors, func(instructor InstructorResponse, _ int) bool {
			return lo.SomeBy(instructor.Offerings, func(offering InstructorOffering) bool {
				return offering.Dept == dept && offering.Number == number
			})
		})
		return instructors, nil
	}

	return nil, ErrNotFound
}
