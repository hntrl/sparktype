package generators

import (
	"fmt"

	"github.com/hntrl/sparktype/internal/config"
	"github.com/hntrl/sparktype/internal/contents"
)

// DiffStatus represents the status of a compared item (type or property).
//
// Used in CompareResult to indicate whether types or properties were
// added, removed, changed, or remain unchanged between existing and expected output.
type DiffStatus int

const (
	// Unchanged indicates the item exists in both and is identical.
	Unchanged DiffStatus = iota

	// Added indicates the item exists in expected but not in existing.
	// For the 'check' command, this means a new type/property needs to be generated.
	Added

	// Removed indicates the item exists in existing but not in expected.
	// For the 'check' command, this means an obsolete type/property should be removed.
	Removed

	// Changed indicates the item exists in both but differs.
	// TypeDiff.Properties provides details on what changed within the type.
	Changed
)

// String returns a human-readable representation of the diff status.
func (s DiffStatus) String() string {
	switch s {
	case Unchanged:
		return "unchanged"
	case Added:
		return "added"
	case Removed:
		return "removed"
	case Changed:
		return "changed"
	default:
		return "unknown"
	}
}

// PropertyDiff represents a difference in a single property or field.
//
// When comparing types, each property that differs produces a PropertyDiff
// describing what changed. For the 'check' command output, these are displayed
// as nested items under the parent TypeDiff.
type PropertyDiff struct {
	// Name is the property/field name that differs.
	Name string

	// Status indicates whether the property was added, removed, or changed.
	Status DiffStatus

	// OldValue is the serialized representation of the existing value.
	// Empty for Added properties.
	OldValue string

	// NewValue is the serialized representation of the expected value.
	// Empty for Removed properties.
	NewValue string
}

// TypeDiff represents a difference in a type definition.
//
// Each type that differs between existing and expected output produces a TypeDiff.
// For types with Status == Changed, the Properties slice details what changed
// within the type (added/removed/changed properties).
type TypeDiff struct {
	// Name is the type name, potentially namespace-qualified (e.g., "Models.User").
	Name string

	// Status indicates whether the type was added, removed, or changed.
	Status DiffStatus

	// Properties contains property-level diffs for Changed types.
	// Empty for Added/Removed types (the whole type is new/gone).
	Properties []PropertyDiff
}

// CompareResult contains the result of comparing existing output against expected.
//
// Used by the 'check' command to determine if generated files are in sync
// with the current OpenAPI specifications.
type CompareResult struct {
	// Match is true if existing and expected outputs are semantically equivalent.
	// When true, Types will be empty or contain only Unchanged entries.
	Match bool

	// Types contains diffs for each type that differs.
	// Only types with Status != Unchanged are included.
	// Types are sorted alphabetically by name for consistent output.
	Types []TypeDiff
}

// Options contains generation options from both global config and output-specific settings.
//
// Generators receive Options to customize their output. Output-specific options
// take precedence over global options, allowing per-file customization while
// maintaining project-wide defaults.
type Options struct {
	// GlobalOptions contains settings from the config's top-level options.
	// These apply to all outputs unless overridden.
	GlobalOptions config.Options

	// OutputOptions contains settings from a specific output's options map.
	// These override GlobalOptions when the same key is present.
	OutputOptions map[string]any
}

// GetOption retrieves an option value by key with a fallback default.
//
// Checks OutputOptions first, then falls back to defaultValue.
// Note: GlobalOptions fields should be accessed directly (e.g., opts.GlobalOptions.GenerateEnums).
func (o Options) GetOption(key string, defaultValue any) any {
	if o.OutputOptions != nil {
		if val, ok := o.OutputOptions[key]; ok {
			return val
		}
	}
	return defaultValue
}

// GetBoolOption retrieves a boolean option with type-safe fallback.
//
// If the key exists and is a bool, returns that value.
// Otherwise returns defaultValue.
func (o Options) GetBoolOption(key string, defaultValue bool) bool {
	val := o.GetOption(key, defaultValue)
	if b, ok := val.(bool); ok {
		return b
	}
	return defaultValue
}

// GetStringOption retrieves a string option with type-safe fallback.
//
// If the key exists and is a string, returns that value.
// Otherwise returns defaultValue.
func (o Options) GetStringOption(key string, defaultValue string) string {
	val := o.GetOption(key, defaultValue)
	if s, ok := val.(string); ok {
		return s
	}
	return defaultValue
}

// GenerateFunc is the function signature for generator implementations.
//
// Generators receive a resolved content tree and options, and produce
// the generated file content as bytes. Errors should include context
// about what failed (e.g., which schema caused the issue).
type GenerateFunc func(tree *contents.ResolvedTree, opts Options) ([]byte, error)

// CompareFunc is the function signature for comparison implementations.
//
// Comparers receive existing file content, the expected content tree, and options.
// They return a CompareResult describing any differences. Errors typically
// indicate parsing failures in the existing file.
type CompareFunc func(existing []byte, tree *contents.ResolvedTree, opts Options) (*CompareResult, error)

// generator wraps a name and generate/compare functions into a unified interface.
//
// This internal type is used by the registry to store registered generators.
// External code interacts with generators through the Get() function and
// the returned interface methods.
type generator struct {
	name     string
	generate GenerateFunc
	compare  CompareFunc
}

// Name returns the generator's registered name (e.g., "typescript", "zod").
func (g *generator) Name() string {
	return g.name
}

// Generate produces output content from a resolved tree.
//
// Delegates to the registered GenerateFunc. The returned bytes are the
// complete file content ready to be written to disk.
func (g *generator) Generate(tree *contents.ResolvedTree, opts Options) ([]byte, error) {
	return g.generate(tree, opts)
}

// Compare compares existing file content against expected output.
//
// Delegates to the registered CompareFunc. Returns an error if comparison
// is not supported by this generator (compare function was not registered).
func (g *generator) Compare(existing []byte, tree *contents.ResolvedTree, opts Options) (*CompareResult, error) {
	if g.compare == nil {
		return nil, fmt.Errorf("generator %s does not support comparison", g.name)
	}
	return g.compare(existing, tree, opts)
}
