package main

import (
	"log"
	"net/http"
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

func (app *application) mount() http.Handler {
	mux := http.NewServeMux()

	// Middleware for recover and logging
	mux.Handle("/", app.middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})))

	mux.HandleFunc("GET /v1/rest/health", app.healthCheckHandler)

	mux.HandleFunc("GET /v1/rest/outlines/all", app.getAllCourseOutlines)
	mux.HandleFunc("GET /v1/rest/outlines/{dept}", app.getCourseOutlinesByDept)
	mux.HandleFunc("GET /v1/rest/outlines/{dept}/{number}", app.getCourseOutlinesByDeptAndNumber)

	return mux
}

func (app *application) middleware(next http.Handler) http.Handler {
	return http.TimeoutHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("Recovered from panic: %v", rec)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	}), 60*time.Second, "Request timed out")
}

func (app *application) run(mux http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	log.Printf("Server has started at %s", srv.Addr)

	return srv.ListenAndServe()
}
