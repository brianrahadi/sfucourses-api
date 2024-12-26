package main

import (
	"errors"
	"net/http"

	"github.com/brianrahadi/sfucourses-api/internal/store"
)

func (app *application) getCoursesByTerm(w http.ResponseWriter, r *http.Request) {
	year := r.PathValue("year")
	term := r.PathValue("term")
	ctx := r.Context()

	courses, err := app.store.Courses.GetByTerm(ctx, year, term)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)

		}
		return
	}

	if err := writeJSON(w, http.StatusOK, courses); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getCoursesByTermAndDept(w http.ResponseWriter, r *http.Request) {
	year := r.PathValue("year")
	term := r.PathValue("term")
	dept := r.PathValue("dept")
	ctx := r.Context()

	courses, err := app.store.Courses.GetByTermAndDept(ctx, year, term, dept)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)

		}
		return
	}

	if err := writeJSON(w, http.StatusOK, courses); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getCoursesByTermAndDeptAndNumber(w http.ResponseWriter, r *http.Request) {
	year := r.PathValue("year")
	term := r.PathValue("term")
	dept := r.PathValue("dept")
	number := r.PathValue("number")
	ctx := r.Context()

	courses, err := app.store.Courses.GetByTermAndDeptAndNumber(ctx, year, term, dept, number)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)

		}
		return
	}

	if err := writeJSON(w, http.StatusOK, courses); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}
