package spec

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/getkin/kin-openapi/openapi3"

	"github.com/hntrl/sparktype/internal/config"
)

// LoadSpec loads an OpenAPI spec from the given configuration.
//
// Parameters:
//   - specConfig: The spec configuration (local, remote, or inline)
//   - basePath: Base directory for resolving relative paths
//   - specName: Logical name for this spec (used in error messages and schema metadata)
func LoadSpec(specConfig config.Spec, basePath string, specName string) ([]Schema, error) {
	// Handle inline schemas
	if specConfig.IsInline() {
		return loadInlineSchemas(specConfig.Schemas, specName)
	}

	var doc *openapi3.T
	var err error

	if specConfig.IsLocal() {
		doc, err = loadLocalSpec(specConfig.Path, basePath)
	} else if specConfig.IsRemote() {
		doc, err = loadRemoteSpec(specConfig.URL, specConfig.Headers)
	} else {
		return nil, fmt.Errorf("spec %q must have path, url, or schemas defined", specName)
	}

	if err != nil {
		return nil, fmt.Errorf("loading spec %q: %w", specName, err)
	}

	// Extract schemas
	return extractSchemas(doc, specName)
}

// loadInlineSchemas converts inline schema definitions to Schema structs
func loadInlineSchemas(inlineSchemas map[string]any, specName string) ([]Schema, error) {
	// Convert inline schemas to OpenAPI format and parse
	// We wrap them in a minimal OpenAPI structure
	spec := map[string]any{
		"openapi": "3.0.0",
		"info": map[string]any{
			"title":   specName,
			"version": "1.0.0",
		},
		"paths": map[string]any{},
		"components": map[string]any{
			"schemas": inlineSchemas,
		},
	}

	// Convert to JSON
	jsonData, err := json.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal inline schemas: %w", err)
	}

	// Parse with OpenAPI loader
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse inline schemas: %w", err)
	}

	return extractSchemas(doc, specName)
}

func loadLocalSpec(path string, basePath string) (*openapi3.T, error) {
	// Resolve relative path
	if !filepath.IsAbs(path) {
		path = filepath.Join(basePath, path)
	}

	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		return nil, fmt.Errorf("spec file not found: %s", path)
	}

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}

	return doc, nil
}

func loadRemoteSpec(url string, headers map[string]string) (*openapi3.T, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: HTTPClientTimeout,
	}

	// Create request
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set default Accept header if not specified
	if req.Header.Get("Accept") == "" {
		req.Header.Set("Accept", DefaultAcceptHeader)
	}

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spec from %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch spec from %s: HTTP %d", url, resp.StatusCode)
	}

	// Read body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Parse spec (kin-openapi handles both JSON and YAML automatically)
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	doc, err := loader.LoadFromData(body)

	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec from %s: %w", url, err)
	}

	return doc, nil
}

func extractSchemas(doc *openapi3.T, specName string) ([]Schema, error) {
	if doc.Components == nil || doc.Components.Schemas == nil {
		return []Schema{}, nil
	}

	var schemas []Schema
	schemaNames := make([]string, 0, len(doc.Components.Schemas))
	for name := range doc.Components.Schemas {
		schemaNames = append(schemaNames, name)
	}
	sort.Strings(schemaNames)

	for _, name := range schemaNames {
		schemaRef := doc.Components.Schemas[name]
		if schemaRef == nil || schemaRef.Value == nil {
			continue
		}

		schema := convertSchema(name, schemaRef.Value, specName)
		schemas = append(schemas, schema)
	}

	return schemas, nil
}

func convertSchema(name string, s *openapi3.Schema, specName string) Schema {
	// Get schema type safely
	schemaType := ""
	if types := s.Type.Slice(); len(types) > 0 {
		schemaType = types[0]
	}

	schema := Schema{
		Name:        name,
		Description: s.Description,
		Type:        schemaType,
		Format:      s.Format,
		Nullable:    s.Nullable,
		ReadOnly:    s.ReadOnly,
		WriteOnly:   s.WriteOnly,
		Default:     s.Default,
		Example:     s.Example,
		Pattern:     s.Pattern,
		UniqueItems: s.UniqueItems,
		SourceSpec:  specName,
	}

	// Handle numeric constraints
	if s.Min != nil {
		schema.Minimum = s.Min
	}
	if s.Max != nil {
		schema.Maximum = s.Max
	}
	if s.MinLength != 0 {
		minLen := int(s.MinLength)
		schema.MinLength = &minLen
	}
	if s.MaxLength != nil {
		maxLen := int(*s.MaxLength)
		schema.MaxLength = &maxLen
	}
	if s.MinItems != 0 {
		minItems := int(s.MinItems)
		schema.MinItems = &minItems
	}
	if s.MaxItems != nil {
		maxItems := int(*s.MaxItems)
		schema.MaxItems = &maxItems
	}

	// Handle enum
	if len(s.Enum) > 0 {
		schema.Enum = s.Enum
	}

	// Handle properties
	if len(s.Properties) > 0 {
		requiredSet := make(map[string]bool)
		for _, r := range s.Required {
			requiredSet[r] = true
		}

		propNames := make([]string, 0, len(s.Properties))
		for propName := range s.Properties {
			propNames = append(propNames, propName)
		}
		sort.Strings(propNames)

		for _, propName := range propNames {
			propRef := s.Properties[propName]
			if propRef == nil {
				continue
			}

			var propSchema Schema
			if propRef.Ref != "" {
				// Property is a reference - extract ref name
				propSchema = Schema{
					Ref:        propRef.Ref,
					SourceSpec: specName,
				}
				// Also copy description from value if available
				if propRef.Value != nil {
					propSchema.Description = propRef.Value.Description
				}
			} else if propRef.Value != nil {
				propSchema = convertSchema(propName, propRef.Value, specName)
			} else {
				continue
			}

			schema.Properties = append(schema.Properties, Property{
				Name:     propName,
				Schema:   propSchema,
				Required: requiredSet[propName],
			})
		}
		schema.Required = s.Required
	}

	// Handle array items
	if s.Items != nil {
		var itemSchema Schema
		if s.Items.Ref != "" {
			itemSchema = Schema{Ref: s.Items.Ref, SourceSpec: specName}
		} else if s.Items.Value != nil {
			itemSchema = convertSchema("", s.Items.Value, specName)
		}
		schema.Items = &itemSchema
	}

	// Handle composition
	for _, allOf := range s.AllOf {
		if allOf.Value != nil {
			schema.AllOf = append(schema.AllOf, convertSchema("", allOf.Value, specName))
		} else if allOf.Ref != "" {
			schema.AllOf = append(schema.AllOf, Schema{Ref: allOf.Ref})
		}
	}

	for _, oneOf := range s.OneOf {
		if oneOf.Value != nil {
			schema.OneOf = append(schema.OneOf, convertSchema("", oneOf.Value, specName))
		} else if oneOf.Ref != "" {
			schema.OneOf = append(schema.OneOf, Schema{Ref: oneOf.Ref})
		}
	}

	for _, anyOf := range s.AnyOf {
		if anyOf.Value != nil {
			schema.AnyOf = append(schema.AnyOf, convertSchema("", anyOf.Value, specName))
		} else if anyOf.Ref != "" {
			schema.AnyOf = append(schema.AnyOf, Schema{Ref: anyOf.Ref})
		}
	}

	return schema
}
