package main

import (
	"errors"
	"net/http"

	"github.com/brianrahadi/sfucourses-api/internal/store"
)

// @Summary		Get instructors
// @Description	Retrieves instructors with optional filtering by department, course number, or name
// @Tags			Instructors
// @Accept			json
// @Produce		json
// @Param			dept	query		string						false	"Department code (e.g., cmpt, math)"
// @Param			number	query		string						false	"Course number (e.g., 120, 225)"
// @Param			name	query		string						false	"Instructor name (URL encoded)"
// @Success		200		{array}		model.InstructorResponse	"List of instructors"
// @Failure		404		{object}	ErrorResponse				"No instructors found"
// @Failure		500		{object}	ErrorResponse				"Internal server error"
// @Router			/v1/rest/instructors [get]
func (app *application) getInstructors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dept := r.URL.Query().Get("dept")
	number := r.URL.Query().Get("number")
	name := r.URL.Query().Get("name")

	instructors, err := app.store.Instructors.Get(ctx, dept, number, name)
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
