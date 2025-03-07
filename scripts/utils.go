package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/brianrahadi/sfucourses-api/internal/model"
	. "github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/samber/lo"
	mo "github.com/samber/mo"
)

const (
	BaseURL = "http://www.sfu.ca/bin/wcm/course-outlines"
)

type DepartmentRes struct {
	Name  string `json:"name"`  // CMPT
	Value string `json:"value"` // Computing Science
}

type CourseRes struct {
	Title string `json:"title"` // 225
	Value string `json:"value"` // Data Structures and Programming
}

type SectionRes struct {
	Title string `json:"title"` // D100
	Value string `json:"value"` // Data Structures and Programming
}

type OutlineRes struct {
	Info CourseOutline `json:"info"`
}

// fetchAndDecode makes an HTTP GET request and decodes the JSON response
func FetchAndDecode(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("HTTP GET failed for %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP request failed with status %d for %s: %s", resp.StatusCode, url, string(body))
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("JSON decode failed for %s: %w", url, err)
	}

	return nil
}

// GetDepartments fetches all departments for a given term
func GetDepartments(year string, term string) ([]DepartmentRes, error) {
	url := fmt.Sprintf("%s?%s/%s", BaseURL, year, term)
	var depts []DepartmentRes
	err := FetchAndDecode(url, &depts)
	return depts, err
}

// GetCourses fetches all courses for a given department in a term
func GetCourses(year string, term string, dept string) ([]CourseRes, error) {
	url := fmt.Sprintf("%s?%s/%s/%s", BaseURL, year, term, dept)
	var courses []CourseRes
	err := FetchAndDecode(url, &courses)
	return courses, err
}

// GetSections fetches all sections for a given course
func GetSections(year string, term string, dept string, number string) ([]SectionRes, error) {
	url := fmt.Sprintf("%s?%s/%s/%s/%s", BaseURL, year, term, dept, number)
	var sections []SectionRes
	err := FetchAndDecode(url, &sections)
	return sections, err
}

// GetCourseOutline fetches the course outline for a given section
func GetCourseOutline(year string, term string, dept string, number string, section string) (OutlineRes, error) {
	url := fmt.Sprintf("%s?%s/%s/%s/%s/%s", BaseURL, year, term, dept, number, section)
	var outline OutlineRes
	err := FetchAndDecode(url, &outline)
	return outline, err
}

// GetSectionDetailRaw fetches the course outline for a given section
func GetSectionDetailRaw(year string, term string, dept string, number string, section string) (SectionDetailRaw, error) {
	url := fmt.Sprintf("%s?%s/%s/%s/%s/%s", BaseURL, year, term, dept, number, section)
	var sectionDetailRaw SectionDetailRaw
	err := FetchAndDecode(url, &sectionDetailRaw)
	return sectionDetailRaw, err
}

// ProcessTerm handles all the fetching for a single term
func ProcessTerm(year string, term string, courseMap mo.Either[map[string]model.CourseOutline, map[string]model.CourseWithSectionDetails]) error {
	depts, err := GetDepartments(year, term)
	if err != nil {
		return fmt.Errorf("error getting departments for term %s: %w", term, err)
	}

	for _, dept := range depts {
		if err := processDepartment(year, term, dept.Value, courseMap); err != nil {
			fmt.Printf("Error processing department %s: %v\n", dept.Value, err)
			continue
		}
	}

	return nil
}

// processDepartment handles all the fetching for a single department
func processDepartment(year string, term string, dept string, courseMap mo.Either[map[string]model.CourseOutline, map[string]CourseWithSectionDetails]) error {
	courses, err := GetCourses(year, term, dept)
	if err != nil {
		return fmt.Errorf("error getting courses for department %s: %w", dept, err)
	}

	if courseMap.IsLeft() {
		outlineMap := courseMap.LeftOrEmpty()
		for _, course := range courses {
			courseKey := fmt.Sprintf("%s %s", dept, course.Value)
			_, ok := outlineMap[courseKey]

			if ok {
				continue
			}

			if err := processCourseOutline(year, term, dept, course.Value, outlineMap); err != nil {
				fmt.Printf("Error processing course %s: %v\n", course.Value, err)
				continue
			}
		}
		return nil
	}
	sectionDetailsMap := courseMap.RightOrEmpty()
	for _, course := range courses {
		if err := processSectionDetails(year, term, dept, course.Value, sectionDetailsMap); err != nil {
			fmt.Printf("Error processing course %s: %v\n", course.Value, err)
			continue
		}
	}

	return nil
}

// processCourseOutline handles all the fetching for a single course
func processCourseOutline(year string, term string, dept string, number string, outlineMap map[string]model.CourseOutline) error {
	sections, err := GetSections(year, term, dept, number)
	if err != nil {
		return fmt.Errorf("error getting sections for course %s: %w", number, err)
	}

	if len(sections) == 0 {
		fmt.Printf("No outline found for %s %s\n", dept, number)
		return nil
	}

	outline, err := GetCourseOutline(year, term, dept, number, sections[0].Value)
	if err != nil {
		return fmt.Errorf("error getting outline for section %s: %w", sections[0].Value, err)
	}

	courseKey := fmt.Sprintf("%s %s", dept, number)
	outlineMap[courseKey] = outline.Info

	// Process the outline as needed
	fmt.Printf("Processed outline of %s %s\n", dept, number)
	return nil
}

