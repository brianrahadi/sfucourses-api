package store

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	. "github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/samber/lo"
)

//go:embed json/sections/2025-spring.json
var spring2025Courses []byte

//go:embed json/sections/2024-fall.json
var fall2024Courses []byte

//go:embed json/sections/2024-summer.json
var summer2024Courses []byte

//go:embed json/sections/2024-spring.json
var spring2024Courses []byte

type SectionsStore struct {
	cachedCourses map[string][]CourseWithSectionDetails
}

func NewSectionStore() (*SectionsStore, error) {
	// Initialize a map of raw JSON data for each schedule
	scheduleMap := map[string][]byte{
		"2025-spring": spring2025Courses,
		"2024-fall":   fall2024Courses,
		"2024-summer": summer2024Courses,
		"2024-spring": spring2024Courses,
	}

	// Initialize the CoursesStore
	store := &SectionsStore{
		cachedCourses: make(map[string][]CourseWithSectionDetails),
	}

	// Unmarshal each schedule and cache it in memory
	for term, data := range scheduleMap {
		var courses []CourseWithSectionDetails
		if err := json.Unmarshal(data, &courses); err != nil {
			return nil, fmt.Errorf("error parsing JSON for term %s: %v", term, err)
		}
		store.cachedCourses[term] = courses
	}

	return store, nil
}

func (s *SectionsStore) GetByTerm(ctx context.Context, year string, term string) ([]CourseWithSectionDetails, error) {
	key := fmt.Sprintf("%s-%s", year, strings.ToLower(term))
	courses, found := s.cachedCourses[key]
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
