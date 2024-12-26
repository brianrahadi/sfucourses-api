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

func (app *application) getCourseOutlinesByDept(w http.ResponseWriter, r *http.Request) {
	deptID := r.PathValue("dept")
	ctx := r.Context()

	outlines, err := app.store.Outlines.GetByDept(ctx, deptID)
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

func (app *application) getCourseOutlinesByDeptAndNumber(w http.ResponseWriter, r *http.Request) {
	dept := r.PathValue("dept")
	number := r.PathValue("number")

	ctx := r.Context()

	outline, err := app.store.Outlines.GetByDeptAndNumber(ctx, dept, number)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	if err := writeJSON(w, http.StatusOK, outline); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
