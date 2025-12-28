// Package spec handles loading OpenAPI specifications and converting them to
// normalized Schema representations.
//
// This package provides the foundation for type generation by loading OpenAPI
// specs from various sources and converting them into a consistent internal
// representation that generators can work with.
//
// # Spec Loading
//
// The LoadSpec function supports three source types:
//
//   - Local files: YAML or JSON OpenAPI specs from the filesystem
//   - Remote URLs: Specs fetched via HTTP(S) with optional custom headers
//   - Inline schemas: Schema definitions embedded directly in the config
//
// The loader uses kin-openapi to parse and validate OpenAPI 3.x specifications.
//
// # Schema Normalization
//
// OpenAPI schemas are converted to the internal Schema type which normalizes:
//
//   - Type information (object, array, string, number, etc.)
//   - Property definitions with required/optional status
//   - Composition (allOf, oneOf, anyOf)
//   - References ($ref) - kept as-is for lazy resolution
//   - Constraints (min/max length, pattern, enum values, etc.)
//   - Metadata (description, format, nullable, readonly)
//
// # Registry
//
// The Registry provides lazy loading and caching of specs by name. When a spec
// is first requested, it's loaded and cached. Subsequent requests return the
// cached schemas. The registry returns deep copies to prevent modification of
// cached data.
//
// # Source Tracking
//
// Each Schema includes a SourceSpec field indicating which named spec it came
// from. This enables error messages to reference the original source and helps
// with debugging configuration issues.
//
// # Reference Resolution
//
// References ($ref) in OpenAPI schemas are preserved rather than dereferenced
// during loading. This allows generators to produce references to other types
// rather than inlining definitions, resulting in cleaner generated code.
package spec

