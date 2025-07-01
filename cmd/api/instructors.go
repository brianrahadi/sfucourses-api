package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/brianrahadi/sfucourses-api/internal/store"
	"github.com/samber/lo"
)

// @Summary		Get all instructors
// @Description	Retrieves a list of all instructors with their course offerings
// @Tags			Instructors
// @Accept			json
// @Produce		json
// @Success		200	{array}		model.InstructorResponse	"Response for instructors"
// @Failure		404	{object}	ErrorResponse				"No instructors found"
// @Router			/v1/rest/instructors [get]
func (app *application) getAllInstructors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	instructors, err := app.store.Instructors.GetAll(ctx)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, instructors); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary		Get instructors by department
// @Description	Retrieves all instructors who teach courses in a specific department
// @Tags			Instructors
// @Accept			json
// @Produce		json
// @Param			dept	path		string						true	"Department code (e.g., CMPT, MATH)"
// @Success		200		{array}		[]model.InstructorResponse	"List of instructors for the department"
// @Failure		404		{object}	ErrorResponse				"Department not found or no instructors available"
// @Failure		500		{object}	ErrorResponse				"Internal server error"
// @Router			/v1/rest/instructors/{dept} [get]
func (app *application) getInstructorsByDept(w http.ResponseWriter, r *http.Request) {
	dept := r.PathValue("dept")
	ctx := r.Context()

	instructors, err := app.store.Instructors.GetByDept(ctx, dept)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, instructors); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary		Get instructors by department and course number
// @Description	Retrieves all instructors who teach a specific course
// @Tags			Instructors
// @Accept			json
// @Produce		json
// @Param			dept	path		string						true	"Department code (e.g., CMPT, MATH)"
// @Param			number	path		string						true	"Course number (e.g., 120, 225)"
// @Success		200		{array}		[]model.InstructorResponse	"List of instructors for the course"
// @Failure		404		{object}	ErrorResponse				"Course not found or no instructors available"
// @Failure		500		{object}	ErrorResponse				"Internal server error"
// @Router			/v1/rest/instructors/{dept}/{number} [get]
func (app *application) getInstructorsByDeptAndNumber(w http.ResponseWriter, r *http.Request) {
	dept := r.PathValue("dept")
	number := r.PathValue("number")
	ctx := r.Context()

	instructors, err := app.store.Instructors.GetByDeptAndNumber(ctx, dept, number)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, instructors); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

// @Summary		Get instructor by name
// @Description	Retrieves a specific instructor containing their name with all their course offerings
// @Tags			Instructors
// @Accept			json
// @Produce		json
// @Param			name	path		string						true	"Instructor name (URL encoded)"
// @Success		200		{object}	model.InstructorResponse	"Instructor details with offerings"
// @Failure		404		{object}	ErrorResponse				"Instructor not found"
// @Failure		500		{object}	ErrorResponse				"Internal server error"
// @Router			/v1/rest/instructors/{name} [get]
func (app *application) getInstructorsByName(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	ctx := r.Context()

	instructors, err := app.store.Instructors.GetAll(ctx)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	filteredInstructors := lo.Filter(instructors, func(instructor model.InstructorResponse, _ int) bool {
		return strings.Contains(instructor.Name, name)
	})

	if err := writeJSON(w, http.StatusOK, filteredInstructors); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func filterValidInstructors(instructors []model.InstructorResponse) []model.InstructorResponse {
	filtered := make([]model.InstructorResponse, 0, len(instructors))
	for _, inst := range instructors {
		if strings.Contains(inst.Name, "TBA") || strings.Contains(strings.ToLower(inst.Name), "faculty") {
			continue
		}
		filtered = append(filtered, inst)
	}
	return filtered
}

func (app *application) getInstructors(w http.ResponseWriter, r *http.Request) {
	instructors, err := loadInstructors()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to load instructors")
		return
	}
	instructors = filterValidInstructors(instructors)
	writeJSON(w, http.StatusOK, instructors)
}

func (app *application) getInstructorByDept(w http.ResponseWriter, r *http.Request) {
	dept := strings.ToUpper(r.PathValue("dept"))
	instructors, err := loadInstructors()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to load instructors")
		return
	}
	instructors = filterValidInstructors(instructors)
	filtered := make([]model.InstructorResponse, 0)
	for _, inst := range instructors {
		for _, off := range inst.Offerings {
			if strings.ToUpper(off.Dept) == dept {
				filtered = append(filtered, inst)
				break
			}
		}
	}
	writeJSON(w, http.StatusOK, filtered)
}

func (app *application) getInstructorByDeptAndNumber(w http.ResponseWriter, r *http.Request) {
	dept := strings.ToUpper(r.PathValue("dept"))
	number := strings.ToUpper(r.PathValue("number"))
	instructors, err := loadInstructors()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to load instructors")
		return
	}
	instructors = filterValidInstructors(instructors)
	filtered := make([]model.InstructorResponse, 0)
	for _, inst := range instructors {
		for _, off := range inst.Offerings {
			if strings.ToUpper(off.Dept) == dept && strings.ToUpper(off.Number) == number {
				filtered = append(filtered, inst)
				break
			}
		}
	}
	writeJSON(w, http.StatusOK, filtered)
}

func (app *application) getInstructorByName(w http.ResponseWriter, r *http.Request) {
	dept := strings.ToUpper(r.PathValue("dept"))
	name := strings.ToLower(r.PathValue("name"))
	instructors, err := loadInstructors()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to load instructors")
		return
	}
	instructors = filterValidInstructors(instructors)
	filtered := make([]model.InstructorResponse, 0)
	for _, inst := range instructors {
		if strings.ToLower(inst.Name) == name {
			for _, off := range inst.Offerings {
				if strings.ToUpper(off.Dept) == dept {
					filtered = append(filtered, inst)
					break
				}
			}
		}
	}
	writeJSON(w, http.StatusOK, filtered)
}
