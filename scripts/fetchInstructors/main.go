package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/model"
	internalUtils "github.com/brianrahadi/sfucourses-api/internal/utils"
	"github.com/samber/lo"
)

const (
	RESULT_PATH = "./internal/store/json/instructors.json"
)

type InstructorOffering struct {
	Dept   string `json:"dept"`
	Number string `json:"number"`
	Term   string `json:"term"`
	Title  string `json:"title"`
}

type Instructor struct {
	Name      string               `json:"name"`
	Offerings []InstructorOffering `json:"offerings"`
}

func main() {
	BASE_PATH := "./internal/store/json"

	outlines, err := internalUtils.ReadCoursesFromJSON[[]model.CourseOutline](BASE_PATH + "/outlines.json")
	if err != nil {
		fmt.Printf("Error reading courses from JSON: %v\n", err)
		return
	}

	instructorMap := make(map[string]*Instructor)

	for _, outline := range outlines {
		if outline.Dept == "" || outline.Number == "" {
			continue
		}

		for _, offering := range outline.Offerings {
			for _, instructorName := range offering.Instructors {
				if strings.TrimSpace(instructorName) == "" {
					continue
				}

				cleanName := strings.TrimSpace(instructorName)

				if _, exists := instructorMap[cleanName]; !exists {
					instructorMap[cleanName] = &Instructor{
						Name:      cleanName,
						Offerings: []InstructorOffering{},
					}
				}

				newOffering := InstructorOffering{
					Dept:   outline.Dept,
					Number: outline.Number,
					Term:   offering.Term,
					Title:  outline.Title,
				}

				instructor := instructorMap[cleanName]
				offeringExists := lo.ContainsBy(instructor.Offerings, func(existing InstructorOffering) bool {
					return existing.Dept == newOffering.Dept &&
						existing.Number == newOffering.Number &&
						existing.Term == newOffering.Term
				})

				if !offeringExists {
					instructor.Offerings = append(instructor.Offerings, newOffering)
				}
			}
		}
	}

	instructors := make([]Instructor, 0, len(instructorMap))
	for _, instructor := range instructorMap {
		slices.SortFunc(instructor.Offerings, func(a, b InstructorOffering) int {
			termA := termToSortableValue(a.Term)
			termB := termToSortableValue(b.Term)
			if termA != termB {
				return termB - termA // Descending order (most recent first)
			}

			if a.Dept != b.Dept {
				return strings.Compare(a.Dept, b.Dept)
			}

			return strings.Compare(a.Number, b.Number)
		})

		instructors = append(instructors, *instructor)
	}

	slices.SortFunc(instructors, func(a, b Instructor) int {
		return strings.Compare(a.Name, b.Name)
	})

	err = writeInstructorsToJSON(instructors, RESULT_PATH)
	if err != nil {
		fmt.Printf("Error writing instructors to JSON: %v\n", err)
		return
	}

	fmt.Printf("Successfully wrote %d instructors to %s\n", len(instructors), RESULT_PATH)

	// Print some statistics
	totalOfferings := 0
	for _, instructor := range instructors {
		totalOfferings += len(instructor.Offerings)
	}
	fmt.Printf("Total offerings: %d\n", totalOfferings)
}

func termToSortableValue(term string) int {
	parts := strings.Split(term, " ")
	if len(parts) != 2 {
		return 0 // Handle invalid format
	}

	season, yearStr := parts[0], parts[1]

	year := 0
	if _, err := fmt.Sscanf(yearStr, "%d", &year); err != nil {
		return 0 // Handle invalid year
	}

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

	return year*100 + month
}

func writeInstructorsToJSON(instructors []Instructor, filePath string) error {
	jsonData, err := json.Marshal(instructors)
	if err != nil {
		return fmt.Errorf("error marshaling instructors to JSON: %w", err)
	}

	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing instructors to file: %w", err)
	}

	return nil
}
