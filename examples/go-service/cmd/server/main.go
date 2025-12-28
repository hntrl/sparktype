// Package main is the entry point for the blog API server.
//
// This demonstrates a complete Go HTTP server using sparktype-generated types.
package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/example/blog-api/internal/handlers"
	"github.com/example/blog-api/internal/store"
)

func main() {
	// Initialize dependencies
	db := store.New()
	h := handlers.New(db)

	// Create router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Post routes
	r.Route("/posts", func(r chi.Router) {
		r.Get("/", h.ListPosts)
		r.Post("/", h.CreatePost)

		r.Route("/{slug}", func(r chi.Router) {
			r.Get("/", h.GetPost)
			r.Put("/", h.UpdatePost)
			r.Delete("/", h.DeletePost)

			// Nested comment routes
			r.Get("/comments", h.ListComments)
			r.Post("/comments", h.CreateComment)
		})
	})

	// Author routes
	r.Route("/authors", func(r chi.Router) {
		r.Get("/", h.ListAuthors)
		r.Get("/{id}", h.GetAuthor)
	})

	// Start server
	addr := ":8080"
	log.Printf("Starting server on %s", addr)
	log.Printf("API docs at http://localhost%s/posts", addr)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}
