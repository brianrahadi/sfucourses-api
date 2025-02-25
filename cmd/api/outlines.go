package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/brianrahadi/sfucourses-api/internal/store"
	"github.com/samber/mo"
)

// GetAllCourseOutlines godoc
//
//	@Summary		Get all course outlines
//	@Description	Retrieves a paginated list of all course outlines
//	@Tags			outlines
//	@Accept			json
//	@Produce		json
//	@Param			limit	query		int						false	"Number of items to return (pagination)"
//	@Param			offset	query		int						false	"Number of items to skip (pagination offset)"
//	@Success		200		{object}	[]model.CourseOutline	"List of course outlines with pagination info"
//	@Failure		404		{object}	ErrorResponse			"No course outlines found"
//	@Router			/outlines [get]
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

// @Summary		Get course outlines by department
// @Description	Retrieves all course outlines for a specific department
// @Tags			outlines
// @Accept			json
// @Produce		json
// @Param			dept	path		string					true	"Department code (e.g., CMPT, MATH)"
// @Success		200		{array}		[]model.CourseOutline	"List of course outlines for the department"
// @Failure		404		{object}	ErrorResponse			"Department not found or no courses available"
// @Failure		500		{object}	ErrorResponse			"Internal server error"
// @Router			/outlines/dept/{dept} [get]
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

// @Summary		Get specific course outline
// @Description	Retrieves course outline for a specific department and course number
// @Tags			outlines
// @Accept			json
// @Produce		json
// @Param			dept	path		string				true	"Department code (e.g., CMPT, MATH)"
// @Param			number	path		string				true	"Course number (e.g., 120, 225)"
// @Success		200		{object}	model.CourseOutline	"Course outline details"
// @Failure		404		{object}	ErrorResponse		"Course not found"
// @Failure		500		{object}	ErrorResponse		"Internal server error"
// @Router			/outlines/dept/{dept}/number/{number} [get]
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
