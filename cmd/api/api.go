package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/brianrahadi/sfucourses-api/internal/store"
)

// application represents the application structure with configuration and data store
type application struct {
	config config
	store  store.Storage
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
	// 		<h1>Welcome to SFU Courses API</h1>
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
				PageTitle: "SFU Courses API",
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

	mux.HandleFunc("GET /v1/rest/outlines/all", app.getAllCourseOutlines)
	mux.HandleFunc("GET /v1/rest/outlines/{dept}", app.getCourseOutlinesByDept)
	mux.HandleFunc("GET /v1/rest/outlines/{dept}/{number}", app.getCourseOutlinesByDeptAndNumber)

	mux.HandleFunc("GET /v1/rest/sections/{yearTerm}", app.getSectionsByTerm)
	mux.HandleFunc("GET /v1/rest/sections/{yearTerm}/{dept}", app.getSectionsByTermAndDept)
	mux.HandleFunc("GET /v1/rest/sections/{yearTerm}/{dept}/{number}", app.getSectionsByTermAndDeptAndNumber)

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

// run starts the HTTP server
//
//	@Description	Start the HTTP server with the provided handler and configuration
func (app *application) run(mux http.Handler) error {

	// // Docs
	// docs.SwaggerInfo.Version = version
	// docs.SwaggerInfo.Host = "api.sfucourses.com"
	// // docs.SwaggerInfo.BasePath = "/v1/rest"
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
	Status  string `json:"status" example:"ok"`
	Version string `json:"version,omitempty" example:"1.0.0"`
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
