package contents

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"
	"github.com/hntrl/sparktype/internal/config"
	"github.com/hntrl/sparktype/internal/spec"
)

// Node represents a single node in the resolved content tree.
//
// A node is either a schema (leaf node) or a namespace (branch node containing
// other nodes). This discriminated union is determined by which field is set:
//   - Schema != nil: This is a schema node
//   - Namespace != nil: This is a namespace node
//
// Exactly one of these fields should be non-nil.
type Node struct {
	// Schema holds the resolved schema when this node represents a type definition.
	// When set, Namespace must be nil.
	Schema *spec.Schema

	// Namespace holds the namespace definition when this node represents a container.
	// When set, Schema must be nil.
	Namespace *NamespaceNode
}

// NamespaceNode represents a namespace containing schemas and nested namespaces.
//
// Namespaces provide hierarchical organization for generated types. They map
// to language-specific constructs like TypeScript namespaces or nested objects.
type NamespaceNode struct {
	// Name is the namespace identifier used in generated code.
	Name string

	// Children contains the schemas and nested namespaces within this namespace.
	// Order is preserved from the configuration.
	Children []Node
}

// IsSchema returns true if this node contains a schema definition.
// When true, node.Schema is safe to dereference.
func (n *Node) IsSchema() bool {
	return n.Schema != nil
}

// IsNamespace returns true if this node contains a namespace.
// When true, node.Namespace is safe to dereference.
func (n *Node) IsNamespace() bool {
	return n.Namespace != nil
}

// ResolvedTree represents a fully resolved content tree ready for code generation.
//
// The tree contains all schemas selected by the configuration's content patterns,
// organized into the namespace hierarchy specified in the configuration. It also
// tracks namespace paths for each schema to enable correct cross-reference resolution.
type ResolvedTree struct {
	// Nodes contains the top-level content nodes (schemas and namespaces).
	// Order is preserved from the configuration.
	Nodes []Node

	// SchemaNamespaces maps each schema name to its namespace path.
	// The path is a slice of namespace names from root to the schema's parent.
	// For root-level schemas, the path is empty.
	// Example: "User" in "Models.Auth" namespace has path ["Models", "Auth"]
	SchemaNamespaces map[string][]string
}

// Resolver resolves configuration content items into a ResolvedTree.
//
// The resolver uses a spec.Registry to load schemas on demand and applies
// glob patterns to filter schemas by name.
type Resolver struct {
	registry *spec.Registry
}

// NewResolver creates a new content resolver using the given spec registry.
//
// The registry provides access to named specs and their schemas. The resolver
// will call registry.GetSchemas() to load schemas when resolving patterns.
func NewResolver(registry *spec.Registry) *Resolver {
	return &Resolver{registry: registry}
}

// Resolve converts configuration content items into a resolved tree.
//
// Resolution involves:
//  1. Parsing each "spec:pattern" string to identify the source spec and filter
//  2. Loading schemas from the spec registry
//  3. Filtering schemas using glob pattern matching
//  4. Building the hierarchical tree structure with namespaces
//  5. Recording namespace paths for each schema
//
// Returns an error if any pattern is malformed, references an unknown spec,
// or uses an invalid glob pattern.
func (r *Resolver) Resolve(contents []config.ContentItem) (*ResolvedTree, error) {
	tree := &ResolvedTree{
		SchemaNamespaces: make(map[string][]string),
	}

	nodes, err := r.resolveContents(contents, []string{}, tree.SchemaNamespaces)
	if err != nil {
		return nil, err
	}

	tree.Nodes = nodes
	return tree, nil
}

// resolveContents recursively resolves content items into nodes.
//
// Parameters:
//   - contents: The content items to resolve (patterns and namespace definitions)
//   - namespacePath: The current namespace path for tracking schema locations
//   - schemaNamespaces: Map to populate with schema name -> namespace path mappings
//
// The namespacePath grows as we descend into nested namespaces, allowing us to
// track where each schema is located in the hierarchy.
func (r *Resolver) resolveContents(contents []config.ContentItem, namespacePath []string, schemaNamespaces map[string][]string) ([]Node, error) {
	var nodes []Node

	for _, item := range contents {
		if item.IsPattern() {
			// Resolve pattern to schemas
			schemas, err := r.resolvePattern(item.Pattern)
			if err != nil {
				return nil, err
			}

			for _, schema := range schemas {
				// Track the namespace path for this schema
				schemaNamespaces[schema.Name] = namespacePath

				// Create a copy to avoid modifying the original
				schemaCopy := schema
				nodes = append(nodes, Node{Schema: &schemaCopy})
			}
		} else if item.IsNamespace() {
			// Recursively resolve namespace contents
			childPath := append(namespacePath, item.Namespace.Name)
			children, err := r.resolveContents(item.Namespace.Contents, childPath, schemaNamespaces)
			if err != nil {
				return nil, err
			}

			nodes = append(nodes, Node{
				Namespace: &NamespaceNode{
					Name:     item.Namespace.Name,
					Children: children,
				},
			})
		}
	}

	return nodes, nil
}

