package main

import (
	"encoding/json"
	"fmt"
	"os"

	model "github.com/brianrahadi/sfucourses-api/internal/model"
	utils "github.com/brianrahadi/sfucourses-api/scripts"
	"github.com/samber/mo"
)

const (
	ResultFilePath = "./json/outlines/outline2.json"
)

func main() {
	terms := [][]string{
		{"2025", "spring"}, {"2024", "fall"}, {"2024", "summer"},
		{"2024", "spring"}, {"2023", "fall"}, {"2023", "summer"},
		{"2023", "spring"},
	}
	var outlineMap = mo.Left[map[string]model.CourseInfo, map[string][]model.SectionDetail](make(map[string]model.CourseInfo))

	for _, term := range terms {
		if err := utils.ProcessTerm(term[0], term[1], outlineMap); err != nil {
			fmt.Printf("Error processing term %s: %v\n", term, err)
			continue
		}
	}

	jsonData, err := json.Marshal(outlineMap.LeftOrEmpty())
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
