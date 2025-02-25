package main

import (
	"net/http"
)

// type Commit struct {
// 	Commit struct {
// 		Author struct {
// 			Date string `json:"date"`
// 		} `json:"author"`
// 	} `json:"commit"`
// }

// @Summary		Health check endpoint
// @Description	Returns status and version information about the API
// @Tags			Health
// @Produce		json
// @Success		200	{object}	HealthResponse	"Returns status and version information"
// @Failure		500	{object}	ErrorResponse	"Internal Server Error"
// @Router			/health [get]
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// resp, err := http.Get("https://api.github.com/repos/brianrahadi/sfucourses-api/commits?per_page=1")
	// if err != nil {
	// 	app.internalServerError(w, r, err)
	// 	return
	// }
	// defer resp.Body.Close()

	// var commits []Commit
	// if err := json.NewDecoder(resp.Body).Decode(&commits); err != nil {
	// 	app.internalServerError(w, r, err)
	// 	return
	// }

	// Extract the latest commit date
	// var lastUpdated string
	// if len(commits) > 0 {
	// 	commitDate, err := time.Parse(time.RFC3339, commits[0].Commit.Author.Date)
	// 	if err != nil {
	// 		app.internalServerError(w, r, err)
	// 		return
	// 	}
	// 	lastUpdated = commitDate.Format(time.RFC3339)
	// }

	data := HealthResponse{
		"ok",
		version,
	}

	if err := writeJSON(w, http.StatusOK, data); err != nil {
		app.internalServerError(w, r, err)
	}
}
