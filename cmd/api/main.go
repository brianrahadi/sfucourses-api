// Package main provides the entry point for the SFU Courses API.
//
// @title SFU Courses API
// @version 0.0.1
// @description API for accessing SFU course schedules and outlines
// @host api.sfucourses.com
// @BasePath /v1/rest
package main

import (
	"log"

	"github.com/brianrahadi/sfucourses-api/internal/store"
)

// Version of the API
const version = "0.0.1"

// @Summary Application entry point
// @Description Main entry point for the SFU Courses API
func main() {
	cfg := config{
		addr: ":8080",
	}

	store := store.NewStorage()

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
