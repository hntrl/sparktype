package generators

import (
	"fmt"

	"github.com/hntrl/sparktype/internal/contents"
)

// Generator is the interface implemented by all type generators.
//
// Generators convert a resolved content tree into language-specific type definitions.
// The interface provides methods for both generating new output and comparing
// existing output against expected content.
type Generator interface {
	// Name returns the generator's registered name (e.g., "typescript", "zod").
	// This matches the format name used in configuration files.
	Name() string

	// Generate produces output content from a resolved tree.
	// Returns the complete file content as bytes, ready to write to disk.
	Generate(tree *contents.ResolvedTree, opts Options) ([]byte, error)

	// Compare compares existing file content against what would be generated.
	// Returns a CompareResult describing any differences between existing and expected.
	// Used by the 'check' command to detect drift.
	Compare(existing []byte, tree *contents.ResolvedTree, opts Options) (*CompareResult, error)
}

// registry is the global map of format names to generator implementations.
// Populated by generator packages calling Register() in their init() functions.
var registry = make(map[string]Generator)

// Register adds a generator to the global registry.
//
// This function should be called from a generator package's init() function
// to make the generator available for use. The CLI imports generator packages
// with blank imports to trigger registration:
//
//	import _ "github.com/hntrl/sparktype/internal/generators/typescript"
//
// Parameters:
//   - name: The format name used in configuration (e.g., "typescript", "zod")
//   - generate: Function that produces output from a content tree
//   - compare: Function that compares existing output to expected (may be nil)
func Register(name string, generate GenerateFunc, compare CompareFunc) {
	registry[name] = &generator{name: name, generate: generate, compare: compare}
}

// Get retrieves a generator by its format name.
//
// Returns an error if no generator is registered with the given name.
// This is typically called by the CLI after loading configuration to
// get the generator for each output's format.
func Get(format string) (Generator, error) {
	g, ok := registry[format]
	if !ok {
		return nil, fmt.Errorf("unknown format: %s", format)
	}
	return g, nil
}

// List returns the names of all registered generators.
//
// Useful for displaying available formats in help text or error messages.
// Note: Order is not guaranteed (map iteration order).
func List() []string {
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}
