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
func getDepartments(term string) ([]DepartmentRes, error) {
	url := fmt.Sprintf("%s?%s", BaseURL, term)
	var depts []DepartmentRes
	err := fetchAndDecode(url, &depts)
	return depts, err
}

// getCourses fetches all courses for a given department in a term
func getCourses(termURL, deptValue string) ([]CourseRes, error) {
	url := fmt.Sprintf("%s/%s", termURL, deptValue)
	var courses []CourseRes
	err := fetchAndDecode(url, &courses)
	return courses, err
}

// getSections fetches all sections for a given course
func getSections(courseURL string) ([]SectionRes, error) {
	var sections []SectionRes
	err := fetchAndDecode(courseURL, &sections)
	return sections, err
}

// getCourseOutline fetches the course outline for a given section
func getCourseOutline(outlineURL string) (OutlineRes, error) {
	var outline OutlineRes
	err := fetchAndDecode(outlineURL, &outline)
	return outline, err
}

// processTerm handles all the fetching for a single term
func processTerm(term string) error {
	depts, err := getDepartments(term)
	if err != nil {
		return fmt.Errorf("error getting departments for term %s: %w", term, err)
	}

	termURL := fmt.Sprintf("%s?%s", BaseURL, term)
	for _, dept := range depts {
		if err := processDepartment(termURL, dept); err != nil {
			fmt.Printf("Error processing department %s: %v\n", dept.Value, err)
			continue
		}
	}

	return nil
}

// processDepartment handles all the fetching for a single department
func processDepartment(termURL string, dept DepartmentRes) error {
	courses, err := getCourses(termURL, dept.Value)
	if err != nil {
		return fmt.Errorf("error getting courses for department %s: %w", dept.Value, err)
	}

	for _, course := range courses {
		courseKey := fmt.Sprintf("%s%s", dept.Value, course.Value)
		_, ok := outlineMap[courseKey]
		if ok {
			continue
		}
		if err := processCourse(termURL, dept.Value, course); err != nil {
			fmt.Printf("Error processing course %s: %v\n", course.Value, err)
			continue
		}
	}

	return nil
}

// processCourse handles all the fetching for a single course
func processCourse(termUrl string, dept string, course CourseRes) error {
	courseURL := fmt.Sprintf("%s/%s/%s", termUrl, dept, course.Value)
	sections, err := getSections(courseURL)
	if err != nil {
		return fmt.Errorf("error getting sections for course %s: %w", course.Value, err)
	}

	if len(sections) == 0 {
		return nil
	}

	outlineURL := fmt.Sprintf("%s/%s", courseURL, sections[0].Value)
	outline, err := getCourseOutline(outlineURL)
	if err != nil {
		return fmt.Errorf("error getting outline for section %s: %w", sections[0].Value, err)
	}

	courseKey := fmt.Sprintf("%s%s", dept, course.Value)
	outlineMap[courseKey] = outline.Info

	// Process the outline as needed
	fmt.Printf("Processed outline of %s %s\n", dept, course.Value)
	return nil
}

func main() {
	terms := []string{
		"2023/spring", "2023/summer", "2023/fall",
		"2024/spring", "2024/summer", "2024/fall",
		"2025/spring",
	}

	for _, term := range terms {
		if err := processTerm(term); err != nil {
			fmt.Printf("Error processing term %s: %v\n", term, err)
			continue
		}
	}

	jsonData, err := json.MarshalIndent(outlineMap, "", "    ")
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
