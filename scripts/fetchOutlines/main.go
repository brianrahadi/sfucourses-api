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
	ResultFilePath = "./json/outlines/outline2.json"
)

func main() {
	terms := [][]string{
		{"2025", "spring"}, {"2024", "fall"}, {"2024", "summer"},
		{"2024", "spring"},
	}
	var outlineMapContainer = mo.Left[map[string]model.CourseInfo, map[string][]model.SectionDetail](make(map[string]model.CourseInfo))

	for _, term := range terms {
		if err := utils.ProcessTerm(term[0], term[1], outlineMapContainer); err != nil {
			fmt.Printf("Error processing term %s: %v\n", term, err)
			continue
		}
	}

	outlineMap := outlineMapContainer.LeftOrEmpty()
	outlineVals := slices.Collect(maps.Values(outlineMap))

	// remove bad data
	outlineVals = lo.Filter(outlineVals, func(course model.CourseInfo, _ int) bool {
		return course.Dept != "" && course.Number != ""
	})

	// sort by department and number
	slices.SortFunc(outlineVals, func(a model.CourseInfo, b model.CourseInfo) int {
		if a.Dept != b.Dept {
			return strings.Compare(a.Dept, b.Dept)
		}
		return strings.Compare(a.Number, b.Number)
	})

	jsonData, err := json.Marshal(outlineVals)
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
