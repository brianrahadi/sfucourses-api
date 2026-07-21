package main

import (
	"net/http"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/model"
)

// @Summary		Get parsed course prerequisites
// @Description	Returns prerequisite expression trees for all courses, optionally filtered by department and/or course number
// @Tags			Prerequisites
// @Accept			json
// @Produce		json
// @Param			dept	query		string				false	"Department code (e.g., cmpt, math)"
// @Param			number	query		string				false	"Course number (e.g., 120, 225)"
// @Success		200		{object}	model.PrereqMap		"Map of course codes to prerequisite expression trees"
// @Failure		500		{object}	ErrorResponse		"Internal server error"
// @Router			/v1/rest/prerequisites [get]
func (app *application) getPrerequisites(w http.ResponseWriter, r *http.Request) {
	dept := r.URL.Query().Get("dept")
	number := r.URL.Query().Get("number")

	prereqMap := app.store.Outlines.GetPrereqMap()

	if dept == "" && number == "" {
		writeJSON(w, http.StatusOK, prereqMap)
		return
	}

	dept = strings.ToUpper(dept)
	number = strings.ToUpper(number)

	filtered := make(model.PrereqMap)
	for code, node := range prereqMap {
		parts := strings.SplitN(code, " ", 2)
		if len(parts) != 2 {
			continue
		}
		codeDept := parts[0]
		codeNumber := parts[1]

		if dept != "" && number != "" {
			if codeDept == dept && codeNumber == number {
				filtered[code] = node
			}
		} else if dept != "" {
			if codeDept == dept {
				filtered[code] = node
			}
		}
	}

	writeJSON(w, http.StatusOK, filtered)
}
