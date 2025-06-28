package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/brianrahadi/sfucourses-api/internal/store"
	"github.com/go-co-op/gocron"
)

// application represents the application structure with configuration and data store
type application struct {
	config config
	store  store.Storage

	lastDataUpdate     time.Time
	lastDataUpdateLock sync.RWMutex
}

// config holds the application configuration
type config struct {
	addr   string
	env    string
	apiURL string
}

// dbConfig holds the database configuration
type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

// gzipResponseWriter wraps http.ResponseWriter to support gzip compression
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// mount sets up all the API routes
//
//	@Summary		Mount all API routes
//	@Description	Sets up all the routes for the API
//	@Return			http.Handler
func (app *application) mount() http.Handler {
	mux := http.NewServeMux()

	// mux.HandleFunc("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	if r.URL.Path != "/" {
	// 		http.NotFound(w, r)
	// 		return
	// 	}
	// 	w.Header().Set("Content-Type", "text/html")
	// 	html := `
	// 	<!DOCTYPE html>
	// 	<html>
	// 	<head>
	// 		<title>Welcome</title>
	// 	</head>
	// 	<body>
	// 		<h1>Welcome to sfucourses API</h1>
	// 		<a href="./docs">Go to docs</a>
	// 	</body>
	// 	</html>
	// 	`
	// 	w.Write([]byte(html))
	// }))

	// docsURL := fmt.Sprintf("%s/swagger/doc.json", app.config.addr)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" && r.URL.Path != "/docs" {
			http.NotFound(w, r)
			return
		}
		htmlContent, err := scalar.ApiReferenceHTML(&scalar.Options{
			SpecURL: "./docs/swagger.json",
			CustomOptions: scalar.CustomOptions{
				PageTitle: "sfucourses API",
			},
			Theme:    scalar.ThemeDefault,
			DarkMode: true,
		})

		if err != nil {
			fmt.Printf("%v", err)
		}

		fmt.Fprintln(w, htmlContent)
	})

	mux.HandleFunc("GET /health", app.healthCheckHandler)
	mux.HandleFunc("POST /update", app.manualUpdateHandler)

	mux.HandleFunc("GET /v1/rest/outlines/all", app.getAllCourseOutlines)
	mux.HandleFunc("GET /v1/rest/outlines/{dept}", app.getCourseOutlinesByDept)
	mux.HandleFunc("GET /v1/rest/outlines/{dept}/{number}", app.getCourseOutlinesByDeptAndNumber)

	mux.HandleFunc("GET /v1/rest/sections/{yearTerm}", app.getSectionsByTerm)
	mux.HandleFunc("GET /v1/rest/sections/{yearTerm}/{dept}", app.getSectionsByTermAndDept)
	mux.HandleFunc("GET /v1/rest/sections/{yearTerm}/{dept}/{number}", app.getSectionsByTermAndDeptAndNumber)

	mux.HandleFunc("GET /v1/rest/instructors", app.getInstructors)
	mux.HandleFunc("GET /v1/rest/instructors/names/{name}", app.getInstructorsByName)
	mux.HandleFunc("GET /v1/rest/instructors/{dept}", app.getInstructorsByDept)
	mux.HandleFunc("GET /v1/rest/instructors/{dept}/{number}", app.getInstructorsByDeptAndNumber)

	return mux
}

// Write implements the io.Writer interface for gzipResponseWriter
func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

// middleware applies common middleware to all requests
//
//	@Description	Apply common middleware including CORS, logging, timeout, and gzip compression
func (app *application) middleware(next http.Handler) http.Handler {
	return http.TimeoutHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Recovered from panic: %v", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")

		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Handle CORS preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && r.URL.Query().Get("gzip") == "true" {
			w.Header().Set("Content-Encoding", "gzip")
			gz := gzip.NewWriter(w)
			defer gz.Close()

			gzipWriter := &gzipResponseWriter{ResponseWriter: w, Writer: gz}
			next.ServeHTTP(gzipWriter, r)
			return
		}

		next.ServeHTTP(w, r)
	}), 60*time.Second, "Request timed out")
}

func (app *application) runDataSync() (int, int) {
	log.Printf("Starting data sync at %v", time.Now().UTC())

	year, term := getCurrentTerm()
	nextTermYear, nextTermTerm := getNextTerm()

	// Run all commands in sequence
	commands := []struct {
		name string
		args []string
	}{
		{"fetch-sections", []string{year, term}},
		{"fetch-sections", []string{nextTermYear, nextTermTerm}},
		{"sync-offerings", []string{}},
		{"sync-instructors", []string{}},
		{"fetch-instructors", []string{}},
	}

	successCount := 0
	totalCommands := len(commands)

	for _, cmdInfo := range commands {
		log.Printf("Running %s...", cmdInfo.name)

		cmd := exec.Command(fmt.Sprintf("./bin/%s", cmdInfo.name), cmdInfo.args...)

		// Capture both stdout and stderr
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Error running %s: %v\nOutput: %s", cmdInfo.name, err, output)
			continue // Continue with next command even if one fails
		}

		log.Printf("Successfully completed %s\nOutput: %s", cmdInfo.name, output)
		successCount++
	}

	// Force reload all data after sync
	if err := app.store.Outlines.ForceReload(); err != nil {
		log.Printf("Error reloading outlines: %v", err)
	}
	if err := app.store.Sections.ForceReload(); err != nil {
		log.Printf("Error reloading sections: %v", err)
	}
	if err := app.store.SectionsWithOutlines.ForceReload(); err != nil {
		log.Printf("Error reloading sections with outlines: %v", err)
	}
	if err := app.store.Instructors.ForceReload(); err != nil {
		log.Printf("Error reloading instructors: %v", err)
	}

	// trigger client ssg revalidation
	app.triggerRevalidation("revalidate-explore")
	app.triggerRevalidation("revalidate-schedule")

	app.lastDataUpdateLock.Lock()
	app.lastDataUpdate = time.Now().UTC()
	app.lastDataUpdateLock.Unlock()

	log.Printf("Completed data sync at %v", app.lastDataUpdate)
	return successCount, totalCommands
}

