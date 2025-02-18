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

//go:embed json/outlines.json
var outlinesJSON2 []byte

type SectionsWithOutlineStore struct {
	cachedSectionsWithOutlines map[string][]CourseOutlineWithSectionDetails
}

func NewSectionsWithOutlineStore() (*SectionsWithOutlineStore, error) {
	// Initialize a map of raw JSON data for each schedule
	scheduleMap := map[string][]byte{
		"2025-spring": spring2025Courses,
		"2024-fall":   fall2024Courses,
		"2024-summer": summer2024Courses,
		"2024-spring": spring2024Courses,
		"2025-summer": summer2025Courses,
	}

	var outlines []CourseOutline
	if err := json.Unmarshal(outlinesJSON2, &outlines); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	outlineMap := make(map[string]CourseOutline)
	for _, outline := range outlines {
		key := fmt.Sprintf("%s-%s", outline.Dept, outline.Number)
		outlineMap[key] = outline
	}

	// Initialize the CoursesStore
	store := &SectionsWithOutlineStore{
		cachedSectionsWithOutlines: make(map[string][]CourseOutlineWithSectionDetails),
	}

	// Unmarshal each schedule and cache it in memory
	for term, data := range scheduleMap {
		var coursesWithSections []CourseWithSectionDetails
		if err := json.Unmarshal(data, &coursesWithSections); err != nil {
			return nil, fmt.Errorf("error parsing JSON for term %s: %v", term, err)
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

		store.cachedSectionsWithOutlines[term] = sectionsWithOutlines
	}

	return store, nil
}

func (s *SectionsWithOutlineStore) GetByTerm(ctx context.Context, year string, term string) ([]CourseOutlineWithSectionDetails, error) {
	key := fmt.Sprintf("%s-%s", year, strings.ToLower(term))
	sectionsWithOutline, found := s.cachedSectionsWithOutlines[key]
	if !found {
		return nil, ErrNotFound
	}
	return sectionsWithOutline, nil
}

func (s *SectionsWithOutlineStore) GetByTermAndDept(ctx context.Context, year string, term string, dept string) ([]CourseOutlineWithSectionDetails, error) {
	courses, err := s.GetByTerm(ctx, year, term)
	if err != nil {
		return nil, err
	}
	dept = strings.ToUpper(dept)
	courses = lo.Filter(courses, func(course CourseOutlineWithSectionDetails, _ int) bool {
		return course.Dept == dept
	})

	if len(courses) == 0 {
		return nil, ErrNotFound
	}

	return courses, nil
}

func (s *SectionsWithOutlineStore) GetByTermAndDeptAndNumber(ctx context.Context, year string, term string, dept string, number string) (CourseOutlineWithSectionDetails, error) {
	courses, err := s.GetByTerm(ctx, year, term)
	if err != nil {
		return CourseOutlineWithSectionDetails{}, err
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

	return CourseOutlineWithSectionDetails{}, ErrNotFound
}