// resolvePattern resolves a "spec:pattern" string to matching schemas.
//
// The pattern format is "specName:globPattern" where:
//   - specName: References a key in the configuration's specs map
//   - globPattern: A glob pattern to filter schema names (e.g., "*", "User*", "*Request")
//
// The special pattern "*" matches all schemas without glob compilation.
func (r *Resolver) resolvePattern(pattern string) ([]spec.Schema, error) {
	// Split pattern into spec name and schema pattern
	parts := strings.SplitN(pattern, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid pattern %q: must be in format 'spec:pattern'", pattern)
	}

	specName := parts[0]
	schemaPattern := parts[1]

	// Get schemas from the spec
	schemas, err := r.registry.GetSchemas(specName)
	if err != nil {
		return nil, fmt.Errorf("failed to get schemas for pattern %q: %w", pattern, err)
	}

	// If pattern is "*", return all schemas
	if schemaPattern == "*" {
		return schemas, nil
	}

	// Compile glob pattern
	g, err := glob.Compile(schemaPattern)
	if err != nil {
		return nil, fmt.Errorf("invalid glob pattern %q: %w", schemaPattern, err)
	}

	// Filter schemas by pattern
	var matched []spec.Schema
	for _, schema := range schemas {
		if g.Match(schema.Name) {
			matched = append(matched, schema)
		}
	}

	return matched, nil
}

// GetNamespacePath returns the fully qualified path for a schema.
//
// For a schema "User" in namespace ["Models"], returns "Models.User".
// For a root-level schema, returns just the schema name.
//
// This is useful for generating imports or fully qualified type names.
func (t *ResolvedTree) GetNamespacePath(schemaName string) string {
	path, ok := t.SchemaNamespaces[schemaName]
	if !ok || len(path) == 0 {
		return schemaName
	}
	return strings.Join(append(path, schemaName), ".")
}

// GetRelativeRef returns the reference path from one namespace to a schema in another.
//
// This method determines how to reference a schema from a different namespace context.
// It's used by generators to produce correct cross-references in generated code.
//
// Parameters:
//   - fromPath: The namespace path of the schema making the reference
//   - toSchemaName: The name of the schema being referenced
//
// Returns:
//   - If both are in the same namespace: Just the schema name
//   - If the target is in a different namespace: Fully qualified path (e.g., "Models.User")
//   - If the target is at root level: Just the schema name
//   - If the target isn't in the tree: Just the schema name (external reference)
func (t *ResolvedTree) GetRelativeRef(fromPath []string, toSchemaName string) string {
	toPath, ok := t.SchemaNamespaces[toSchemaName]
	if !ok {
		// Schema not in tree, just return the name
		return toSchemaName
	}

	// If same namespace path, just return the name
	if pathEqual(fromPath, toPath) {
		return toSchemaName
	}

	// Return fully qualified path from root
	if len(toPath) == 0 {
		return toSchemaName
	}
	return strings.Join(append(toPath, toSchemaName), ".")
}

// pathEqual compares two namespace paths for equality.
func pathEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// CollectAllSchemas returns all schemas in the tree as a flat slice.
//
// This traverses the entire tree (including nested namespaces) and collects
// all schema nodes. Order follows a depth-first traversal.
//
// Useful for operations that need to process all schemas regardless of their
// namespace organization, such as building reference lookup tables.
func (t *ResolvedTree) CollectAllSchemas() []spec.Schema {
	var schemas []spec.Schema
	collectSchemas(t.Nodes, &schemas)
	return schemas
}

// collectSchemas recursively collects schemas from a node slice.
func collectSchemas(nodes []Node, schemas *[]spec.Schema) {
	for _, node := range nodes {
		if node.IsSchema() {
			*schemas = append(*schemas, *node.Schema)
		} else if node.IsNamespace() {
			collectSchemas(node.Namespace.Children, schemas)
		}
	}
}
