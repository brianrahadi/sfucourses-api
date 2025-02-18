package main

import (
	"log"

	"github.com/brianrahadi/sfucourses-api/internal/store"
)

const version = "0.0.1"

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
