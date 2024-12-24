package main

import (
	"encoding/json"
	"fmt"
	"os"

	. "github.com/brianrahadi/sfucourses-api/internal/model"
	utils "github.com/brianrahadi/sfucourses-api/scripts"
	mo "github.com/samber/mo"
)

var ResultFilePath = fmt.Sprintf("../json/schedules/%s-%s.json", termFetched[0], termFetched[1])
var termFetched = []string{"2023", "spring"}

func main() {
	var sectionDetailMap = mo.Right[map[string]CourseInfo, map[string][]SectionDetail](make(map[string][]SectionDetail))

	if err := utils.ProcessTerm(termFetched[0], termFetched[1], sectionDetailMap); err != nil {
		fmt.Printf("Error processing term %s %s: %v\n", termFetched[0], termFetched[1], err)
	}

	jsonData, err := json.Marshal(sectionDetailMap)
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
