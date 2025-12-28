// Package store provides an in-memory data store for the blog API.
//
// This demonstrates how generated types flow through your data layer.
// In a real application, you would use a database like PostgreSQL with sqlx.
package store

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/example/blog-api/internal/types"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// Store is an in-memory store for blog data.
type Store struct {
	mu       sync.RWMutex
	posts    map[string]*types.Post    // keyed by slug
	authors  map[string]*types.Author  // keyed by ID
	comments map[string][]*types.Comment // keyed by post ID
}

// New creates a new store with sample data.
func New() *Store {
	s := &Store{
		posts:    make(map[string]*types.Post),
		authors:  make(map[string]*types.Author),
		comments: make(map[string][]*types.Comment),
	}
	s.seedData()
	return s
}

func (s *Store) seedData() {
	// Create sample author
	author := &types.Author{
		ID:        "550e8400-e29b-41d4-a716-446655440001",
		Name:      "Jane Developer",
		Email:     "jane@example.com",
		Bio:       stringPtr("Software engineer and technical writer."),
		AvatarUrl: stringPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=jane"),
		Website:   stringPtr("https://janedev.example.com"),
		SocialLinks: &types.SocialLinks{
			Twitter: stringPtr("janedev"),
			Github:  stringPtr("janedev"),
		},
		PostCount: intPtr(2),
	}
	s.authors[author.ID] = author

	now := time.Now()

	// Create sample posts
	post1 := &types.Post{
		ID:        "550e8400-e29b-41d4-a716-446655440010",
		Slug:      "getting-started-with-go",
		Title:     "Getting Started with Go",
		Excerpt:   stringPtr("A beginner's guide to the Go programming language."),
		Content:   "# Getting Started with Go\n\nGo is a statically typed, compiled language...",
		Status:    types.PostStatusPublished,
		Author:    *author,
		Tags:      []string{"go", "programming", "tutorial"},
		ViewCount:   intPtr(1523),
		CommentCount: intPtr(5),
		PublishedAt: timePtr(now.Add(-24 * time.Hour)),
		CreatedAt:   now.Add(-48 * time.Hour),
		UpdatedAt:   timePtr(now.Add(-24 * time.Hour)),
	}
	s.posts[post1.Slug] = post1

	post2 := &types.Post{
		ID:        "550e8400-e29b-41d4-a716-446655440011",
		Slug:      "type-safe-apis-with-sparktype",
		Title:     "Building Type-Safe APIs with sparktype",
		Excerpt:   stringPtr("How to use sparktype for end-to-end type safety."),
		Content:   "# Type-Safe APIs\n\nsparktype generates types from OpenAPI specs...",
		Status:    types.PostStatusPublished,
		Author:    *author,
		Tags:      []string{"go", "typescript", "api", "sparktype"},
		ViewCount:   intPtr(892),
		CommentCount: intPtr(3),
		PublishedAt: timePtr(now.Add(-12 * time.Hour)),
		CreatedAt:   now.Add(-36 * time.Hour),
		UpdatedAt:   timePtr(now.Add(-12 * time.Hour)),
	}
	s.posts[post2.Slug] = post2

	// Draft post
	post3 := &types.Post{
		ID:        "550e8400-e29b-41d4-a716-446655440012",
		Slug:      "upcoming-features",
		Title:     "Upcoming Features in Go 1.23",
		Content:   "# Draft\n\nThis post is still being written...",
		Status:    types.PostStatusDraft,
		Author:    *author,
		Tags:      []string{"go"},
		CreatedAt: now.Add(-2 * time.Hour),
	}
	s.posts[post3.Slug] = post3

	// Sample comments
	s.comments[post1.ID] = []*types.Comment{
		{
			ID:          "550e8400-e29b-41d4-a716-446655440020",
			PostId:      post1.ID,
			AuthorName:  "Alice",
			AuthorEmail: stringPtr("alice@example.com"),
			Content:     "Great introduction! Very helpful for beginners.",
			Approved:    boolPtr(true),
			CreatedAt:   now.Add(-20 * time.Hour),
		},
		{
			ID:          "550e8400-e29b-41d4-a716-446655440021",
			PostId:      post1.ID,
			AuthorName:  "Bob",
			AuthorEmail: stringPtr("bob@example.com"),
			Content:     "Could you write a follow-up on concurrency?",
			Approved:    boolPtr(true),
			CreatedAt:   now.Add(-18 * time.Hour),
		},
	}
}

// ListPosts returns posts matching the given filters.
func (s *Store) ListPosts(status *types.PostStatus, authorID, tag *string, limit, offset int) (*types.PostList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var posts []types.Post
	for _, p := range s.posts {
		// Apply filters
		if status != nil && p.Status != *status {
			continue
		}
		if authorID != nil && p.Author.ID != *authorID {
			continue
		}
		if tag != nil && !containsTag(p.Tags, *tag) {
			continue
		}
		posts = append(posts, *p)
	}

	// Simple pagination
	total := len(posts)
	if offset >= len(posts) {
		posts = []types.Post{}
	} else {
		end := offset + limit
		if end > len(posts) {
			end = len(posts)
		}
		posts = posts[offset:end]
	}

	hasMore := offset+len(posts) < total

	return &types.PostList{
		Posts:   posts,
		Total:   total,
		HasMore: &hasMore,
	}, nil
}

