package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/brianrahadi/sfucourses-api/internal/store"
)

// @Summary		Get course outlines
// @Description	Retrieves course outlines, optionally filtered by department and/or course number
// @Tags			Outlines
// @Accept			json
// @Produce		json
// @Param			dept	query		string				false	"Department code (e.g., cmpt, math)"
// @Param			number	query		string				false	"Course number (e.g., 120, 225)"
// @Param			short	query		bool				false	"Return short outline with only dept, number, title, and units"
// @Param			SHORT	query		bool				false	"Return short outline with only dept, number, title, and units"
// @Success		200		{array}		model.CourseOutline	"List of course outlines"
// @Failure		404		{object}	ErrorResponse		"No outlines found for the specified criteria"
// @Failure		500		{object}	ErrorResponse		"Internal server error"
// @Router			/v1/rest/outlines [get]
func (app *application) getCourseOutlines(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dept := r.URL.Query().Get("dept")
	number := r.URL.Query().Get("number")
	
	isShort := strings.ToLower(r.URL.Query().Get("short")) == "true" || strings.ToLower(r.URL.Query().Get("SHORT")) == "true"

	outlines, err := app.store.Outlines.Get(ctx, dept, number)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if isShort {
		shortOutlines := make([]model.ShortCourseOutline, 0, len(outlines))
		for _, o := range outlines {
			shortOutlines = append(shortOutlines, model.ShortCourseOutline{
				Dept:   o.Dept,
				Number: o.Number,
				Title:  o.Title,
				Units:  o.Units,
			})
		}
		writeJSON(w, http.StatusOK, shortOutlines)
		return
	}

	writeJSON(w, http.StatusOK, outlines)
}
