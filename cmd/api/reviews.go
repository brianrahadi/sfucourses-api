package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/model"
)

const reviewDataDir = "internal/store/json/instructor_reviews"

// normalizeName converts a name to match the file naming convention
func normalizeName(name string) string {
	// Convert to lowercase and replace spaces with underscores
	return strings.ReplaceAll(strings.ToLower(name), " ", "_")
}

// findInstructorFile searches for an instructor file across all departments
func findInstructorFile(instructorName string) (string, error) {
	// Try different variations of the name
	variations := []string{
		normalizeName(instructorName),
		strings.ReplaceAll(instructorName, " ", "_"),
		strings.ReplaceAll(instructorName, " ", ""),
	}

	// Walk through all department directories
	err := filepath.Walk(reviewDataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(path, ".json") {
			filename := strings.TrimSuffix(filepath.Base(path), ".json")

			for _, variation := range variations {
				if strings.EqualFold(filename, variation) {
					return filepath.SkipDir // Found it, stop walking
				}
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	// Try to find the file by checking each department
	departments, err := os.ReadDir(reviewDataDir)
	if err != nil {
		return "", err
	}

	for _, dept := range departments {
		if dept.IsDir() {
			deptPath := filepath.Join(reviewDataDir, dept.Name())
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

// @Summary		Get all reviews overview
// @Description	Returns overview of all available instructor review data
// @Tags			Reviews
// @Accept			json
// @Produce		json
// @Success		200	{object}	map[string]interface{}	"Overview of available review data"
// @Failure		500	{object}	ErrorResponse			"Internal server error"
// @Router			/v1/rest/reviews [get]
func (app *application) getAllReviews(w http.ResponseWriter, r *http.Request) {
	departments, err := os.ReadDir(reviewDataDir)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	var departmentStats []map[string]interface{}
	totalInstructors := 0

	for _, dept := range departments {
		if dept.IsDir() {
			deptPath := filepath.Join(reviewDataDir, dept.Name())
			files, err := os.ReadDir(deptPath)
			if err != nil {
				continue
			}

			instructorCount := 0
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
					instructorCount++
				}
			}

			departmentStats = append(departmentStats, map[string]interface{}{
				"department":       dept.Name(),
				"instructor_count": instructorCount,
			})
			totalInstructors += instructorCount
		}
	}

	response := map[string]interface{}{
		"message":           "Precomputed instructor review data available",
		"total_instructors": totalInstructors,
		"total_departments": len(departmentStats),
		"departments":       departmentStats,
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
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
// @Description	Retrieves aggregated review data for a specific course code
// @Tags			Reviews
// @Accept			json
// @Produce		json
// @Param			course_code	path		string					false	"Course code (e.g., BUS251, CMPT225)"
// @Success		200			{object}	model.CourseReviewData	"Course review data"
// @Failure		404			{object}	ErrorResponse			"Course not found"
// @Failure		500			{object}	ErrorResponse			"Internal server error"
// @Router			/v1/rest/reviews/courses/{course_code} [get]
func (app *application) getCourseReviews(w http.ResponseWriter, r *http.Request) {
	// Extract course code from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	var courseCode string
	for i, part := range parts {
		if part == "courses" && i+1 < len(parts) {
			courseCode = parts[i+1]
			break
		}
	}

	if courseCode == "" {
		app.badRequestResponse(w, r, errors.New("course code is required"))
		return
	}

	// Search for instructors who taught this course
	var instructors []model.InstructorSummary
	totalReviews := 0

	departments, err := os.ReadDir(reviewDataDir)
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}

	for _, dept := range departments {
		if dept.IsDir() {
			deptPath := filepath.Join(reviewDataDir, dept.Name())
			files, err := os.ReadDir(deptPath)
			if err != nil {
				continue
			}

			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
					filePath := filepath.Join(deptPath, file.Name())
					data, err := os.ReadFile(filePath)
					if err != nil {
						continue
					}

					var reviewData model.InstructorReviewData
					if err := json.Unmarshal(data, &reviewData); err != nil {
						continue
					}

					// Check if this instructor taught the requested course
					courseReviews := 0
					for _, review := range reviewData.Reviews {
						if strings.EqualFold(review.CourseCode, courseCode) {
							courseReviews++
						}
					}

					if courseReviews > 0 {
						// Calculate averages for this instructor's reviews of this course
						var totalRating, totalDifficulty float64
						for _, review := range reviewData.Reviews {
							if strings.EqualFold(review.CourseCode, courseCode) {
								if rating, err := strconv.ParseFloat(review.Rating, 64); err == nil {
									totalRating += rating
								}
								if difficulty, err := strconv.ParseFloat(review.Difficulty, 64); err == nil {
									totalDifficulty += difficulty
								}
							}
						}

						avgRating := totalRating / float64(courseReviews)
						avgDifficulty := totalDifficulty / float64(courseReviews)

						instructor := model.InstructorSummary{
							ProfessorID:    reviewData.ProfessorID,
							ProfessorName:  reviewData.ProfessorName,
							Department:     dept.Name(),
							AvgRating:      avgRating,
							AvgDifficulty:  avgDifficulty,
							ReviewCount:    courseReviews,
							WouldTakeAgain: reviewData.WouldTakeAgain + "%",
						}

						instructors = append(instructors, instructor)
						totalReviews += courseReviews
					}
				}
			}
		}
	}

	if len(instructors) == 0 {
		app.notFoundResponse(w, r, errors.New("course not found"))
		return
	}

	courseData := model.CourseReviewData{
		CourseCode:   strings.ToUpper(courseCode),
		TotalReviews: totalReviews,
		Instructors:  instructors,
	}

	if err := writeJSON(w, http.StatusOK, courseData); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
