package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func ReadCoursesFromJSON[T any](filePath string) (T, error) {
	var result T

	if strings.HasPrefix(filePath, "https://") {
		return ReadFromURL[T](filePath)
	}

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

func ReadFromURL[T any](url string) (T, error) {
	var result T

	// Make HTTP request
	resp, err := http.Get(url)
	if err != nil {
		return result, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, fmt.Errorf("error reading response: %v", err)
	}

	// Parse JSON into struct
	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("error parsing JSON: %v", err)
	}

	return result, nil
}
