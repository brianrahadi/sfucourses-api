package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/model"
)

const instructorReviewDataDir = "internal/store/json/instructor_reviews"
const courseReviewDataDir = "internal/store/json/course_reviews"
const reviewsSummaryFile = "internal/store/json/reviews.json"

// normalizeName converts a name to match the file naming convention
func normalizeName(name string) string {
	// Convert to lowercase and replace spaces with underscores
	return strings.ReplaceAll(strings.ToLower(name), " ", "_")
}

// normalizeCourseCode converts a course code to match the file naming convention
func normalizeCourseCode(courseCode string) string {
	// Convert to uppercase and remove spaces
	return strings.ReplaceAll(strings.ToUpper(courseCode), " ", "")
}

// findInstructorFile searches for an instructor file across all departments
func findInstructorFile(instructorName string) (string, error) {
	// Try different variations of the name
	variations := []string{
		normalizeName(instructorName),
		strings.ReplaceAll(instructorName, " ", "_"),
		strings.ReplaceAll(instructorName, " ", ""),
	}

	// Try to find the file by checking each department
	departments, err := os.ReadDir(instructorReviewDataDir)
	if err != nil {
		return "", err
	}

	for _, dept := range departments {
		if dept.IsDir() {
			deptPath := filepath.Join(instructorReviewDataDir, dept.Name())
			files, err := os.ReadDir(deptPath)
			if err != nil {
				continue
			}

			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
					filename := strings.TrimSuffix(file.Name(), ".json")

					for _, variation := range variations {
						if strings.EqualFold(filename, variation) {
							return filepath.Join(deptPath, file.Name()), nil
						}
					}
				}
			}
		}
	}

	return "", errors.New("instructor not found")
}

// findCourseFile searches for a course file across all departments
func findCourseFile(courseCode string) (string, error) {
	// Try different variations of the course code
	variations := []string{
		normalizeCourseCode(courseCode),
		strings.ToUpper(courseCode),
		strings.ToLower(courseCode),
	}

	// Try to find the file by checking each department
	departments, err := os.ReadDir(courseReviewDataDir)
	if err != nil {
		return "", err
	}

	for _, dept := range departments {
		if dept.IsDir() {
			deptPath := filepath.Join(courseReviewDataDir, dept.Name())
			files, err := os.ReadDir(deptPath)
			if err != nil {
				continue
			}

			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
					filename := strings.TrimSuffix(file.Name(), ".json")

					for _, variation := range variations {
						if strings.EqualFold(filename, variation) {
							return filepath.Join(deptPath, file.Name()), nil
						}
					}
				}
			}
		}
	}

	return "", errors.New("course not found")
}

// findAllCourseFiles searches for all course JSON files matching the course code across all department directories.
// It returns all matching files to allow for data aggregation.
func findAllCourseFiles(courseCode string) ([]string, error) {
	// Try different variations of the course code
	variations := []string{
		normalizeCourseCode(courseCode),
		strings.ToUpper(courseCode),
		strings.ToLower(courseCode),
	}

	var matchingFiles []string

	// Try to find all matching files by checking each department
	departments, err := os.ReadDir(courseReviewDataDir)
	if err != nil {
		return nil, err
	}

	for _, dept := range departments {
		if dept.IsDir() {
			deptPath := filepath.Join(courseReviewDataDir, dept.Name())
			files, err := os.ReadDir(deptPath)
			if err != nil {
				continue
			}

			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
					filename := strings.TrimSuffix(file.Name(), ".json")

					for _, variation := range variations {
						if strings.EqualFold(filename, variation) {
							matchingFiles = append(matchingFiles, filepath.Join(deptPath, file.Name()))
							break // Found a match for this file, no need to check other variations
						}
					}
				}
			}
		}
	}

	return matchingFiles, nil
}

// @Summary		Get all reviews overview
// @Description	Returns summary review data for all professors from reviews.json
// @Tags			Reviews
// @Accept			json
// @Produce		json
// @Success		200	{array}		model.ProfessorSummary	"List of professor summaries"
// @Failure		500	{object}	ErrorResponse			"Internal server error"
// @Router			/v1/rest/reviews [get]
func (app *application) getAllReviews(w http.ResponseWriter, r *http.Request) {
	// Read and parse the reviews.json file
	data, err := os.ReadFile(reviewsSummaryFile)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var summaries []model.ProfessorSummary
	if err := json.Unmarshal(data, &summaries); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, summaries); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary		Get instructor reviews
// @Description	Retrieves detailed review data for a specific instructor by name
// @Tags			Reviews
// @Accept			json
// @Produce		json
// @Param			instructor_name	path		string						false	"Instructor name (e.g., Angela_Lin, Angela Lin)"
// @Success		200				{object}	model.InstructorReviewData	"Instructor review data"
// @Failure		404				{object}	ErrorResponse				"Instructor not found"
// @Failure		500				{object}	ErrorResponse				"Internal server error"
// @Router			/v1/rest/reviews/instructors/{instructor_name} [get]
func (app *application) getInstructorReviews(w http.ResponseWriter, r *http.Request) {
	// Extract instructor name from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var instructorName string
	for i, part := range parts {
		if part == "instructors" && i+1 < len(parts) {
			instructorName = parts[i+1]
			break
		}
	}

	if instructorName == "" {
		app.badRequestResponse(w, r, errors.New("instructor name is required"))
		return
	}

	// Find the instructor file
	filePath, err := findInstructorFile(instructorName)
	if err != nil {
		app.notFoundResponse(w, r, errors.New("instructor not found"))
		return
	}

	// Read and parse the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var reviewData model.InstructorReviewData
	if err := json.Unmarshal(data, &reviewData); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	if err := writeJSON(w, http.StatusOK, reviewData); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary		Get course reviews
// @Description	Retrieves precomputed review data for a specific course code
// @Tags			Reviews
// @Accept			json
// @Produce		json
// @Param			course_code	path		string					false	"Course code (e.g., CMPT353, BUS251)"
// @Success		200			{object}	model.CourseReviewData	"Course review data"
// @Failure		404			{object}	ErrorResponse			"Course not found"
// @Failure		500			{object}	ErrorResponse			"Internal server error"
// @Router			/v1/rest/reviews/courses/{course_code} [get]
func (app *application) getCourseReviews(w http.ResponseWriter, r *http.Request) {
	courseCode := r.PathValue("course_code")

	if courseCode == "" {
		app.badRequestResponse(w, r, errors.New("course code is required"))
		return
	}

	// Find all matching course files
	filePaths, err := findAllCourseFiles(courseCode)
	if err != nil || len(filePaths) == 0 {
		app.notFoundResponse(w, r, errors.New("course not found"))
		return
	}

	// Aggregate data from all matching files
	var aggregatedData model.CourseReviewData
	var allInstructors []model.InstructorSummary
	totalReviews := 0

	for _, filePath := range filePaths {
		data, err := os.ReadFile(filePath)
		if err != nil {
			continue // Skip files that can't be read
		}

		var courseData model.CourseReviewData
		if err := json.Unmarshal(data, &courseData); err != nil {
			continue // Skip files that can't be parsed
		}

		// Set course code from first file
		if aggregatedData.CourseCode == "" {
			aggregatedData.CourseCode = courseData.CourseCode
		}

		// Aggregate instructors and reviews
		allInstructors = append(allInstructors, courseData.Instructors...)
		totalReviews += courseData.TotalReviews
	}

	// Set aggregated data
	aggregatedData.Instructors = allInstructors
	aggregatedData.TotalReviews = totalReviews

	if err := writeJSON(w, http.StatusOK, aggregatedData); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
