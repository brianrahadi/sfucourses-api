package main

import (
	"encoding/json"
	"net/http"
	"time"
)

var BuildTime string = time.Now().Format(time.RFC3339)

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
	app.lastDataUpdateLock.RLock()
	lastUpdate := app.lastDataUpdate
	app.lastDataUpdateLock.RUnlock()

	resp := HealthResponse{
		Status:         "ok",
		Version:        "1.0.0", // or your actual version
		LastDataUpdate: "",
	}
	if !lastUpdate.IsZero() {
		resp.LastDataUpdate = lastUpdate.Format(time.RFC3339)
	} else if BuildTime != "" {
		resp.LastDataUpdate = BuildTime
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
