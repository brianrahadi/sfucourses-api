// This file should be saved as scripts/fetchSections/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/brianrahadi/sfucourses-api/internal/model"
	utils "github.com/brianrahadi/sfucourses-api/scripts"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

// Configuration for concurrency
const (
	MaxConcurrentDepartments = 10
	MaxConcurrentCourses     = 5
	MaxConcurrentSections    = 20
	RequestTimeout           = 10 * time.Second
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: program <year> <term>")
		fmt.Println("Example: program 2025 spring")
		os.Exit(1)
	}

	year := os.Args[1]
	term := os.Args[2]

	fmt.Printf("Fetching sections for %s %s...\n", term, year)
	startTime := time.Now()

	resultFilePath := fmt.Sprintf("./internal/store/json/sections/%s-%s.json", year, term)

	// Create a context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	// Initialize the container for course with section details
	var courseWithSectionDetailsMapContainer = mo.Right[map[string]model.CourseOutline](make(map[string]model.CourseWithSectionDetails))

	// Process the term with concurrency
	if err := processTerm(ctx, year, term, courseWithSectionDetailsMapContainer); err != nil {
		fmt.Printf("Error processing term %s %s: %v\n", year, term, err)
		os.Exit(1)
	}

	// Extract and process the map of course with section details
	courseWithSectionDetailsMap := courseWithSectionDetailsMapContainer.RightOrEmpty()
	courseWithSectionDetails := slices.Collect(maps.Values(courseWithSectionDetailsMap))

	// Remove bad data
	courseWithSectionDetails = lo.Filter(courseWithSectionDetails, func(courseWithSectionDetails model.CourseWithSectionDetails, _ int) bool {
		return courseWithSectionDetails.Dept != "" && courseWithSectionDetails.Number != ""
	})

	// Sort each course's sections by section code
	for i := range courseWithSectionDetails {
		// Sort the section details by section code
		slices.SortFunc(courseWithSectionDetails[i].SectionDetails, func(a model.SectionDetail, b model.SectionDetail) int {
			return strings.Compare(a.Section, b.Section)
		})
	}

	// Sort courses by department and number
	slices.SortFunc(courseWithSectionDetails, func(a model.CourseWithSectionDetails, b model.CourseWithSectionDetails) int {
		if a.Dept != b.Dept {
			return strings.Compare(a.Dept, b.Dept)
		}
		return strings.Compare(a.Number, b.Number)
	})

	// Marshal to JSON
	jsonData, err := json.Marshal(courseWithSectionDetails)
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v\n", err)
		os.Exit(1)
	}

	// Write to file
	err = os.WriteFile(resultFilePath, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		os.Exit(1)
	}

	elapsedTime := time.Since(startTime)
	fmt.Printf("Successfully wrote course data to %s\n", resultFilePath)
	fmt.Printf("Total time: %s\n", elapsedTime)
	fmt.Printf("Processed %d courses\n", len(courseWithSectionDetails))
}

// processTerm fetches all departments and processes them concurrently
func processTerm(ctx context.Context, year, term string, courseMap mo.Either[map[string]model.CourseOutline, map[string]model.CourseWithSectionDetails]) error {
	// Fetch departments
	depts, err := utils.GetDepartments(year, term)
	if err != nil {
		return fmt.Errorf("error getting departments for term %s: %w", term, err)
	}

	fmt.Printf("Found %d departments\n", len(depts))

	// Create a semaphore to limit concurrent department processing
	depSemaphore := make(chan struct{}, MaxConcurrentDepartments)
	var wg sync.WaitGroup

	// Create a mutex for courseMap
	var mu sync.Mutex

	// Process each department concurrently
	for _, dept := range depts {
		// Skip departments with empty Value
		if dept.Value == "" {
			fmt.Printf("Skipping department with empty value: %+v\n", dept)
			continue
		}

		wg.Add(1)
		depSemaphore <- struct{}{} // Acquire semaphore

		go func(department utils.DepartmentRes) {
			defer wg.Done()
			defer func() { <-depSemaphore }() // Release semaphore

			deptCtx, deptCancel := context.WithTimeout(ctx, 5*time.Minute)
			defer deptCancel()

			// Use department.Value instead of department.Name
			err := processDepartment(deptCtx, year, term, department.Value, courseMap, &mu)
			if err != nil {
				fmt.Printf("Error processing department %s: %v\n", department.Value, err)
			}
		}(dept)
	}

	// Wait for all departments to be processed
	wg.Wait()
	return nil
}

