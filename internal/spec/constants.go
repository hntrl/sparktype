package spec

import "time"

const (
	// HTTPClientTimeout is the timeout for fetching remote OpenAPI specs.
	HTTPClientTimeout = 30 * time.Second

	// DefaultAcceptHeader is the Accept header sent when fetching remote specs.
	DefaultAcceptHeader = "application/json, application/yaml, text/yaml, */*"
)
