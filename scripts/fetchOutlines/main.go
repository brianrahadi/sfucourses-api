package main

import (
	"fmt"

	model "github.com/brianrahadi/sfucourses-api/internal/model"
	utils "github.com/brianrahadi/sfucourses-api/scripts"
	"github.com/samber/mo"
)

const (
	RESULT_PATH = "./internal/store/json/outlines.json"
)

func main() {
	terms := [][]string{
		{"2026", "spring"}, {"2025", "fall"}, {"2025", "summer"}, {"2025", "spring"}, {"2024", "fall"}, {"2024", "summer"},
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
	utils.ProcessAndWriteOutlines(outlineMap, RESULT_PATH)
}
