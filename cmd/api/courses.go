package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/store"
)

func splitYearTerm(yearTerm string) (year, term string, err error) {
	parts := strings.Split(yearTerm, "-")
	if len(parts) != 2 {
		return "", "", errors.New("invalid year-term format")
	}
	return parts[0], parts[1], nil
}

func (app *application) getCoursesByTerm(w http.ResponseWriter, r *http.Request) {
	yearTerm := r.PathValue("yearTerm")
	ctx := r.Context()

	year, term, err := splitYearTerm(yearTerm)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

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
	yearTerm := r.PathValue("yearTerm")
	dept := r.PathValue("dept")
	ctx := r.Context()

	year, term, err := splitYearTerm(yearTerm)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

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
	yearTerm := r.PathValue("yearTerm")
	dept := r.PathValue("dept")
	number := r.PathValue("number")
	ctx := r.Context()

	year, term, err := splitYearTerm(yearTerm)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

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
