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
//  6. Auto-including any referenced schemas not explicitly included (dependencies)
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

	// Auto-include any referenced schemas that aren't explicitly included
	if err := r.resolveDependencies(tree); err != nil {
		return nil, err
	}

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

// refInfo tracks information about a referenced schema for dependency resolution.
type refInfo struct {
	sourceSpec    string
	namespacePath []string
}

// resolveDependencies finds and adds any schemas referenced by the tree's schemas
// but not explicitly included. This ensures the generated output is complete.
//
// Dependencies are added to the same namespace as the schema that first references them.
// The process repeats until no new dependencies are found (handling transitive deps).
func (r *Resolver) resolveDependencies(tree *ResolvedTree) error {
	for {
		// Build set of currently included schema names
		included := make(map[string]bool)
		for _, schema := range tree.CollectAllSchemas() {
			included[schema.Name] = true
		}

		// Find all references and their referencing schema's namespace
		// Map: refName -> refInfo
		missingRefs := make(map[string]refInfo)

		collectMissingRefs(tree.Nodes, []string{}, included, missingRefs)

		// If no missing references, we're done
		if len(missingRefs) == 0 {
			break
		}

		// Load and add missing schemas
		for refName, info := range missingRefs {
			schema, err := r.registry.GetSchema(info.sourceSpec, refName)
			if err != nil {
				// Schema not found in spec - skip it (will remain as unresolved reference)
				continue
			}

			// Add schema to the appropriate namespace
			if err := addSchemaToNamespace(tree, schema, info.namespacePath); err != nil {
				return err
			}
		}
	}

	return nil
}

// collectMissingRefs finds all schema references that aren't in the included set.
// It records the source spec and namespace path for each missing reference.
func collectMissingRefs(nodes []Node, namespacePath []string, included map[string]bool, missingRefs map[string]refInfo) {
	for _, node := range nodes {
		if node.IsSchema() {
			collectSchemaRefs(*node.Schema, namespacePath, included, missingRefs)
		} else if node.IsNamespace() {
			childPath := append([]string{}, namespacePath...)
			childPath = append(childPath, node.Namespace.Name)
			collectMissingRefs(node.Namespace.Children, childPath, included, missingRefs)
		}
	}
}

// collectSchemaRefs finds all references within a schema and adds missing ones to the map.
func collectSchemaRefs(schema spec.Schema, namespacePath []string, included map[string]bool, missingRefs map[string]refInfo) {
	// Check direct ref
	if schema.Ref != "" {
		refName := extractRefName(schema.Ref)
		if !included[refName] {
			if _, exists := missingRefs[refName]; !exists {
				missingRefs[refName] = refInfo{schema.SourceSpec, namespacePath}
			}
		}
	}

	// Check properties
	for _, prop := range schema.Properties {
		collectSchemaRefs(prop.Schema, namespacePath, included, missingRefs)
	}

	// Check array items
	if schema.Items != nil {
		collectSchemaRefs(*schema.Items, namespacePath, included, missingRefs)
	}

	// Check composition types
	for _, s := range schema.AllOf {
		collectSchemaRefs(s, namespacePath, included, missingRefs)
	}
	for _, s := range schema.OneOf {
		collectSchemaRefs(s, namespacePath, included, missingRefs)
	}
	for _, s := range schema.AnyOf {
		collectSchemaRefs(s, namespacePath, included, missingRefs)
	}
}

// extractRefName extracts the schema name from a $ref path.
// Transforms: "#/components/schemas/User" -> "User"
func extractRefName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}

// addSchemaToNamespace adds a schema to the specified namespace path in the tree.
// If namespacePath is empty, adds to root level.
func addSchemaToNamespace(tree *ResolvedTree, schema spec.Schema, namespacePath []string) error {
	// Track the namespace for this schema
	tree.SchemaNamespaces[schema.Name] = namespacePath

	if len(namespacePath) == 0 {
		// Add to root level
		tree.Nodes = append(tree.Nodes, Node{Schema: &schema})
		return nil
	}

	// Find or create the namespace path
	nodes := &tree.Nodes
	for _, nsName := range namespacePath {
		found := false
		for i := range *nodes {
			if (*nodes)[i].IsNamespace() && (*nodes)[i].Namespace.Name == nsName {
				nodes = &(*nodes)[i].Namespace.Children
				found = true
				break
			}
		}
		if !found {
			// This shouldn't happen if dependencies come from schemas in existing namespaces
			// but handle it gracefully by creating the namespace
			newNs := &NamespaceNode{Name: nsName, Children: []Node{}}
			*nodes = append(*nodes, Node{Namespace: newNs})
			nodes = &newNs.Children
		}
	}

	*nodes = append(*nodes, Node{Schema: &schema})
	return nil
}
