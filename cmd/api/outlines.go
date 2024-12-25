package main

import (
	"errors"
	"net/http"

	"github.com/brianrahadi/sfucourses-api/internal/store"
)

func (app *application) getAllCourseOutlines(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	outlines, err := app.store.Outlines.GetAll(ctx)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)

		}
		return
	}

	if err := writeJSON(w, http.StatusOK, outlines); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
