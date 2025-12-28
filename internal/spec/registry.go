package spec

import (
	"fmt"
	"sync"

	"github.com/hntrl/sparktype/internal/config"
)

// Registry manages named specs with lazy loading
type Registry struct {
	specs    map[string]config.Spec
	loaded   map[string][]Schema
	basePath string
	mu       sync.RWMutex
}

// NewRegistry creates a new spec registry
func NewRegistry(specs map[string]config.Spec, basePath string) *Registry {
	return &Registry{
		specs:    specs,
		loaded:   make(map[string][]Schema),
		basePath: basePath,
	}
}

// GetSchemas returns the schemas for a named spec, loading if necessary
// Returns a deep copy to prevent modification of cached schemas
func (r *Registry) GetSchemas(name string) ([]Schema, error) {
	r.mu.RLock()
	cached, ok := r.loaded[name]
	r.mu.RUnlock()

	if !ok {
		// Need to load the spec
		r.mu.Lock()

		// Double-check after acquiring write lock
		if cached, ok = r.loaded[name]; !ok {
			specConfig, specOk := r.specs[name]
			if !specOk {
				r.mu.Unlock()
				return nil, fmt.Errorf("spec %q not found in configuration", name)
			}

			var err error
			cached, err = LoadSpec(specConfig, r.basePath, name)
			if err != nil {
				r.mu.Unlock()
				return nil, fmt.Errorf("failed to load spec %q: %w", name, err)
			}

			r.loaded[name] = cached
		}
		r.mu.Unlock()
	}

	// Return deep copy to prevent modification of cached schemas
	return deepCopySchemas(cached), nil
}

// deepCopySchemas creates a deep copy of a schema slice
func deepCopySchemas(schemas []Schema) []Schema {
	result := make([]Schema, len(schemas))
	for i, s := range schemas {
		result[i] = deepCopySchema(s)
	}
	return result
}

// deepCopySchema creates a deep copy of a schema
func deepCopySchema(s Schema) Schema {
	result := s // Copy basic fields

	// Deep copy properties
	if len(s.Properties) > 0 {
		result.Properties = make([]Property, len(s.Properties))
		for i, p := range s.Properties {
			result.Properties[i] = Property{
				Name:     p.Name,
				Schema:   deepCopySchema(p.Schema),
				Required: p.Required,
			}
		}
	}

	// Deep copy required slice
	if len(s.Required) > 0 {
		result.Required = make([]string, len(s.Required))
		copy(result.Required, s.Required)
	}

	// Deep copy items
	if s.Items != nil {
		copied := deepCopySchema(*s.Items)
		result.Items = &copied
	}

	// Deep copy enum
	if len(s.Enum) > 0 {
		result.Enum = make([]any, len(s.Enum))
		copy(result.Enum, s.Enum)
	}

	// Deep copy composition
	if len(s.AllOf) > 0 {
		result.AllOf = make([]Schema, len(s.AllOf))
		for i, schema := range s.AllOf {
			result.AllOf[i] = deepCopySchema(schema)
		}
	}
	if len(s.OneOf) > 0 {
		result.OneOf = make([]Schema, len(s.OneOf))
		for i, schema := range s.OneOf {
			result.OneOf[i] = deepCopySchema(schema)
		}
	}
	if len(s.AnyOf) > 0 {
		result.AnyOf = make([]Schema, len(s.AnyOf))
		for i, schema := range s.AnyOf {
			result.AnyOf[i] = deepCopySchema(schema)
		}
	}

	return result
}

// GetAllSpecNames returns all configured spec names
func (r *Registry) GetAllSpecNames() []string {
	names := make([]string, 0, len(r.specs))
	for name := range r.specs {
		names = append(names, name)
	}
	return names
}