// GetPost returns a post by slug.
func (s *Store) GetPost(slug string) (*types.Post, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	post, ok := s.posts[slug]
	if !ok {
		return nil, ErrNotFound
	}
	return post, nil
}

// CreatePost creates a new post.
func (s *Store) CreatePost(req types.CreatePostRequest, authorID string) (*types.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate slug from title if not provided
	slug := req.Slug
	if slug == nil {
		generated := generateSlug(req.Title)
		slug = &generated
	}

	// Check for duplicate slug
	if _, exists := s.posts[*slug]; exists {
		return nil, ErrAlreadyExists
	}

	// Get author
	author, ok := s.authors[authorID]
	if !ok {
		return nil, errors.New("author not found")
	}

	now := time.Now()
	status := types.PostStatusDraft
	if req.Status != nil {
		status = *req.Status
	}

	post := &types.Post{
		ID:        generateID(),
		Slug:      *slug,
		Title:     req.Title,
		Excerpt:   req.Excerpt,
		Content:   req.Content,
		Status:    status,
		Author:    *author,
		Tags:      req.Tags,
		CreatedAt: now,
	}

	if req.FeaturedImageUrl != nil {
		post.FeaturedImage = &types.Image{
			Url: *req.FeaturedImageUrl,
			Alt: req.Title,
		}
	}

	if status == types.PostStatusPublished {
		if req.PublishedAt != nil {
			post.PublishedAt = req.PublishedAt
		} else {
			post.PublishedAt = &now
		}
	}

	s.posts[*slug] = post
	return post, nil
}

// UpdatePost updates an existing post.
func (s *Store) UpdatePost(slug string, req types.UpdatePostRequest) (*types.Post, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, ok := s.posts[slug]
	if !ok {
		return nil, ErrNotFound
	}

	now := time.Now()

	if req.Title != nil {
		post.Title = *req.Title
	}
	if req.Excerpt != nil {
		post.Excerpt = req.Excerpt
	}
	if req.Content != nil {
		post.Content = *req.Content
	}
	if req.Tags != nil {
		post.Tags = req.Tags
	}

	// Handle status change
	if req.Status != nil {
		oldStatus := post.Status
		post.Status = *req.Status

		// Set publishedAt when publishing
		if oldStatus != types.PostStatusPublished && *req.Status == types.PostStatusPublished {
			post.PublishedAt = &now
		}
	}

	// Handle featured image - can be set to nil
	if req.FeaturedImageUrl != nil {
		post.FeaturedImage = &types.Image{
			Url: *req.FeaturedImageUrl,
			Alt: post.Title,
		}
	}

	post.UpdatedAt = &now

	return post, nil
}

// DeletePost deletes a post by slug.
func (s *Store) DeletePost(slug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.posts[slug]; !ok {
		return ErrNotFound
	}
	delete(s.posts, slug)
	return nil
}

// ListComments returns comments for a post.
func (s *Store) ListComments(postSlug string) ([]types.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// First get the post to get its ID
	post, ok := s.posts[postSlug]
	if !ok {
		return nil, ErrNotFound
	}

	comments := s.comments[post.ID]
	if comments == nil {
		return []types.Comment{}, nil
	}

	// Convert to value slice
	result := make([]types.Comment, len(comments))
	for i, c := range comments {
		result[i] = *c
	}
	return result, nil
}

// CreateComment creates a new comment on a post.
func (s *Store) CreateComment(postSlug string, req types.CreateCommentRequest) (*types.Comment, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	post, ok := s.posts[postSlug]
	if !ok {
		return nil, ErrNotFound
	}

	comment := &types.Comment{
		ID:          generateID(),
		PostId:      post.ID,
		ParentId:    req.ParentId,
		AuthorName:  req.AuthorName,
		AuthorEmail: req.AuthorEmail,
		AuthorUrl:   req.AuthorUrl,
		Content:     req.Content,
		Approved:    boolPtr(false), // Requires moderation
		CreatedAt:   time.Now(),
	}

	s.comments[post.ID] = append(s.comments[post.ID], comment)

	// Update comment count
	count := len(s.comments[post.ID])
	post.CommentCount = &count

	return comment, nil
}

// ListAuthors returns all authors.
func (s *Store) ListAuthors() []types.Author {
	s.mu.RLock()
	defer s.mu.RUnlock()

	authors := make([]types.Author, 0, len(s.authors))
	for _, a := range s.authors {
		authors = append(authors, *a)
	}
	return authors
}

// GetAuthor returns an author by ID.
func (s *Store) GetAuthor(id string) (*types.Author, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	author, ok := s.authors[id]
	if !ok {
		return nil, ErrNotFound
	}
	return author, nil
}

// Helper functions

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Simple cleanup - in production use a proper slugify library
	return slug
}

func generateID() string {
	// In production, use github.com/google/uuid
	return time.Now().Format("20060102150405.000000")
}

func stringPtr(s string) *string { return &s }
func intPtr(i int) *int          { return &i }
func boolPtr(b bool) *bool       { return &b }
func timePtr(t time.Time) *time.Time { return &t }

