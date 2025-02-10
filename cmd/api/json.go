package main

import (
	"compress/gzip"
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) error {
	// Set JSON response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if w.Header().Get("Content-Encoding") == "gzip" {
		gz := gzip.NewWriter(w)
		defer gz.Close()
		return json.NewEncoder(gz).Encode(data)
	}

	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1_048_578
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string) error {
	type envelope struct {
		Error string `json:"error"`
	}
	return writeJSON(w, status, &envelope{Error: message})
}