func processSectionDetails(year string, term string, dept string, number string, courseWithSectionDetailsMap map[string]model.CourseWithSectionDetails) error {
	sections, err := GetSections(year, term, dept, number)
	if err != nil {
		return fmt.Errorf("error getting sections for course %s: %w", number, err)
	}

	sectionDetailRawArr := make([]SectionDetailRaw, 0, len(sections))

	for _, section := range sections {
		sectionDetailRaw, err := GetSectionDetailRaw(year, term, dept, number, section.Value)
		if err != nil {
			fmt.Printf("error getting outline for section %s: %v", sections[0].Value, err)
			continue
		}
		sectionDetailRawArr = append(sectionDetailRawArr, sectionDetailRaw)
	}
	maybeCourseWithSectionDetails := ToCourseWithSectionDetails(sectionDetailRawArr)
	if maybeCourseWithSectionDetails.IsAbsent() {
		fmt.Printf("Error converting course with section details - %s %s %s %s", year, term, dept, number)
		return nil
	}
	courseKey := fmt.Sprintf("%s %s", dept, number)
	courseWithSectionDetailsMap[courseKey] = maybeCourseWithSectionDetails.MustGet()

	fmt.Printf("Processed section details for %s %s %s %s\n", dept, number, term, year)
	return nil
}

// ToCourseWithSectionDetails converts an array of SectionDetailRaw to CourseWithSectionDetails
func ToCourseWithSectionDetails(sectionDetailRawArr []SectionDetailRaw) mo.Option[CourseWithSectionDetails] {
	if len(sectionDetailRawArr) == 0 {
		return mo.None[CourseWithSectionDetails]()
	}
	var courseWithSections CourseWithSectionDetails
	courseWithSections.Dept = sectionDetailRawArr[0].Info.Dept
	courseWithSections.Number = sectionDetailRawArr[0].Info.Number
	courseWithSections.Title = sectionDetailRawArr[0].Info.Title
	courseWithSections.Units = sectionDetailRawArr[0].Info.Units
	courseWithSections.Term = sectionDetailRawArr[0].Info.Term

	sectionDetails := lo.Map(sectionDetailRawArr, func(sectionDetailRaw SectionDetailRaw, _ int) SectionDetail {
		instructors := sectionDetailRaw.Instructor
		if instructors == nil {
			instructors = []Instructor{}
		}

		schedules := sectionDetailRaw.CourseSchedule
		if schedules == nil {
			schedules = []SectionSchedule{}
		}
		schedules = lo.Map(schedules, func(schedule SectionSchedule, _ int) SectionSchedule {
			startDate, err := TransformDate(schedule.StartDate)
			if err != nil {
				fmt.Printf("Error transforming date: %v", err)
			}
			endDate, err := TransformDate(schedule.EndDate)
			if err != nil {
				fmt.Printf("Error transforming date: %v", err)
			}
			return SectionSchedule{
				StartDate:   startDate,
				EndDate:     endDate,
				StartTime:   schedule.StartTime,
				EndTime:     schedule.EndTime,
				Days:        schedule.Days,
				Campus:      schedule.Campus,
				SectionCode: schedule.SectionCode,
			}
		})

		return SectionDetail{
			Section:        sectionDetailRaw.Info.Section,
			DeliveryMethod: sectionDetailRaw.Info.DeliveryMethod,
			ClassNumber:    sectionDetailRaw.Info.ClassNumber,
			Instructors:    instructors,
			Schedules:      schedules,
		}
	})
	courseWithSections.SectionDetails = sectionDetails
	return mo.Some(courseWithSections)
}

// ProcessAndWriteOutlines takes a map of course outlines and a destination path,
// processes the data (filters, sorts), and writes it to a JSON file
func ProcessAndWriteOutlines(outlineMap map[string]model.CourseOutline, destPath string) error {
	// Convert map to slice
	outlines := slices.Collect(maps.Values(outlineMap))

	// Remove entries with empty department or number
	outlines = lo.Filter(outlines, func(course model.CourseOutline, _ int) bool {
		return course.Dept != "" && course.Number != ""
	})

	// Sort by department and number
	slices.SortFunc(outlines, func(a model.CourseOutline, b model.CourseOutline) int {
		if a.Dept != b.Dept {
			return strings.Compare(a.Dept, b.Dept)
		}
		return strings.Compare(a.Number, b.Number)
	})

	// Marshal to JSON
	jsonData, err := json.Marshal(outlines)
	if err != nil {
		return fmt.Errorf("error marshaling to JSON: %w", err)
	}

	// Write to file
	if err := os.WriteFile(destPath, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func TransformDate(inputDate string) (string, error) {
	parsedTime, err := time.Parse("Mon Jan 02 15:04:05 MST 2006", inputDate)
	if err != nil {
		return "", err
	}

	return parsedTime.Format("2006-01-02"), nil
}
