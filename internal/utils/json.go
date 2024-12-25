package utils

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadCoursesFromJSON[T any](filePath string) (T, error) {
	var result T

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return result, fmt.Errorf("file does not exist: %s", filePath)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return result, fmt.Errorf("error reading file: %v", err)
	}

	// Parse JSON into struct
	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("error parsing JSON: %v", err)
	}

	return result, nil
}
