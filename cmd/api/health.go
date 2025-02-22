package main

import (
	"encoding/json"
	"net/http"
	"time"
)

type Commit struct {
	Commit struct {
		Author struct {
			Date string `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("https://api.github.com/repos/brianrahadi/sfucourses-api/commits?per_page=1")
	if err != nil {
		app.internalServerError(w, r, err)
		return
	}
	defer resp.Body.Close()

	var commits []Commit
	if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
		app.internalServerError(w, r, err)
		return
	}

	// Extract the latest commit date
	var lastUpdated string
	if len(commits) > 0 {
		commitDate, err := time.Parse(time.RFC3339, commits[0].Commit.Author.Date)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}
		lastUpdated = commitDate.Format(time.RFC3339)
	}

	data := map[string]string{
		"status":      "ok",
		"version":     version,
		"lastUpdated": lastUpdated,
	}

	if err := writeJSON(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
	}
}
