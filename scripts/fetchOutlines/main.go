package main

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	model "github.com/brianrahadi/sfucourses-api/internal/model"
	utils "github.com/brianrahadi/sfucourses-api/scripts"
	"github.com/samber/lo"
	"github.com/samber/mo"
)

const (
	ResultFilePath = "./json/outlines/outline.json"
)

func main() {
	terms := [][]string{
		{"2025", "spring"}, {"2024", "fall"}, {"2024", "summer"},
		{"2024", "spring"},
	}
	var outlineMapContainer = mo.Left[map[string]model.CourseOutline, map[string]model.CourseWithSectionDetails](make(map[string]model.CourseOutline))

	for _, term := range terms {
		if err := utils.ProcessTerm(term[0], term[1], outlineMapContainer); err != nil {
			fmt.Printf("Error processing term %s: %v\n", term, err)
			continue
		}
	}

	outlineMap := outlineMapContainer.LeftOrEmpty()
	outlines := slices.Collect(maps.Values(outlineMap))

	// remove bad data
	outlines = lo.Filter(outlines, func(course model.CourseOutline, _ int) bool {
		return course.Dept != "" && course.Number != ""
	})

	// sort by department and number
	slices.SortFunc(outlines, func(a model.CourseOutline, b model.CourseOutline) int {
		if a.Dept != b.Dept {
			return strings.Compare(a.Dept, b.Dept)
		}
		return strings.Compare(a.Number, b.Number)
	})

	jsonData, err := json.Marshal(outlines)
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
