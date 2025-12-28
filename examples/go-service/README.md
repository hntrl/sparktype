# Go Service Example

This example demonstrates how to use sparktype in a Go HTTP service for type-safe API development.

## What This Example Shows

- **Generated Go structs** - Type definitions with JSON tags
- **HTTP handlers** - Request/response handling with generated types
- **Data layer** - How types flow through your storage layer
- **Enum handling** - Type-safe status values as string constants

## Project Structure

```
go-service/
├── openapi.yaml              # Blog API spec
├── typegen.jsonc             # sparktype configuration
├── go.mod
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
└── internal/
    ├── types/
    │   ├── doc.go
    │   └── api.go            # Generated Go types
    ├── handlers/
    │   └── handlers.go       # HTTP handlers
    └── store/
        └── store.go          # Data layer
```

## Getting Started

### 1. Generate types

```bash
sparktype generate
```

This generates `internal/types/api.go` with Go structs for all schemas.

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Run the server

```bash
go run ./cmd/server
```

The server starts at http://localhost:8080

### 4. Test the API

```bash
# List posts
curl http://localhost:8080/posts

# Get a specific post
curl http://localhost:8080/posts/getting-started-with-go

# Create a post
curl -X POST http://localhost:8080/posts \
  -H "Content-Type: application/json" \
  -d '{"title": "New Post", "content": "# Hello\n\nThis is a new post."}'

# Filter by status
curl "http://localhost:8080/posts?status=published"
```

## How It Works

### Generated Structs

sparktype generates Go structs with proper JSON tags:

```go
// Generated in internal/types/api.go
type Post struct {
    ID           string      `json:"id"`
    Slug         string      `json:"slug"`
    Title        string      `json:"title"`
    Excerpt      *string     `json:"excerpt,omitempty"`
    Content      string      `json:"content"`
    Status       PostStatus  `json:"status"`
    Author       Author      `json:"author"`
    Tags         []string    `json:"tags,omitempty"`
    FeaturedImage *Image     `json:"featuredImage,omitempty"`
    ViewCount    *int        `json:"viewCount,omitempty"`
    PublishedAt  *time.Time  `json:"publishedAt,omitempty"`
    CreatedAt    time.Time   `json:"createdAt"`
    UpdatedAt    *time.Time  `json:"updatedAt,omitempty"`
}

type PostStatus string

const (
    PostStatusDraft     PostStatus = "draft"
    PostStatusPublished PostStatus = "published"
    PostStatusArchived  PostStatus = "archived"
)
```

### Using Types in Handlers

The handlers use generated types for type-safe request/response handling:

```go
import "github.com/example/blog-api/internal/types"

// CreatePost handles POST /posts
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
    // Parse request body into generated type
    var req types.CreatePostRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        writeError(w, http.StatusBadRequest, "invalid_json", "Invalid JSON body")
        return
    }

    // req.Title, req.Content, req.Status are all properly typed
    if req.Title == "" {
        writeError(w, http.StatusBadRequest, "validation_error", "Title is required")
        return
    }

    // Create returns *types.Post
    post, err := h.store.CreatePost(req, authorID)
    if err != nil {
        // ...
    }

    // Response matches the Post schema
    writeJSON(w, http.StatusCreated, post)
}
```

### Type-Safe Error Responses

Error responses use the generated `ErrorResponse` type:

```go
func writeError(w http.ResponseWriter, status int, errCode, message string) {
    resp := types.ErrorResponse{
        Error:   errCode,
        Message: message,
    }
    writeJSON(w, status, resp)
}
```

### Enum Handling

The generated `PostStatus` type provides type safety for status values:

```go
// Filter by status
func (h *Handler) ListPosts(w http.ResponseWriter, r *http.Request) {
    var status *types.PostStatus
    if s := r.URL.Query().Get("status"); s != "" {
        ps := types.PostStatus(s)
        status = &ps
    }
    
    posts, err := h.store.ListPosts(status, ...)
}

// Use in business logic
if post.Status == types.PostStatusPublished {
    // ...
}
```

### Data Layer Integration

Types flow through your entire application:

```go
// Store methods return generated types
func (s *Store) GetPost(slug string) (*types.Post, error) {
    // ...
    return &types.Post{
        ID:      record.ID,
        Slug:    record.Slug,
        Title:   record.Title,
        Status:  record.Status,  // types.PostStatus
        Author:  *author,        // types.Author
        // ...
    }, nil
}
```

## CI Integration

Add type checking to your CI pipeline:

```yaml
# .github/workflows/ci.yml
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      
      - name: Install sparktype
        run: npm install -g sparktype
      
      - name: Check generated types
        run: sparktype check
      
      - name: Build
        run: go build ./...
      
      - name: Test
        run: go test ./...
```

## Key Patterns

### 1. Optional Fields

Optional fields in the OpenAPI spec become pointers in Go:

```go
type Post struct {
    Excerpt   *string    `json:"excerpt,omitempty"`   // Optional
    UpdatedAt *time.Time `json:"updatedAt,omitempty"` // Optional
}
```

### 2. Array Fields

Array fields default to `omitempty` for empty slices:

```go
type Post struct {
    Tags []string `json:"tags,omitempty"`
}
```

### 3. Nested Objects

Referenced schemas become separate types:

```go
type Post struct {
    Author        Author  `json:"author"`        // Required reference
    FeaturedImage *Image  `json:"featuredImage"` // Optional reference
}
```

### 4. Date-Time Handling

Fields with `format: date-time` use `time.Time`:

```go
type Post struct {
    CreatedAt   time.Time  `json:"createdAt"`
    PublishedAt *time.Time `json:"publishedAt,omitempty"`
}
```

## Makefile

Add a Makefile for common operations:

```makefile
.PHONY: generate build run

generate:
	sparktype generate

build: generate
	go build -o bin/server ./cmd/server

run: build
	./bin/server

check:
	sparktype check
	go vet ./...
```

## Next Steps

- Add middleware for authentication
- Integrate with a real database (PostgreSQL with sqlx)
- Add request validation using go-playground/validator
- Set up structured logging with slog

