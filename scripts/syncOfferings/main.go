package main

import (
	_ "embed"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/model"
	internalUtils "github.com/brianrahadi/sfucourses-api/internal/utils"
	utils "github.com/brianrahadi/sfucourses-api/scripts"
	"github.com/samber/lo"
)

func termToSortableValue(term string) int {
	// Split the term into season and year
	parts := strings.Split(term, " ")
	if len(parts) != 2 {
		return 0 // Handle invalid format
	}

	season, yearStr := parts[0], parts[1]
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		return 0 // Handle invalid year
	}

	// Assign a month value based on the season
	var month int
	switch season {
	case "Spring":
		month = 1
	case "Summer":
		month = 5
	case "Fall":
		month = 9
	default:
		month = 0 // Handle unknown season
	}

	// Combine year and month into a sortable value
	return year*100 + month
}

// 2025-spring to Spring 2025
func formatTermCode(termCode string) string {
	// Split the term code into year and season
	parts := strings.Split(termCode, "-")
	if len(parts) != 2 {
		return termCode // Return original if format is unexpected
	}

	year := parts[0]
	season := lo.Capitalize(parts[1]) // Capitalize first letter of season
	return season + " " + year
}

func main() {
	BASE_PATH := "./internal/store/json"
	RESULT_PATH := BASE_PATH + "/outlines.json"

	outlines, err := internalUtils.ReadCoursesFromJSON[[]model.CourseOutline](BASE_PATH + "/outlines.json")
	if err != nil {
		fmt.Errorf("Error reading courses from JSON, %s", err.Error())
		return
	}
	for i := range outlines {
		outlines[i].Offerings = []model.CourseOffering{} // reset offerings
	}

	outlineMap := make(map[string]model.CourseOutline, len(outlines))

	for _, outline := range outlines {
		outlineMap[fmt.Sprintf("%s-%s", outline.Dept, outline.Number)] = outline
	}

	termCodes := []string{"2024-spring", "2024-summer", "2024-fall", "2025-spring", "2025-summer"}
	coursesMap := map[string][]model.CourseWithSectionDetails{}

	for _, term := range termCodes {
		courses, err := internalUtils.ReadCoursesFromJSON[[]model.CourseWithSectionDetails](BASE_PATH + fmt.Sprintf("/sections/%s.json", term))
		if err != nil {
			fmt.Errorf("Error reading schedules from JSON %s", term)
		}
		coursesMap[term] = courses
	}

	for term, courses := range coursesMap {
		for _, course := range courses {
			newOffering := model.CourseOffering{Instructors: []string{}, Term: formatTermCode(term)}
			instructorNames := []string{}
			for _, sectionDetail := range course.SectionDetails {
				newInstructorNames := lo.Map(sectionDetail.Instructors, func(instructor model.Instructor, _ int) string { return instructor.Name })
				instructorNames = append(instructorNames, newInstructorNames...)
			}
			instructorNames = lo.Uniq(instructorNames)
			instructorNames = lo.Filter(instructorNames, func(name string, _ int) bool { return name != "" })
			newOffering.Instructors = instructorNames
			outlineKey := fmt.Sprintf("%s-%s", course.Dept, course.Number)
			outline := outlineMap[outlineKey]
			outline.Offerings = append(outline.Offerings, newOffering)
			sort.Slice(outline.Offerings, func(i, j int) bool {
				termI := termToSortableValue(outline.Offerings[i].Term)
				termJ := termToSortableValue(outline.Offerings[j].Term)
				return termI > termJ // Sort in descending order (most recent first)
			})
			outlineMap[outlineKey] = outline
		}
	}

	utils.ProcessAndWriteOutlines(outlineMap, RESULT_PATH)
}
