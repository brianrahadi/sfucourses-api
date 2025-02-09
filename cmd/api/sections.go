package main

import (
	"errors"
	"net/http"
	"strconv"
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
