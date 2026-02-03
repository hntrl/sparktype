package typescript

import (
	"fmt"
	"strings"

	"github.com/hntrl/sparktype/internal/contents"
	"github.com/hntrl/sparktype/internal/generators"
	"github.com/hntrl/sparktype/internal/spec"
)

func init() {
	generators.Register("typescript", Generate, Compare)
}

// BuildContext holds state during AST construction.
//
// The context is passed through all Build* functions to provide access to:
//   - The resolved tree for cross-reference resolution
//   - Generation options that affect output style
type BuildContext struct {
	// tree is the resolved content tree, used for GetRelativeRef lookups.
	tree *contents.ResolvedTree

	// exportType controls object declaration style: "interface" or "type".
	exportType string

	// readonlyProps adds readonly modifier to all interface properties.
	readonlyProps bool

	// generateEnums produces TypeScript enums instead of union types for string enums.
	generateEnums bool
}

// Generate produces TypeScript type definitions from a resolved content tree.
//
// This is the main entry point called by the generator registry. It builds
// a File AST and serializes it to produce the final output.
func Generate(tree *contents.ResolvedTree, opts generators.Options) ([]byte, error) {
	file := BuildFile(tree, opts)
	return []byte(file.Serialize()), nil
}

// BuildFile constructs a complete File AST from a resolved content tree.
//
// The File contains:
//   - A header comment indicating the file was generated
//   - All type definitions organized according to the tree structure
//
// This function is also used by Compare to build the expected AST.
func BuildFile(tree *contents.ResolvedTree, opts generators.Options) *File {
	ctx := &BuildContext{
		tree:          tree,
		exportType:    opts.GetStringOption("exportType", "interface"),
		readonlyProps: opts.GetBoolOption("readonlyProperties", false),
		generateEnums: opts.GlobalOptions.GenerateEnums,
	}
	return &File{
		Header: "// " + generators.GeneratedFileHeader + "\n",
		Nodes:  BuildNodes(tree.Nodes, ctx, []string{}),
	}
}

// BuildNodes converts a slice of content nodes into AST nodes.
//
// This function handles both schema nodes (producing interfaces, type aliases, or enums)
// and namespace nodes (producing namespace blocks with nested content).
//
// Parameters:
//   - nodes: Content nodes to convert
//   - ctx: Build context with options and tree reference
//   - namespacePath: Current namespace path for cross-reference resolution
func BuildNodes(nodes []contents.Node, ctx *BuildContext, namespacePath []string) []Node {
	var result []Node

	for _, node := range nodes {
		if node.IsSchema() {
			result = append(result, BuildSchemaNode(*node.Schema, ctx, namespacePath))
		} else if node.IsNamespace() {
			result = append(result, buildNamespace(node.Namespace, ctx, namespacePath))
		}
	}

	return result
}

// buildNamespace converts a namespace content node into a Namespace AST node.
//
// Recursively builds child nodes, passing the updated namespace path to enable
// correct cross-reference resolution for nested schemas.
func buildNamespace(ns *contents.NamespaceNode, ctx *BuildContext, parentPath []string) Node {
	currentPath := append(parentPath, ns.Name)
	return &Namespace{
		Name:     ns.Name,
		Children: BuildNodes(ns.Children, ctx, currentPath),
	}
}

// BuildSchemaNode converts a schema into the appropriate AST node type.
//
// The node type is determined by the schema's characteristics:
//   - Enum schemas → Enum or TypeAlias (union of literals)
//   - Object schemas → Interface
//   - Other schemas → TypeAlias
func BuildSchemaNode(schema spec.Schema, ctx *BuildContext, namespacePath []string) Node {
	// Handle enums
	if schema.IsEnum() {
		return BuildEnum(schema, ctx)
	}

	// Handle object types
	if schema.IsObject() {
		return BuildInterface(schema, ctx, namespacePath)
	}

	// Handle type aliases for primitives
	return BuildTypeAlias(schema, ctx, namespacePath)
}

// BuildEnum converts an enum schema into either an Enum or TypeAlias node.
//
// When generateEnums is true and all values are strings, produces a TypeScript enum:
//
//	export enum Status { Active = "active", Inactive = "inactive" }
//
// Otherwise produces a union of literal types:
//
//	export type Status = "active" | "inactive";
func BuildEnum(schema spec.Schema, ctx *BuildContext) Node {
	if ctx.generateEnums && allStrings(schema.Enum) {
		// Generate TypeScript enum
		members := make([]EnumMember, len(schema.Enum))
		for i, val := range schema.Enum {
			strVal := fmt.Sprintf("%v", val)
			members[i] = EnumMember{
				Key:   toEnumKey(strVal),
				Value: strVal,
			}
		}
		return &Enum{
			Name:        schema.Name,
			Description: Description(schema.Description),
			Members:     members,
		}
	}

	// Generate union type
	types := make([]TypeExpr, len(schema.Enum))
	for i, val := range schema.Enum {
		types[i] = LiteralType{Value: val}
	}
	return &TypeAlias{
		Name:        schema.Name,
		Description: Description(schema.Description),
		Type:        UnionType{Types: types},
	}
}

