package main

import (
	"errors"
	"net/http"

	"github.com/brianrahadi/sfucourses-api/internal/store"
)

func (app *application) getSectionsWithOutlinesByTerm(w http.ResponseWriter, r *http.Request) {
	yearTerm := r.PathValue("yearTerm")
	ctx := r.Context()

	year, term, err := splitYearTerm(yearTerm)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	courses, err := app.store.SectionsWithOutlines.GetByTerm(ctx, year, term)
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

func (app *application) getSectionsWithOutlinesByTermAndDept(w http.ResponseWriter, r *http.Request) {
	yearTerm := r.PathValue("yearTerm")
	dept := r.PathValue("dept")
	ctx := r.Context()

	year, term, err := splitYearTerm(yearTerm)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	courses, err := app.store.SectionsWithOutlines.GetByTermAndDept(ctx, year, term, dept)
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

func (app *application) getSectionsWithOutlinesByTermAndDeptAndNumber(w http.ResponseWriter, r *http.Request) {
	yearTerm := r.PathValue("yearTerm")
	dept := r.PathValue("dept")
	number := r.PathValue("number")
	ctx := r.Context()

	year, term, err := splitYearTerm(yearTerm)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	courses, err := app.store.SectionsWithOutlines.GetByTermAndDeptAndNumber(ctx, year, term, dept, number)
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
