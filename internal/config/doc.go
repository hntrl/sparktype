// Package config handles loading and validation of sparktype configuration files.
//
// The configuration is stored in a typegen.jsonc (or user specified) file, which
// uses JSON with comments (JSONC) format. This package provides types that map to the
// configuration structure and utilities for loading and validating them.
//
// # Configuration Structure
//
// A typegen.jsonc file has three main sections:
//
//   - specs: Named OpenAPI specification sources (local files, remote URLs, or inline schemas)
//   - outputs: Output file configurations with format and content selection
//   - options: Global generation options that apply to all outputs
//
// # Spec Sources
//
// Specs can be loaded from three sources:
//
//   - Local files: { "path": "./openapi.yaml" }
//   - Remote URLs: { "url": "https://api.example.com/openapi.json", "headers": {...} }
//   - Inline schemas: { "schemas": { "User": { "type": "object", ... } } }
//
// # Content Selection
//
// The contents array in each output uses a tree-based structure supporting:
//
//   - Schema patterns: "specName:PatternGlob" (e.g., "users:*Request")
//   - Namespaces: { "namespace": "Models", "contents": [...] }
//
// Namespaces can be nested to create hierarchical type organization.
//
// # Environment Variables
//
// The loader expands environment variables in string values using ${VAR} syntax,
// allowing sensitive data like API keys to be kept out of configuration files.
package config
