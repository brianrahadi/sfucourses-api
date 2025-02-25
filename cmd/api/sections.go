package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/store"
)

// Helper function to split yearTerm parameter into year and term components
func splitYearTerm(yearTerm string) (year, term string, err error) {
	parts := strings.Split(yearTerm, "-")
	if len(parts) != 2 {
		return "", "", errors.New("invalid year-term format")
	}
	return parts[0], parts[1], nil
}

// Helper function to parse withOutlines query parameter
func getWithOutlines(app *application, w http.ResponseWriter, r *http.Request) bool {
	withOutlines := false
	withOutlinesStr := r.URL.Query().Get("withOutlines")
	if withOutlinesStr != "" {
		var err error
		withOutlines, err = strconv.ParseBool(withOutlinesStr)
		if err != nil {
			app.badRequestResponse(w, r, err)
			return false
		}
	}
	return withOutlines
}

// @Summary		Get sections by term
// @Description	Retrieves all course sections for a specific year and term
// @Tags			sections
// @Accept			json
// @Produce		json
// @Param			yearTerm		path		string									true	"Year and term in format YYYY-Term (e.g., 2024-Spring)"
// @Param			withOutlines	query		boolean									false	"Whether to include course outline data (default: false)"
// @Success		200				{array}		[]model.CourseWithSectionDetails		"List of sections without outlines"
// @Success		200				{array}		[]model.CourseOutlineWithSectionDetails	"List of sections with outlines (if withOutlines=true)"
// @Failure		400				{object}	ErrorResponse							"Invalid yearTerm format or query parameters"
// @Failure		404				{object}	ErrorResponse							"No sections found for the specified term"
// @Failure		500				{object}	ErrorResponse							"Internal server error"
// @Router			/sections/{yearTerm} [get]
func (app *application) getSectionsByTerm(w http.ResponseWriter, r *http.Request) {
	yearTerm := r.PathValue("yearTerm")
	ctx := r.Context()

	year, term, err := splitYearTerm(yearTerm)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	withOutlines := getWithOutlines(app, w, r)

	if withOutlines {
		sectionsWithOutlines, err := app.store.SectionsWithOutlines.GetByTerm(ctx, year, term)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		if err := writeJSON(w, http.StatusOK, sectionsWithOutlines); err != nil {
			app.internalServerError(w, r, err)
			return
		}
		return
	}

	sections, err := app.store.Sections.GetByTerm(ctx, year, term)
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

// @Summary		Get sections by term and department
// @Description	Retrieves all course sections for a specific year, term, and department
// @Tags			sections
// @Accept			json
// @Produce		json
// @Param			yearTerm		path		string									true	"Year and term in format YYYY-Term (e.g., 2024-Spring)"
// @Param			dept			path		string									true	"Department code (e.g., CMPT, MATH)"
// @Param			withOutlines	query		boolean									false	"Whether to include course outline data (default: false)"
// @Success		200				{array}		[]model.CourseWithSectionDetails		"List of sections without outlines"
// @Success		200				{array}		[]model.CourseOutlineWithSectionDetails	"List of sections with outlines (if withOutlines=true)"
// @Failure		400				{object}	ErrorResponse							"Invalid yearTerm format or query parameters"
// @Failure		404				{object}	ErrorResponse							"No sections found for the specified term and department"
// @Failure		500				{object}	ErrorResponse							"Internal server error"
// @Router			/sections/{yearTerm}/{dept} [get]
func (app *application) getSectionsByTermAndDept(w http.ResponseWriter, r *http.Request) {
	yearTerm := r.PathValue("yearTerm")
	dept := r.PathValue("dept")
	ctx := r.Context()

	year, term, err := splitYearTerm(yearTerm)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	withOutlines := getWithOutlines(app, w, r)

	if withOutlines {
		sectionsWithOutlines, err := app.store.SectionsWithOutlines.GetByTermAndDept(ctx, year, term, dept)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		if err := writeJSON(w, http.StatusOK, sectionsWithOutlines); err != nil {
			app.internalServerError(w, r, err)
			return
		}
		return
	}
	sections, err := app.store.Sections.GetByTermAndDept(ctx, year, term, dept)
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

// @Summary		Get sections by term, department, and course number
// @Description	Retrieves all course sections for a specific year, term, department, and course number
// @Tags			sections
// @Accept			json
// @Produce		json
// @Param			yearTerm		path		string									true	"Year and term in format YYYY-Term (e.g., 2024-Spring)"
// @Param			dept			path		string									true	"Department code (e.g., CMPT, MATH)"
// @Param			number			path		string									true	"Course number (e.g., 120, 225)"
// @Param			withOutlines	query		boolean									false	"Whether to include course outline data (default: false)"
// @Success		200				{array}		[]model.CourseWithSectionDetails		"List of sections without outlines"
// @Success		200				{array}		[]model.CourseOutlineWithSectionDetails	"List of sections with outlines (if withOutlines=true)"
// @Failure		400				{object}	ErrorResponse							"Invalid yearTerm format or query parameters"
// @Failure		404				{object}	ErrorResponse							"No sections found for the specified criteria"
// @Failure		500				{object}	ErrorResponse							"Internal server error"
// @Router			/sections/{yearTerm}/{dept}/{number} [get]
func (app *application) getSectionsByTermAndDeptAndNumber(w http.ResponseWriter, r *http.Request) {
	yearTerm := r.PathValue("yearTerm")
	dept := r.PathValue("dept")
	number := r.PathValue("number")
	ctx := r.Context()

	year, term, err := splitYearTerm(yearTerm)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	withOutlines := getWithOutlines(app, w, r)

	if withOutlines {
		sectionsWithOutlines, err := app.store.SectionsWithOutlines.GetByTermAndDeptAndNumber(ctx, year, term, dept, number)
		if err != nil {
			switch {
			case errors.Is(err, store.ErrNotFound):
				app.notFoundResponse(w, r, err)
			default:
				app.internalServerError(w, r, err)
			}
			return
		}

		if err := writeJSON(w, http.StatusOK, sectionsWithOutlines); err != nil {
			app.internalServerError(w, r, err)
			return
		}
		return
	}
	sections, err := app.store.Sections.GetByTermAndDeptAndNumber(ctx, year, term, dept, number)
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
