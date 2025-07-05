package main

import (
	"errors"
	"net/http"

	"github.com/brianrahadi/sfucourses-api/internal/store"
)

//	@Summary		Get course outlines
//	@Description	Retrieves course outlines, optionally filtered by department and/or course number
//	@Tags			Outlines
//	@Accept			json
//	@Produce		json
//	@Param			dept	query		string				false	"Department code (e.g., CMPT, MATH)"
//	@Param			number	query		string				false	"Course number (e.g., 120, 225)"
//	@Success		200		{array}		model.CourseOutline	"List of course outlines"
//	@Failure		404		{object}	ErrorResponse		"No outlines found for the specified criteria"
//	@Failure		500		{object}	ErrorResponse		"Internal server error"
//	@Router			/v1/rest/outlines [get]
func (app *application) getCourseOutlines(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	dept := r.URL.Query().Get("dept")
	number := r.URL.Query().Get("number")

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

	writeJSON(w, http.StatusOK, outlines)
}
