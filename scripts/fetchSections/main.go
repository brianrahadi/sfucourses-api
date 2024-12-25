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
	mo "github.com/samber/mo"
)

var ResultFilePath = fmt.Sprintf("./json/schedules/%s-%s.json", termFetched[0], termFetched[1])
var termFetched = []string{"2024", "spring"}

func main() {
	var courseWithSectionDetailsMapContainer = mo.Right[map[string]CourseInfo](make(map[string]CourseWithSectionDetails))

	if err := utils.ProcessTerm(termFetched[0], termFetched[1], courseWithSectionDetailsMapContainer); err != nil {
		fmt.Printf("Error processing term %s %s: %v\n", termFetched[0], termFetched[1], err)
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
		return
	}

	err = os.WriteFile(ResultFilePath, jsonData, 0644)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		return
	}

	fmt.Printf("Successfully wrote course data to %s\n", ResultFilePath)
}
