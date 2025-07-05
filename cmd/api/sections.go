package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/store"
)

// Helper function to split term parameter into year and term components
func splitYearTerm(yearTerm string) (year, term string, err error) {
	parts := strings.Split(yearTerm, "-")
	if len(parts) != 2 {
		return "", "", errors.New("invalid year-term format")
	}
	return parts[0], parts[1], nil
}

// @Summary		Get sections
// @Description	Retrieves course sections for a specific year and term, optionally filtered by department and/or course number
// @Tags			Sections
// @Accept			json
// @Produce		json
// @Param			term	query		string								true	"Year and term in format YYYY-Term (e.g., 2024-spring)"
// @Param			dept	query		string								false	"Department code (e.g., cmpt, math)"
// @Param			number	query		string								false	"Course number (e.g., 120, 225)"
// @Success		200		{array}		[]model.CourseWithSectionDetails	"List of sections"
// @Failure		400		{object}	ErrorResponse						"Invalid year-term format or query parameters"
// @Failure		404		{object}	ErrorResponse						"No sections found for the specified criteria"
// @Failure		500		{object}	ErrorResponse						"Internal server error"
// @Router			/v1/rest/sections [get]
func (app *application) getSections(w http.ResponseWriter, r *http.Request) {
	term := r.URL.Query().Get("term")
	dept := r.URL.Query().Get("dept")
	number := r.URL.Query().Get("number")
	ctx := r.Context()

	if term == "" {
		app.badRequestResponse(w, r, errors.New("term parameter is required"))
		return
	}

	year, term, err := splitYearTerm(term)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	sections, err := app.store.Sections.Get(ctx, year, term, dept, number)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, sections); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
