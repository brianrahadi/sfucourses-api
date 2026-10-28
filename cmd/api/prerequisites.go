package main

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/brianrahadi/sfucourses-api/internal/store"
)

// @Summary		Get parsed course prerequisites
// @Description	Returns prerequisite expression trees for all courses, optionally filtered by department and/or course number
// @Tags			Prerequisites
// @Accept			json
// @Produce		json
// @Param			dept	query		string			false	"Department code (e.g., cmpt, math)"
// @Param			number	query		string			false	"Course number (e.g., 120, 225)"
// @Success		200		{object}	model.PrereqMap	"Map of course codes to prerequisite expression trees"
// @Failure		404		{object}	ErrorResponse	"No prerequisites found for the specified criteria"
// @Failure		500		{object}	ErrorResponse	"Internal server error"
// @Router			/v1/rest/prerequisites [get]
func (app *application) getPrerequisites(w http.ResponseWriter, r *http.Request) {
	dept := r.URL.Query().Get("dept")
	number := r.URL.Query().Get("number")

	if dept == "" && number != "" {
		app.notFoundResponse(w, r, store.ErrNotFound)
		return
	}

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

	if dept != "" && number != "" && len(filtered) == 0 {
		app.notFoundResponse(w, r, store.ErrNotFound)
		return
	}

	writeJSON(w, http.StatusOK, filtered)
}

// @Summary		Get the course prerequisite graph
// @Description	Returns one entry per course with the course codes it requires (prerequisites) and may take alongside (corequisites), parsed from the catalog text. Unit counts and permission clauses are omitted; use /v1/rest/prerequisites for the full and/or logic. Refreshed weekly from the newest SFU terms.
// @Tags			Prerequisites
// @Produce		json
// @Param			dept	query		string						false	"Only courses in this department (e.g., cmpt)"
// @Success		200		{array}		model.PrereqGraphEntry		"Prerequisite graph entries"
// @Router			/v1/rest/prerequisites/graph [get]
func (app *application) getPrerequisiteGraph(w http.ResponseWriter, r *http.Request) {
	graph := app.store.Outlines.GetPrereqGraph()
	if dept := strings.ToUpper(r.URL.Query().Get("dept")); dept != "" {
		filtered := make([]model.PrereqGraphEntry, 0)
		for _, g := range graph {
			if g.Dept == dept {
				filtered = append(filtered, g)
			}
		}
		graph = filtered
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	writeJSON(w, http.StatusOK, graph)
}

const prereqMetaFile = "internal/store/json/prereq_meta.json"

// readPrereqsLastUpdated returns when requirement text last changed (RFC3339),
// as recorded by scripts/syncPrereqs, or "" if unavailable.
func readPrereqsLastUpdated() string {
	data, err := os.ReadFile(prereqMetaFile)
	if err != nil {
		return ""
	}
	var meta struct {
		LastUpdated string `json:"lastUpdated"`
	}
	if err := json.Unmarshal(data, &meta); err != nil {
		return ""
	}
	t, err := time.Parse(time.RFC3339, meta.LastUpdated)
	if err != nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
