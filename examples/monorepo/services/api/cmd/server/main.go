// Package main is the entry point for the API service.
//
// This service uses types shared with the frontend and worker services.
package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/example/monorepo/services/api/internal/types"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Example endpoint using generated types
	r.Get("/api/users/{id}", func(w http.ResponseWriter, r *http.Request) {
		// Response uses the generated User type
		user := types.User{
			ID:    chi.URLParam(r, "id"),
			Email: "user@example.com",
			Name:  "Example User",
			Role:  (*types.UserRole)(stringPtr("member")),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"healthy"}`))
	})

	log.Println("API server starting on :8080")
	http.ListenAndServe(":8080", r)
}

func stringPtr(s string) *string { return &s }