// processDepartment fetches all courses for a department and processes them concurrently
func processDepartment(ctx context.Context, year, term, dept string, courseMap mo.Either[map[string]model.CourseOutline, map[string]model.CourseWithSectionDetails], mu *sync.Mutex) error {
	// Skip if context is done
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Skip empty department
	if dept == "" {
		return fmt.Errorf("empty department value")
	}

	// Fetch courses for the department
	courses, err := utils.GetCourses(year, term, dept)
	if err != nil {
		return fmt.Errorf("error getting courses for department %s: %w", dept, err)
	}

	fmt.Printf("Department %s: Found %d courses\n", dept, len(courses))

	// Create a semaphore to limit concurrent course processing
	courseSemaphore := make(chan struct{}, MaxConcurrentCourses)
	var wg sync.WaitGroup

	// Process sections only for CourseWithSectionDetails
	if courseMap.IsRight() {
		sectionDetailsMap := courseMap.RightOrEmpty()

		// Process each course concurrently
		for _, course := range courses {
			// Skip courses with empty Title
			if course.Title == "" {
				fmt.Printf("Skipping course with empty title in department %s\n", dept)
				continue
			}

			wg.Add(1)
			courseSemaphore <- struct{}{} // Acquire semaphore

			go func(courseObj utils.CourseRes) {
				defer wg.Done()
				defer func() { <-courseSemaphore }() // Release semaphore

				courseCtx, courseCancel := context.WithTimeout(ctx, 2*time.Minute)
				defer courseCancel()

				err := processSectionDetails(courseCtx, year, term, dept, courseObj.Value, sectionDetailsMap, mu)
				if err != nil {
					fmt.Printf("Error processing course %s %s: %v\n", dept, courseObj.Value, err)
				}
			}(course)
		}
	}

	// Wait for all courses to be processed
	wg.Wait()
	return nil
}

// processSectionDetails fetches all sections for a course and processes them concurrently
func processSectionDetails(ctx context.Context, year, term, dept, number string, courseWithSectionDetailsMap map[string]model.CourseWithSectionDetails, mu *sync.Mutex) error {
	// Skip if context is done
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Validate inputs to prevent malformed requests
	if dept == "" || number == "" {
		return fmt.Errorf("empty department or course number: dept=%s, number=%s", dept, number)
	}

	// Fetch sections for the course
	sections, err := utils.GetSections(year, term, dept, number)
	if err != nil {
		return fmt.Errorf("error getting sections for course %s: %w", number, err)
	}

	if len(sections) == 0 {
		fmt.Printf("No sections found for %s %s\n", dept, number)
		return nil
	}

	// Create a slice to hold the section detail raw data
	sectionDetailRawArr := make([]model.SectionDetailRaw, 0, len(sections))
	var sectionMu sync.Mutex

	// Create a semaphore to limit concurrent section processing
	sectionSemaphore := make(chan struct{}, MaxConcurrentSections)
	var wg sync.WaitGroup

	// Process each section concurrently
	for _, section := range sections {
		// Skip sections with empty Title
		if section.Title == "" {
			fmt.Printf("Skipping section with empty title for course %s %s\n", dept, number)
			continue
		}

		wg.Add(1)
		sectionSemaphore <- struct{}{} // Acquire semaphore

		go func(sectionObj utils.SectionRes) {
			defer wg.Done()
			defer func() { <-sectionSemaphore }() // Release semaphore

			// Create a context with timeout for this request
			_, reqCancel := context.WithTimeout(ctx, RequestTimeout)
			defer reqCancel()

			// Fetch section detail
			sectionDetailRaw, err := utils.GetSectionDetailRaw(year, term, dept, number, sectionObj.Value)
			if err != nil {
				fmt.Printf("Error getting section detail for %s %s %s: %v\n", dept, number, sectionObj.Value, err)
				return
			}

			// Add section detail to the array (thread-safe)
			sectionMu.Lock()
			sectionDetailRawArr = append(sectionDetailRawArr, sectionDetailRaw)
			sectionMu.Unlock()
		}(section)
	}

	// Wait for all sections to be processed
	wg.Wait()

	// Skip if no section details were collected
	if len(sectionDetailRawArr) == 0 {
		fmt.Printf("No section details collected for %s %s\n", dept, number)
		return nil
	}

	// Convert the section details to CourseWithSectionDetails
	maybeCourseWithSectionDetails := utils.ToCourseWithSectionDetails(sectionDetailRawArr)
	if maybeCourseWithSectionDetails.IsAbsent() {
		fmt.Printf("Error converting course with section details - %s %s %s %s\n", year, term, dept, number)
		return nil
	}

	// Get the course with section details
	courseWithSectionDetails := maybeCourseWithSectionDetails.MustGet()

	// Sort the section details by section code before adding to the map
	slices.SortFunc(courseWithSectionDetails.SectionDetails, func(a model.SectionDetail, b model.SectionDetail) int {
		return strings.Compare(a.Section, b.Section)
	})

	// Add the course with section details to the map (thread-safe)
	courseKey := fmt.Sprintf("%s %s", dept, number)
	mu.Lock()
	courseWithSectionDetailsMap[courseKey] = courseWithSectionDetails
	mu.Unlock()

	fmt.Printf("Processed section details for %s %s %s %s (%d sections)\n", dept, number, term, year, len(sections))
	return nil
}
