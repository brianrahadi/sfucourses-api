package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"unicode"
	"unicode/utf8"

	"github.com/brianrahadi/sfucourses-api/internal/model"
	. "github.com/brianrahadi/sfucourses-api/internal/model"
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
	Info CourseInfo `json:"info"`
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

// getDepartments fetches all departments for a given term
func getDepartments(year string, term string) ([]DepartmentRes, error) {
	url := fmt.Sprintf("%s?%s/%s", BaseURL, year, term)
	var depts []DepartmentRes
	err := FetchAndDecode(url, &depts)
	return depts, err
}

// getCourses fetches all courses for a given department in a term
func getCourses(year string, term string, dept string) ([]CourseRes, error) {
	url := fmt.Sprintf("%s?%s/%s/%s", BaseURL, year, term, dept)
	var courses []CourseRes
	err := FetchAndDecode(url, &courses)
	return courses, err
}

// getSections fetches all sections for a given course
func getSections(year string, term string, dept string, number string) ([]SectionRes, error) {
	url := fmt.Sprintf("%s?%s/%s/%s/%s", BaseURL, year, term, dept, number)
	var sections []SectionRes
	err := FetchAndDecode(url, &sections)
	return sections, err
}

// getCourseOutline fetches the course outline for a given section
func getCourseOutline(year string, term string, dept string, number string, section string) (OutlineRes, error) {
	url := fmt.Sprintf("%s?%s/%s/%s/%s/%s", BaseURL, year, term, dept, number, section)
	var outline OutlineRes
	err := FetchAndDecode(url, &outline)
	return outline, err
}

// getSectionDetail fetches the course outline for a given section
func getSectionDetail(year string, term string, dept string, number string, section string) (SectionDetail, error) {
	url := fmt.Sprintf("%s?%s/%s/%s/%s/%s", BaseURL, year, term, dept, number, section)
	var sectionDetailRaw SectionDetailRaw
	err := FetchAndDecode(url, &sectionDetailRaw)
	sectionDetail := toSectionDetail(sectionDetailRaw)
	return sectionDetail, err
}

// ProcessTerm handles all the fetching for a single term
func ProcessTerm(year string, term string, courseMap mo.Either[map[string]model.CourseInfo, map[string][]model.SectionDetail]) error {
	depts, err := getDepartments(year, term)
	if err != nil {
		return fmt.Errorf("error getting departments for term %s: %w", term, err)
	}

	for _, dept := range depts {
		if err := processDepartment(year, term, dept.Value, courseMap); err != nil {
			fmt.Printf("Error processing department %s: %v\n", dept.Value, err)
			continue
		}
		break
	}

	return nil
}

// processDepartment handles all the fetching for a single department
func processDepartment(year string, term string, dept string, courseMap mo.Either[map[string]model.CourseInfo, map[string][]model.SectionDetail]) error {
	courses, err := getCourses(year, term, dept)
	if err != nil {
		return fmt.Errorf("error getting courses for department %s: %w", dept, err)
	}

	if courseMap.IsLeft() {
		outlineMap := courseMap.LeftOrEmpty()
		for _, course := range courses {
			courseKey := fmt.Sprintf("%s %s", dept, course.Value)
			courseInfo, ok := outlineMap[courseKey]

			r, size := utf8.DecodeRuneInString(term)
			capitalizedTerm := string(unicode.ToUpper(r)) + term[size:]

			courseInfo.Terms = append(outlineMap[courseKey].Terms, fmt.Sprintf("%s %s", capitalizedTerm, year))
			outlineMap[courseKey] = courseInfo

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
		break
	}

	return nil
}

// processCourseOutline handles all the fetching for a single course
func processCourseOutline(year string, term string, dept string, number string, outlineMap map[string]model.CourseInfo) error {
	sections, err := getSections(year, term, dept, number)
	if err != nil {
		return fmt.Errorf("error getting sections for course %s: %w", number, err)
	}

	if len(sections) == 0 {
		fmt.Printf("No outline found for %s %s\n", dept, number)
		return nil
	}

	outline, err := getCourseOutline(year, term, dept, number, sections[0].Value)
	if err != nil {
		return fmt.Errorf("error getting outline for section %s: %w", sections[0].Value, err)
	}

	courseKey := fmt.Sprintf("%s %s", dept, number)
	outline.Info.Terms = outlineMap[courseKey].Terms
	outlineMap[courseKey] = outline.Info

	// Process the outline as needed
	fmt.Printf("Processed outline of %s %s\n", dept, number)
	return nil
}

func processSectionDetails(year string, term string, dept string, number string, sectionDetailsMap map[string][]model.SectionDetail) error {
	sections, err := getSections(year, term, dept, number)
	if err != nil {
		return fmt.Errorf("error getting sections for course %s: %w", number, err)
	}

	for _, section := range sections {
		sectionDetail, err := getSectionDetail(year, term, dept, number, section.Value)
		if err != nil {
			fmt.Printf("error getting outline for section %s: %v", sections[0].Value, err)
			continue
		}
		courseKey := fmt.Sprintf("%s %s", dept, number)
		sectionDetailsMap[courseKey] = append(sectionDetailsMap[courseKey], sectionDetail)
	}

	fmt.Printf("Processed section details for %s %s %s %s\n", dept, number, term, year)
	return nil
}

func toSectionDetail(sectionDetailRaw SectionDetailRaw) SectionDetail {
	var sectionDetail SectionDetail
	sectionDetail.CourseSchedules = sectionDetailRaw.CourseSchedules
	sectionDetail.Instructors = sectionDetailRaw.Instructors
	sectionDetail.Dept = sectionDetailRaw.Info.Dept
	sectionDetail.Number = sectionDetailRaw.Info.Number
	sectionDetail.Section = sectionDetailRaw.Info.Section
	sectionDetail.Term = sectionDetailRaw.Info.Term
	sectionDetail.OutlinePath = sectionDetailRaw.Info.OutlinePath
	sectionDetail.DeliveryMethod = sectionDetailRaw.Info.DeliveryMethod
	sectionDetail.ClassNumber = sectionDetailRaw.Info.ClassNumber
	return sectionDetail
}
