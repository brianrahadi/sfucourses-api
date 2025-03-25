package main

import (
	"log"

	"github.com/brianrahadi/sfucourses-api/internal/db"
	"github.com/brianrahadi/sfucourses-api/internal/env"
	"github.com/brianrahadi/sfucourses-api/internal/store"
)

// Version of the API
const version = "1.0.0"

// Package main provides the entry point for the SFU Courses API.
//
//	@title			SFU Courses API
//	@description	Unofficial API for accessing SFU course outlines, sections, and instructors robustly and used to power [sfucourses.com](https://sfucourses.com). Data is pulled from [SFU Course Outlines REST API](https://www.sfu.ca/outlines/help/api.html). This API is not affiliated with Simon Fraser University.
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
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/sfucourses?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 10),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
	}

	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Print("database connection pool established")

	store := store.NewStorage(db)

	app := &application{
		config: cfg,
		store:  store,
	}

	mux := app.mount()
	log.Fatal(app.run(mux))
}
