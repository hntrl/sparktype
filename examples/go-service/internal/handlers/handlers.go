// Package handlers provides HTTP handlers for the blog API.
//
// This demonstrates how to use sparktype-generated types in Go HTTP handlers:
// - Request body parsing into generated structs
// - Response serialization from generated structs
// - Type-safe error responses
package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/example/blog-api/internal/store"
	"github.com/example/blog-api/internal/types"
)

// Handler holds dependencies for HTTP handlers.
type Handler struct {
	store *store.Store
}

// New creates a new Handler with the given store.
func New(s *store.Store) *Handler {
	return &Handler{store: s}
}

// ===========================================================================
// Response helpers
// ===========================================================================

// writeJSON writes a JSON response with the given status code.
// The value should be a generated type from sparktype.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

// writeError writes an error response using the generated ErrorResponse type.
func writeError(w http.ResponseWriter, status int, errCode, message string) {
	resp := types.ErrorResponse{
		Error:   errCode,
		Message: message,
	}
	writeJSON(w, status, resp)
}

// ===========================================================================
// Post handlers
// ===========================================================================

// ListPosts handles GET /posts
func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	var status *types.PostStatus
	if s := r.URL.Query().Get("status"); s != "" {
		ps := types.PostStatus(s)
		status = &ps
	}

	var authorID *string
	if id := r.URL.Query().Get("author_id"); id != "" {
		authorID = &id
	}

	var tag *string
	if t := r.URL.Query().Get("tag"); t != "" {
		tag = &t
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Fetch posts - returns *types.PostList
	posts, err := h.store.ListPosts(status, authorID, tag, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// Response is the generated PostList type
	writeJSON(w, http.StatusOK, posts)
}

// GetPost handles GET /posts/{slug}
func (h *Handler) GetPost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Returns *types.Post
	post, err := h.store.GetPost(slug)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, post)
}

// CreatePost handles POST /posts
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	// Parse request body into generated CreatePostRequest type
	var req types.CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON body")
		return
	}

	// Basic validation - in production use a validation library
	if req.Title == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Title is required")
		return
	}
	if req.Content == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Content is required")
		return
	}

	// In a real app, get author ID from authentication context
	authorID := "550e8400-e29b-41d4-a716-446655440001"

	// Create post - returns *types.Post
	post, err := h.store.CreatePost(req, authorID)
	if err != nil {
		if errors.Is(err, store.ErrAlreadyExists) {
			writeError(w, http.StatusConflict, "duplicate_slug", "A post with this slug already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, post)
}

// UpdatePost handles PUT /posts/{slug}
func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Parse into generated UpdatePostRequest type
	var req types.UpdatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON body")
		return
	}

	// Update post - returns *types.Post
	post, err := h.store.UpdatePost(slug, req)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, post)
}

// DeletePost handles DELETE /posts/{slug}
func (h *Handler) DeletePost(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	if err := h.store.DeletePost(slug); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ===========================================================================
// Comment handlers
// ===========================================================================

// ListComments handles GET /posts/{slug}/comments
func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Returns []types.Comment
	comments, err := h.store.ListComments(slug)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, comments)
}

// CreateComment handles POST /posts/{slug}/comments
func (h *Handler) CreateComment(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")

	// Parse into generated CreateCommentRequest type
	var req types.CreateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON body")
		return
	}

	// Validation
	if req.AuthorName == "" || req.Content == "" {
		writeError(w, http.StatusBadRequest, "validation_error", "Author name and content are required")
		return
	}

	// Create comment - returns *types.Comment
	comment, err := h.store.CreateComment(slug, req)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Post not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

// ===========================================================================
// Author handlers
// ===========================================================================

// ListAuthors handles GET /authors
func (h *Handler) ListAuthors(w http.ResponseWriter, r *http.Request) {
	// Returns []types.Author
	authors := h.store.ListAuthors()
	writeJSON(w, http.StatusOK, authors)
}

// GetAuthor handles GET /authors/{id}
func (h *Handler) GetAuthor(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Returns *types.Author
	author, err := h.store.GetAuthor(id)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "not_found", "Author not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, author)
}

