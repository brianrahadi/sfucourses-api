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
//	@description	Unofficial API for accessing SFU course outlines, sections, and instructors in a seamless way. All data is pulled from SFU Course Outline REST API. This API is not affiliated with Simon Fraser University.
//	@openapi		3.1.0
//	@schemes		https
//	@host			api.sfucourses.com

//	@tag.name			Health
//	@tag.description	Health endpoints for monitoring API status and availability

// @tag.externalDocs.description	Health
// @tag.externalDocs.url			https://example.com/health-docs
//
// @tag.name						Outlines
// @tag.description				Outline endpoints for retrieving course outlines, including its offerings
// @tag.name						Sections
// @tag.description				Section endpoints for retrieving section info, including its schedules and instructor(s)
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
