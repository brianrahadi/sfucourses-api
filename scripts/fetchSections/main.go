package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	. "github.com/brianrahadi/sfucourses-api/internal/model"
	utils "github.com/brianrahadi/sfucourses-api/scripts"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: program <year> <term>")
		fmt.Println("Example: program 2025 spring")
		os.Exit(1)
	}

	year := os.Args[1]
	term := os.Args[2]

	resultFilePath := fmt.Sprintf("./internal/store/json/schedules/%s-%s.json", year, term)

	var courseWithSectionDetailsMapContainer = mo.Right[map[string]CourseOutline](make(map[string]CourseWithSectionDetails))

	if err := utils.ProcessTerm(year, term, courseWithSectionDetailsMapContainer); err != nil {
		fmt.Printf("Error processing term %s %s: %v\n", year, term, err)
		os.Exit(1)
	}

	courseWithSectionDetailsMap := courseWithSectionDetailsMapContainer.RightOrEmpty()
	courseWithSectionDetails := slices.Collect(maps.Values(courseWithSectionDetailsMap))

	// remove bad data
	courseWithSectionDetails = lo.Filter(courseWithSectionDetails, func(courseWithSectionDetails CourseWithSectionDetails, _ int) bool {
		return courseWithSectionDetails.Dept != "" && courseWithSectionDetails.Number != ""
	})

	// sort by department and number
	slices.SortFunc(courseWithSectionDetails, func(a CourseWithSectionDetails, b CourseWithSectionDetails) int {
		if a.Dept != b.Dept {
			return strings.Compare(a.Dept, b.Dept)
		}
		return strings.Compare(a.Number, b.Number)
	})

	jsonData, err := json.Marshal(courseWithSectionDetails)
	if err != nil {
		fmt.Printf("Error marshaling to JSON: %v\n", err)
		os.Exit(1)
	}

	err = os.WriteFile(resultFilePath, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully wrote course data to %s\n", resultFilePath)
}
