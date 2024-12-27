package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/brianrahadi/sfucourses-api/internal/store"
	"github.com/samber/mo"
)

func (app *application) getAllCourseOutlines(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get pagination parameters from query string
	limitOpt := mo.None[int]()
	offsetOpt := mo.None[int]()

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit < 1 {
			app.badRequestResponse(w, r, errors.New("invalid limit parameter"))
			return
		}
		limitOpt = mo.Some(parsedLimit)
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err != nil || parsedOffset < 0 {
			app.badRequestResponse(w, r, errors.New("invalid offset parameter"))
			return
		}
		offsetOpt = mo.Some(parsedOffset)
	}

	outlines, totalCount, err := app.store.Outlines.GetAll(ctx, limitOpt, offsetOpt)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrNotFound):
			app.notFoundResponse(w, r, err)
		default:
			app.internalServerError(w, r, err)
		}
		return
	}

	response := model.CourseOutlinesResponse{
		Data:       outlines,
		TotalCount: totalCount,
	}

	if limitOpt.IsPresent() {
		limit := limitOpt.MustGet()
		var nextURL string
		nextOffset := offsetOpt.OrElse(0) + limit
		if nextOffset < totalCount {
			// Create a copy of the current URL
			u := *r.URL
			q := u.Query()
			q.Set("limit", strconv.Itoa(limit))
			q.Set("offset", strconv.Itoa(nextOffset))
			u.RawQuery = q.Encode()
			nextURL = u.String()
		}
		response.NextURL = nextURL
	}

	if err := writeJSON(w, http.StatusOK, response); err != nil {
		app.internalServerError(w, r, err)
		return
	}
}

func (app *application) getCourseOutlinesByDept(w http.ResponseWriter, r *http.Request) {
	dept := r.PathValue("dept")
	ctx := r.Context()

	outlines, err := app.store.Outlines.GetByDept(ctx, dept)
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
