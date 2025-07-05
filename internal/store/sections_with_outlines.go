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

type SectionsWithOutlineStore struct {
	cachedSectionsWithOutlines map[string][]CourseOutlineWithSectionDetails
	lastLoaded                 time.Time
	mu                         sync.RWMutex
}

func NewSectionsWithOutlineStore() (*SectionsWithOutlineStore, error) {
	store := &SectionsWithOutlineStore{
		cachedSectionsWithOutlines: make(map[string][]CourseOutlineWithSectionDetails),
	}

	if err := store.loadSectionsWithOutlines(); err != nil {
		return nil, fmt.Errorf("error loading sections with outlines: %v", err)
	}

	return store, nil
}

func (s *SectionsWithOutlineStore) loadSectionsWithOutlines() error {
	outlinesData, err := os.ReadFile("./internal/store/json/outlines.json")
	if err != nil {
		return fmt.Errorf("error reading outlines file: %v", err)
	}

	var outlines []CourseOutline
	if err := json.Unmarshal(outlinesData, &outlines); err != nil {
		return fmt.Errorf("error parsing outlines JSON: %v", err)
	}

	outlineMap := make(map[string]CourseOutline)
	for _, outline := range outlines {
		key := fmt.Sprintf("%s-%s", outline.Dept, outline.Number)
		outlineMap[key] = outline
	}

	scheduleFiles := map[string]string{
		"2025-spring": "./internal/store/json/sections/2025-spring.json",
		"2024-fall":   "./internal/store/json/sections/2024-fall.json",
		"2024-summer": "./internal/store/json/sections/2024-summer.json",
		"2024-spring": "./internal/store/json/sections/2024-spring.json",
		"2025-summer": "./internal/store/json/sections/2025-summer.json",
		"2025-fall":   "./internal/store/json/sections/2025-fall.json",
	}

	newSectionsWithOutlines := make(map[string][]CourseOutlineWithSectionDetails)

	// Load each schedule file and merge with outlines
	for term, filePath := range scheduleFiles {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading file %s: %v", filePath, err)
		}

		var coursesWithSections []CourseWithSectionDetails
		if err := json.Unmarshal(data, &coursesWithSections); err != nil {
			return fmt.Errorf("error parsing JSON for term %s: %v", term, err)
		}

		var sectionsWithOutlines []CourseOutlineWithSectionDetails
		for _, courseWithSections := range coursesWithSections {
			key := fmt.Sprintf("%s-%s", courseWithSections.Dept, courseWithSections.Number)
			outline, found := outlineMap[key]
			if found {
				sectionsWithOutlines = append(sectionsWithOutlines, CourseOutlineWithSectionDetails{
					Dept:           outline.Dept,
					Number:         outline.Number,
					Title:          outline.Title,
					Units:          outline.Units,
					Description:    outline.Description,
					Designation:    outline.Designation,
					DeliveryMethod: outline.DeliveryMethod,
					Prerequisites:  outline.Prerequisites,
					Corequisites:   outline.Corequisites,
					Term:           courseWithSections.Term,
					SectionDetails: courseWithSections.SectionDetails,
				})
			}
		}

		newSectionsWithOutlines[term] = sectionsWithOutlines
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.cachedSectionsWithOutlines = newSectionsWithOutlines
	s.lastLoaded = time.Now()

	return nil
}

func (s *SectionsWithOutlineStore) ForceReload() error {
	return s.loadSectionsWithOutlines()
}

func (s *SectionsWithOutlineStore) reloadIfNeeded() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	shouldReload := time.Since(s.lastLoaded) > 5*time.Minute // Check every 5 minutes

	if shouldReload {
		if err := s.loadSectionsWithOutlines(); err != nil {
			return err
		}
	}
	return nil
}

func (s *SectionsWithOutlineStore) Get(ctx context.Context, year, term, dept, number string) ([]CourseOutlineWithSectionDetails, error) {
	if err := s.reloadIfNeeded(); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if year == "" || term == "" {
		return nil, ErrNotFound
	}

	key := fmt.Sprintf("%s-%s", year, strings.ToLower(term))
	sectionsWithOutline, found := s.cachedSectionsWithOutlines[key]
	if !found {
		return nil, ErrNotFound
	}

	if dept == "" && number == "" {
		return sectionsWithOutline, nil
	}

	dept = strings.ToUpper(dept)
	number = strings.ToUpper(number)

	if dept != "" && number == "" {
		courses := lo.Filter(sectionsWithOutline, func(course CourseOutlineWithSectionDetails, _ int) bool {
			return course.Dept == dept
		})
		return courses, nil
	}

	if dept != "" && number != "" {
		courses := lo.Filter(sectionsWithOutline, func(course CourseOutlineWithSectionDetails, _ int) bool {
			return course.Dept == dept && course.Number == number
		})
		return courses, nil
	}

	return nil, ErrNotFound
}