// manualUpdateHandler triggers a manual data update
func (app *application) manualUpdateHandler(w http.ResponseWriter, r *http.Request) {
	// Check if request method is POST
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Check password from environment variable
	expectedPassword := os.Getenv("UPDATE_PASSWORD")
	if expectedPassword == "" {
		http.Error(w, "Update password not configured", http.StatusInternalServerError)
		return
	}

	if string(body) != expectedPassword {
		print(string(body), expectedPassword)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	successCount, totalCommands := app.runDataSync()

	// Return response
	response := map[string]interface{}{
		"status":        "completed",
		"successCount":  successCount,
		"totalCommands": totalCommands,
		"lastUpdate":    app.lastDataUpdate.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// startCronJobs initializes and starts all cron jobs
func (app *application) startCronJobs() {
	s := gocron.NewScheduler(time.UTC)

	_, err := s.Every(1).Hours().At("00:00").Do(func() {
		app.runDataSync()
	})

	if err != nil {
		log.Printf("Error scheduling cron job: %v", err)
		return
	}

	s.StartAsync()
}

// getCurrentTerm returns the current academic year and term
func getCurrentTerm() (string, string) {
	return getTermForDate(time.Now())
}

// getTermForDate determines the academic term for a given date
func getTermForDate(date time.Time) (string, string) {
	year := fmt.Sprintf("%d", date.Year())

	// Determine the term based on the month
	var term string
	switch {
	case date.Month() >= time.January && date.Month() <= time.April:
		term = "spring"
	case date.Month() >= time.May && date.Month() <= time.August:
		term = "summer"
	case date.Month() >= time.September && date.Month() <= time.December:
		term = "fall"
	}

	return year, term
}

// getNextTerm returns the next academic term
func getNextTerm() (string, string) {
	now := time.Now()
	year, currentTerm := getTermForDate(now)
	currentYear, _ := strconv.Atoi(year)

	var nextDate time.Time
	switch currentTerm {
	case "spring":
		nextDate = time.Date(currentYear, time.May, 1, 0, 0, 0, 0, time.UTC)
	case "summer":
		nextDate = time.Date(currentYear, time.September, 1, 0, 0, 0, 0, time.UTC)
	case "fall":
		nextDate = time.Date(currentYear+1, time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	return getTermForDate(nextDate)
}

// run starts the HTTP server
//
//	@Description	Start the HTTP server with the provided handler and configuration
func (app *application) run(mux http.Handler) error {
	// Start cron jobs
	app.startCronJobs()

	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      app.middleware(mux),
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server has started at %s", srv.Addr)

	return srv.ListenAndServe()
}

// HealthResponse represents the health check response
//
//	@Description	Health check status information
type HealthResponse struct {
	Status         string `json:"status" example:"ok"`
	Version        string `json:"version,omitempty" example:"1.0.0"`
	LastDataUpdate string `json:"lastDataUpdate,omitempty" example:"2025-06-17T07:40:46Z"`
}

// OutlineResponse represents a course outline response
//
//	@Description	Course outline information
type OutlineResponse struct {
	Department    string `json:"department"`
	CourseNumber  string `json:"courseNumber"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	Prerequisites string `json:"prerequisites,omitempty"`
	Corequisites  string `json:"corequisites,omitempty"`
	Credits       int    `json:"credits"`
	OutlineURL    string `json:"outlineUrl"`
	LastUpdated   string `json:"lastUpdated"`
}

// SectionResponse represents a course section response
//
//	@Description	Course section information
type SectionResponse struct {
	YearTerm       string `json:"yearTerm"`
	Department     string `json:"department"`
	CourseNumber   string `json:"courseNumber"`
	Section        string `json:"section"`
	Title          string `json:"title"`
	InstructorName string `json:"instructorName,omitempty"`
	Days           string `json:"days,omitempty"`
	Time           string `json:"time,omitempty"`
	Location       string `json:"location,omitempty"`
	Capacity       int    `json:"capacity"`
	Enrolled       int    `json:"enrolled"`
}

// ErrorResponse represents an error response
//
//	@Description	Error information
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Update triggerRevalidation to accept a type parameter and add it to the query string
func (app *application) triggerRevalidation(revalidationType string) {
	revalidateSecret := os.Getenv("REVALIDATE_SECRET")
	if revalidateSecret == "" {
		log.Printf("REVALIDATE_SECRET not set, skipping revalidation call")
		return
	}
	url := "https://sfucourses.com/api/revalidate?secret=" + revalidateSecret + "&type=" + revalidationType
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		log.Printf("Error creating revalidate request: %v", err)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error calling revalidate endpoint: %v", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Revalidate response: %s", string(body))
}
