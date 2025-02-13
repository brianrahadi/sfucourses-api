package main

import (
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/brianrahadi/sfucourses-api/internal/store"
)

type application struct {
	config config
	store  store.Storage
}

type config struct {
	addr string
	env  string
}

type dbConfig struct {
	addr         string
	maxOpenConns int
	maxIdleConns int
	maxIdleTime  string
}

type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (app *application) mount() http.Handler {
	mux := http.NewServeMux()

	// Middleware for recover and logging
	mux.HandleFunc("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Welcome</title>
		</head>
		<body>
			<h1>Welcome to SFU Courses API</h1>
			<p>Use the API to access course schedules and outlines.</p>
		</body>
		</html>
		`
		w.Write([]byte(html))
	}))

	mux.HandleFunc("GET /v1/rest/health", app.healthCheckHandler)

	mux.HandleFunc("GET /v1/rest/outlines/all", app.getAllCourseOutlines)
	mux.HandleFunc("GET /v1/rest/outlines/{dept}", app.getCourseOutlinesByDept)
	mux.HandleFunc("GET /v1/rest/outlines/{dept}/{number}", app.getCourseOutlinesByDeptAndNumber)

	mux.HandleFunc("GET /v1/rest/sections/{yearTerm}", app.getSectionsByTerm)
	mux.HandleFunc("GET /v1/rest/sections/{yearTerm}/{dept}", app.getSectionsByTermAndDept)
	mux.HandleFunc("GET /v1/rest/sections/{yearTerm}/{dept}/{number}", app.getSectionsByTermAndDeptAndNumber)

	return mux
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	return g.Writer.Write(b)
}

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

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
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
func (app *application) run(mux http.Handler) error {
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