// BuildInterface converts an object schema into an Interface AST node.
//
// The interface includes:
//   - JSDoc comment from schema description
//   - All properties with their types, optionality, and descriptions
//   - readonly modifiers based on schema or global option
func BuildInterface(schema spec.Schema, ctx *BuildContext, namespacePath []string) Node {
	// Build required set for quick lookup
	requiredSet := make(map[string]bool)
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	properties := make([]Property, len(schema.Properties))
	for i, prop := range schema.Properties {
		isRequired := prop.Required || requiredSet[prop.Name]
		properties[i] = Property{
			Name:        prop.Name,
			Type:        BuildTypeExpr(prop.Schema, ctx, namespacePath),
			Optional:    !isRequired,
			ReadOnly:    ctx.readonlyProps || prop.Schema.ReadOnly,
			Nullable:    prop.Schema.Nullable,
			Description: Description(prop.Schema.Description),
		}
	}

	return &Interface{
		Name:        schema.Name,
		Description: Description(schema.Description),
		Properties:  properties,
		ExportStyle: ctx.exportType,
	}
}

// BuildTypeAlias converts a non-object schema into a TypeAlias AST node.
//
// Used for primitive types, arrays, unions, and other non-object schemas
// that can't be represented as interfaces.
func BuildTypeAlias(schema spec.Schema, ctx *BuildContext, namespacePath []string) Node {
	typeExpr := BuildTypeExpr(schema, ctx, namespacePath)
	if schema.Nullable {
		typeExpr = UnionType{Types: []TypeExpr{typeExpr, NullType{}}}
	}
	return &TypeAlias{
		Name:        schema.Name,
		Description: Description(schema.Description),
		Type:        typeExpr,
	}
}

// BuildTypeExpr converts a schema into a TypeExpr for use within type definitions.
//
// This handles all type expression forms:
//   - References ($ref) → ReferenceType with namespace-aware path
//   - Composition (allOf) → IntersectionType
//   - Composition (oneOf/anyOf) → UnionType
//   - Arrays → ArrayType
//   - Objects without properties → RecordType
//   - Inline objects → ObjectType
//   - Primitives → StringType, NumberType, etc.
//
// The namespacePath enables correct cross-namespace references by calling
// tree.GetRelativeRef to determine the proper reference path.
func BuildTypeExpr(schema spec.Schema, ctx *BuildContext, namespacePath []string) TypeExpr {
	// Handle $ref
	if schema.Ref != "" {
		refName := extractRefName(schema.Ref)
		resolvedRef := ctx.tree.GetRelativeRef(namespacePath, refName)
		return ReferenceType{Name: resolvedRef}
	}

	// Handle composition
	if len(schema.AllOf) > 0 {
		types := make([]TypeExpr, len(schema.AllOf))
		for i, s := range schema.AllOf {
			types[i] = BuildTypeExpr(s, ctx, namespacePath)
		}
		return IntersectionType{Types: types}
	}

	if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 {
		var schemas []spec.Schema
		if len(schema.OneOf) > 0 {
			schemas = schema.OneOf
		} else {
			schemas = schema.AnyOf
		}
		types := make([]TypeExpr, len(schemas))
		for i, s := range schemas {
			types[i] = BuildTypeExpr(s, ctx, namespacePath)
		}
		return UnionType{Types: types}
	}

	// Handle arrays
	if schema.IsArray() {
		if schema.Items != nil {
			return ArrayType{Element: BuildTypeExpr(*schema.Items, ctx, namespacePath)}
		}
		return ArrayType{Element: UnknownType{}}
	}

	// Handle objects without named properties (record type)
	if schema.Type == "object" && len(schema.Properties) == 0 {
		return RecordType{Key: StringType{}, Value: UnknownType{}}
	}

	// Handle inline objects
	if schema.IsObject() && schema.Name == "" {
		properties := make([]Property, len(schema.Properties))
		for i, prop := range schema.Properties {
			properties[i] = Property{
				Name:     prop.Name,
				Type:     BuildTypeExpr(prop.Schema, ctx, namespacePath),
				Optional: !prop.Required,
			}
		}
		return ObjectType{Properties: properties}
	}

	// Handle primitives
	switch schema.Type {
	case "string":
		return StringType{}
	case "integer", "number":
		return NumberType{}
	case "boolean":
		return BooleanType{}
	case "null":
		return NullType{}
	default:
		// Schemas with no type information are treated as unknown.
		// Note: schema.Name being set does NOT indicate a reference - that's what
		// schema.Ref is for. The Name field is just the schema's declared name,
		// which may come from the property name for inline schemas.
		return UnknownType{}
	}
}

// allStrings returns true if all values in the slice are strings.
// Used to determine if an enum can be represented as a TypeScript enum.
func allStrings(values []any) bool {
	for _, v := range values {
		if _, ok := v.(string); !ok {
			return false
		}
	}
	return true
}

// toEnumKey converts a string value to a valid TypeScript enum key.
//
// Transforms: "some-value" → "SomeValue", "SOME_VALUE" → "SomeValue"
func toEnumKey(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, ".", "_")

	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

// extractRefName extracts the type name from a $ref path.
//
// Transforms: "#/components/schemas/User" → "User"
func extractRefName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}
