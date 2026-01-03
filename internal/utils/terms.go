package utils

import "fmt"

// TermCodes is the single source of truth for all academic term codes.
// Format: "YYYY-season" (e.g., "2024-spring", "2025-fall")
// This list should be updated when new terms are added to the system.
var TermCodes = []string{
	"2024-spring",
	"2024-summer",
	"2024-fall",
	"2025-spring",
	"2025-summer",
	"2025-fall",
	"2026-spring",
	"2026-summer",
}

// GetTermCodes returns all term codes as a slice of strings.
// Example: []string{"2024-spring", "2024-summer", "2024-fall", ...}
func GetTermCodes() []string {
	return TermCodes
}

// GetSectionFilePath returns the file path for a given term code's section JSON file.
// Example: GetSectionFilePath("2024-spring") returns "./internal/store/json/sections/2024-spring.json"
func GetSectionFilePath(termCode string) string {
	return fmt.Sprintf("./internal/store/json/sections/%s.json", termCode)
}

// GetSectionFilePaths returns a map of all term codes to their corresponding section file paths.
//
//	Example: map[string]string{
//	  "2024-spring": "./internal/store/json/sections/2024-spring.json",
//	  "2024-summer": "./internal/store/json/sections/2024-summer.json",
//	  ...
//	}
func GetSectionFilePaths() map[string]string {
	filePaths := make(map[string]string, len(TermCodes))
	for _, termCode := range TermCodes {
		filePaths[termCode] = GetSectionFilePath(termCode)
	}
	return filePaths
}

// GetTermCodesAsYearTerm converts term codes to [year, term] format.
//
//	Example: [][]string{
//	  {"2024", "spring"},
//	  {"2024", "summer"},
//	  {"2025", "fall"},
//	  ...
//	}
//
// Useful for scripts that need year and term as separate values.
func GetTermCodesAsYearTerm() [][]string {
	result := make([][]string, 0, len(TermCodes))
	for _, termCode := range TermCodes {
		parts := SplitTermCode(termCode)
		if len(parts) == 2 {
			result = append(result, []string{parts[0], parts[1]})
		}
	}
	return result
}

// SplitTermCode splits a term code into its year and season components.
// Example: SplitTermCode("2024-spring") returns []string{"2024", "spring"}
// Example: SplitTermCode("2025-fall") returns []string{"2025", "fall"}
func SplitTermCode(termCode string) []string {
	parts := make([]string, 0, 2)
	for i := 0; i < len(termCode); i++ {
		if termCode[i] == '-' {
			parts = append(parts, termCode[:i], termCode[i+1:])
			break
		}
	}
	return parts
}
