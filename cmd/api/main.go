package main

import (
	"log"

	"github.com/brianrahadi/sfucourses-api/internal/store"
)

// Version of the API
const version = "0.0.1"

// Package main provides the entry point for the SFU Courses API.
//
//	@title			SFU Courses API
//	@description	API for accessing SFU course outlines, sections, and instructors
//	@BasePath		/v1/rest
//	@host			api.sfucourses.com

// @description
func main() {
	cfg := config{
		addr:   ":8080",
		env:    "dev",
		apiURL: "api.sfucourses.com",
	}

	store := store.NewStorage()

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
