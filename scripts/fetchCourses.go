package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	. "github.com/brianrahadi/sfucourses-api/internal/model"
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

const (
	BaseURL        = "http://www.sfu.ca/bin/wcm/course-outlines"
	ResultFilePath = "./json/outlines/outline.json"
)

var outlineMap map[string]CourseInfo = make(map[string]CourseInfo)

// fetchAndDecode makes an HTTP GET request and decodes the JSON response
func fetchAndDecode(url string, target interface{}) error {
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
	err := fetchAndDecode(url, &depts)
	return depts, err
}

// getCourses fetches all courses for a given department in a term
func getCourses(year string, term string, dept string) ([]CourseRes, error) {
	url := fmt.Sprintf("%s?%s/%s/%s", BaseURL, year, term, dept)
	var courses []CourseRes
	err := fetchAndDecode(url, &courses)
	return courses, err
}

// getSections fetches all sections for a given course
func getSections(year string, term string, dept string, number string) ([]SectionRes, error) {
	url := fmt.Sprintf("%s?%s/%s/%s/%s", BaseURL, year, term, dept, number)
	var sections []SectionRes
	err := fetchAndDecode(url, &sections)
	return sections, err
}

// getCourseOutline fetches the course outline for a given section
func getCourseOutline(year string, term string, dept string, number string, section string) (OutlineRes, error) {
	url := fmt.Sprintf("%s?%s/%s/%s/%s/%s", BaseURL, year, term, dept, number, section)
	var outline OutlineRes
	err := fetchAndDecode(url, &outline)
	return outline, err
}

// processTerm handles all the fetching for a single term
func processTerm(year string, term string) error {
	depts, err := getDepartments(year, term)
	if err != nil {
		return fmt.Errorf("error getting departments for term %s: %w", term, err)
	}

	for _, dept := range depts {
		if err := processDepartment(year, term, dept.Value); err != nil {
			fmt.Printf("Error processing department %s: %v\n", dept.Value, err)
			continue
		}
	}

	return nil
}

// processDepartment handles all the fetching for a single department
func processDepartment(year string, term string, dept string) error {
	courses, err := getCourses(year, term, dept)
	if err != nil {
		return fmt.Errorf("error getting courses for department %s: %w", dept, err)
	}

	for _, course := range courses {
		courseKey := fmt.Sprintf("%s %s", dept, course.Value)
		courseInfo, ok := outlineMap[courseKey]
		courseInfo.TermsOffered = append(outlineMap[courseKey].TermsOffered, fmt.Sprintf("%s %s", term, year))
		outlineMap[courseKey] = courseInfo
		if ok {
			continue
		}
		if err := processCourse(year, term, dept, course.Value); err != nil {
			fmt.Printf("Error processing course %s: %v\n", course.Value, err)
			continue
		}
	}

	return nil
}

// processCourse handles all the fetching for a single course
func processCourse(year string, term string, dept string, number string) error {
	sections, err := getSections(year, term, dept, number)
	if err != nil {
		return fmt.Errorf("error getting sections for course %s: %w", number, err)
	}

	if len(sections) == 0 {
		return nil
	}

	outline, err := getCourseOutline(year, term, dept, number, sections[0].Value)
	if err != nil {
		return fmt.Errorf("error getting outline for section %s: %w", sections[0].Value, err)
	}

	courseKey := fmt.Sprintf("%s %s", dept, number)
	outlineMap[courseKey] = outline.Info

	// Process the outline as needed
	fmt.Printf("Processed outline of %s %s\n", dept, number)
	return nil
}

func main() {
	terms := [][]string{
		{"2025", "spring"}, {"2024", "fall"}, {"2024", "summer"},
		{"2024", "spring"}, {"2023", "fall"}, {"2023", "summer"},
		{"2023", "spring"},
	}

	for _, term := range terms {
		if err := processTerm(term[0], term[1]); err != nil {
			fmt.Printf("Error processing term %s: %v\n", term, err)
			continue
		}
	}

	jsonData, err := json.Marshal(outlineMap)
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v\n", err)
		return
	}

	// Write to file
	err = os.WriteFile(ResultFilePath, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}

	fmt.Printf("Successfully wrote course data to %s\n", ResultFilePath)
}
