package config

// Config represents the root configuration structure from typegen.jsonc.
//
// A configuration file defines which OpenAPI specs to load, how to select
// schemas from those specs, and where/how to generate the output files.
type Config struct {
	// Schema is the optional JSON Schema reference for IDE validation.
	// Typically points to the sparktype schema: "./schema/typegen.json"
	Schema string `json:"$schema,omitempty"`

	// Specs maps logical names to OpenAPI specification sources.
	// Each spec can be a local file, remote URL, or inline schema definitions.
	// The name is used in content patterns (e.g., "users:*" references the "users" spec).
	Specs map[string]Spec `json:"specs"`

	// Outputs defines the files to generate.
	// Each output specifies a path, format, and which schemas to include.
	Outputs []Output `json:"outputs"`

	// Options contains global generation settings applied to all outputs.
	// Individual outputs can override these with their own options.
	Options Options `json:"options,omitempty"`
}

// Spec represents a named OpenAPI specification source.
//
// Exactly one of Path, URL, or Schemas must be set. This allows specs to be
// loaded from local files, fetched from remote servers, or defined inline
// within the configuration file itself.
type Spec struct {
	// Path is a filesystem path to a local OpenAPI spec file.
	// Supports both YAML and JSON formats. Relative paths are resolved
	// from the directory containing typegen.jsonc.
	// Example: "./api/openapi.yaml"
	Path string `json:"path,omitempty"`

	// URL is an HTTP(S) URL to fetch the OpenAPI spec from.
	// The content type is auto-detected from the response or URL extension.
	// Example: "https://api.example.com/v1/openapi.json"
	URL string `json:"url,omitempty"`

	// Headers are custom HTTP headers to include when fetching a remote spec.
	// Useful for authentication (e.g., API keys, Bearer tokens).
	// Values support environment variable expansion: "${API_KEY}"
	Headers map[string]string `json:"headers,omitempty"`

	// Schemas contains inline schema definitions.
	// Each key is a schema name, and values are OpenAPI Schema Objects.
	// This mode is useful for simple schemas without full OpenAPI boilerplate.
	// Example: {"User": {"type": "object", "properties": {...}}}
	Schemas map[string]any `json:"schemas,omitempty"`
}

// Output represents a single output file configuration.
//
// Each output specifies what to generate (format), where to write it (path),
// and what to include (contents). Multiple outputs can target different
// formats or select different subsets of schemas.
type Output struct {
	// Path is the filesystem path where the generated file will be written.
	// Parent directories are created automatically if they don't exist.
	// Relative paths are resolved from the current working directory.
	Path string `json:"path"`

	// Format specifies the generator to use.
	// Supported values: "typescript", "zod", "python", "golang"
	Format string `json:"format"`

	// Contents defines which schemas to include and how to organize them.
	// Items can be schema patterns ("specName:glob") or namespace definitions.
	// Order is preserved in the generated output.
	Contents []ContentItem `json:"contents"`

	// Options are format-specific generation settings for this output.
	// These override global Options for this output only.
	// See each generator's documentation for supported options.
	Options map[string]any `json:"options,omitempty"`
}

// ContentItem represents either a schema pattern or a namespace definition.
//
// In JSON, a content item can be either:
//   - A string: Schema pattern like "users:*Request" or "api:User"
//   - An object: Namespace definition with nested contents
//
// The custom UnmarshalJSON in loader.go handles this polymorphism.
type ContentItem struct {
	// Pattern is a schema selection pattern in "specName:glob" format.
	// The spec name references a key in Config.Specs.
	// The glob pattern matches schema names (e.g., "*", "User*", "*Request").
	// Only set when this item represents a pattern (mutually exclusive with Namespace).
	Pattern string

	// Namespace is a namespace definition containing nested content items.
	// Namespaces allow hierarchical organization of generated types.
	// Only set when this item represents a namespace (mutually exclusive with Pattern).
	Namespace *NamespaceDef
}

// NamespaceDef represents a namespace containing schemas and/or nested namespaces.
//
// Namespaces provide hierarchical organization for generated types. Not all
// generators support namespaces; those that don't will return an error if
// namespaces are used in their outputs.
type NamespaceDef struct {
	// Name is the namespace identifier.
	// This becomes the namespace/module name in generated code.
	// Must be a valid identifier in the target language.
	Name string `json:"namespace"`

	// Contents defines the schemas and nested namespaces within this namespace.
	// Supports the same pattern and namespace syntax as Output.Contents.
	Contents []ContentItem `json:"contents"`
}

// Options contains global generation settings applied to all outputs.
//
// These settings can be overridden per-output using Output.Options.
type Options struct {
	// DereferenceRefs controls whether $ref references are dereferenced.
	// When true, referenced schemas are inlined rather than referenced by name.
	// Default: false (references produce type references)
	DereferenceRefs bool `json:"dereferenceRefs,omitempty"`

	// GenerateEnums controls whether string enums produce enum types.
	// When true, string enums become TypeScript enums or similar constructs.
	// When false, string enums become union types of literal strings.
	// Default: false (union types)
	GenerateEnums bool `json:"generateEnums,omitempty"`

	// NullableHandling controls how nullable fields are represented.
	// Supported values depend on the generator.
	// Default varies by generator.
	NullableHandling string `json:"nullableHandling,omitempty"`
}

// IsLocal returns true if the spec is loaded from a local file.
// Mutually exclusive with IsRemote() and IsInline().
func (s *Spec) IsLocal() bool {
	return s.Path != ""
}

// IsRemote returns true if the spec is loaded from a remote URL.
// Mutually exclusive with IsLocal() and IsInline().
func (s *Spec) IsRemote() bool {
	return s.URL != ""
}

// IsInline returns true if the spec contains inline schema definitions.
// Mutually exclusive with IsLocal() and IsRemote().
func (s *Spec) IsInline() bool {
	return len(s.Schemas) > 0
}

// GetSource returns a human-readable description of the spec source.
// Used in error messages and logging.
func (s *Spec) GetSource() string {
	if s.Path != "" {
		return s.Path
	}
	if s.URL != "" {
		return s.URL
	}
	return "inline"
}

// IsPattern returns true if this content item is a schema pattern.
// When true, Pattern is set and Namespace is nil.
func (c *ContentItem) IsPattern() bool {
	return c.Pattern != ""
}

// IsNamespace returns true if this content item is a namespace definition.
// When true, Namespace is set and Pattern is empty.
func (c *ContentItem) IsNamespace() bool {
	return c.Namespace != nil
}
