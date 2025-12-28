package zod

import (
	"fmt"
	"strings"

	"github.com/hntrl/sparktype/internal/contents"
	"github.com/hntrl/sparktype/internal/generators"
	"github.com/hntrl/sparktype/internal/spec"
)

func init() {
	generators.Register("zod", Generate, Compare)
}

// BuildContext holds state during AST construction.
//
// The context is passed through all Build* functions to provide access to:
//   - The resolved tree for cross-reference resolution
//   - A set of known schema names for validating references
//   - The inferTypes option that controls output structure
type BuildContext struct {
	// tree is the resolved content tree, used for GetRelativeRef lookups.
	tree *contents.ResolvedTree

	// schemaNames is a set of all schema names in the tree.
	// Used to validate references and determine if z.lazy() is needed.
	schemaNames map[string]bool

	// inferTypes enables JSDoc transfer mode:
	// - Internal schema types (non-exported)
	// - Public interfaces with indexed property types and JSDoc comments
	inferTypes bool
}

// Generate produces Zod schema definitions from a resolved content tree.
//
// This is the main entry point called by the generator registry. It builds
// a File AST and serializes it to produce the final output.
func Generate(tree *contents.ResolvedTree, opts generators.Options) ([]byte, error) {
	file := BuildFile(tree, opts)
	return []byte(file.Serialize()), nil
}

// BuildFile constructs a complete File AST from a resolved content tree.
//
// The File contains multiple sections:
//  1. Header comment and Zod import
//  2. Schema declarations (const userSchema = z.object({...})) - all flattened to top level
//  3. Type section (z.infer type aliases or namespace blocks)
//  4. Interface section (when inferTypes=true, TypeScript interfaces with JSDoc)
//
// Schemas are always at the top level. Namespaces only contain type exports
// that reference those top-level schemas.
//
// This function is also used by Compare to build the expected AST.
func BuildFile(tree *contents.ResolvedTree, opts generators.Options) *File {
	// Build lookup set of all schema names for reference validation
	allSchemas := tree.CollectAllSchemas()
	schemaNames := make(map[string]bool)
	for _, s := range allSchemas {
		schemaNames[s.Name] = true
	}

	ctx := &BuildContext{
		tree:        tree,
		schemaNames: schemaNames,
		inferTypes:  opts.GetBoolOption("inferTypes", false),
	}

	// Flatten all schemas to top-level declarations
	file := &File{
		Header:  "// " + generators.GeneratedFileHeader + "\n",
		Import:  "import * as z from \"zod\";\n",
		Schemas: BuildFlattenedSchemas(tree.Nodes, ctx, []string{}),
	}

	// Add type section header
	if ctx.inferTypes {
		file.TypeSection = append(file.TypeSection, &TypeSectionComment{Text: "Internal schema types"})
	} else {
		file.TypeSection = append(file.TypeSection, &TypeSectionComment{Text: "Inferred types"})
	}
	file.TypeSection = append(file.TypeSection, BuildTypeNodes(tree.Nodes, ctx, []string{})...)

	// Add interfaces section when inferTypes is enabled
	if ctx.inferTypes {
		file.Interfaces = append(file.Interfaces, &TypeSectionComment{Text: "TypeScript interfaces"})
		file.Interfaces = append(file.Interfaces, BuildInterfaceNodes(tree.Nodes, ctx, []string{})...)
	}

	return file
}

// ============================================================================
// Schema Building
// ============================================================================

// BuildFlattenedSchemas recursively collects all schemas from the content tree
// and creates top-level SchemaDecl nodes.
//
// All schemas are flattened to the top level regardless of namespace depth.
// This allows namespaces to reference them via type exports.
func BuildFlattenedSchemas(nodes []contents.Node, ctx *BuildContext, namespacePath []string) []Node {
	var result []Node

	for _, node := range nodes {
		if node.IsSchema() {
			result = append(result, BuildSchemaDecl(*node.Schema, ctx, namespacePath))
		} else if node.IsNamespace() {
			// Recursively collect schemas from nested namespaces
			currentPath := append(namespacePath, node.Namespace.Name)
			result = append(result, BuildFlattenedSchemas(node.Namespace.Children, ctx, currentPath)...)
		}
	}

	return result
}

// BuildSchemaDecl creates a top-level schema declaration.
//
// Produces: export const userSchema = z.object({...});
func BuildSchemaDecl(schema spec.Schema, ctx *BuildContext, namespacePath []string) Node {
	schemaVarName := toCamelCase(schema.Name) + "Schema"
	return &SchemaDecl{
		Name:   schemaVarName,
		Schema: SchemaToZodExpr(schema, ctx, namespacePath),
	}
}

// SchemaToZodExpr converts a schema into its Zod expression representation.
//
// This handles the top-level schema structure:
//   - Enums → z.enum([...]) or z.union([z.literal(...), ...])
//   - Objects → z.object({...})
//   - Other types → delegated to SchemaToZodType
func SchemaToZodExpr(schema spec.Schema, ctx *BuildContext, namespacePath []string) ZodExpr {
	// Handle enums
	if schema.IsEnum() {
		if allStrings(schema.Enum) {
			values := make([]string, len(schema.Enum))
			for i, v := range schema.Enum {
				values[i] = fmt.Sprintf("%v", v)
			}
			return ZEnum{Values: values}
		}
		// Mixed enum - use union of literals
		schemas := make([]ZodExpr, len(schema.Enum))
		for i, v := range schema.Enum {
			schemas[i] = ZLiteral{Value: v}
		}
		return ZUnion{Schemas: schemas}
	}

	// Handle object types
	if schema.IsObject() {
		return BuildObjectExpr(schema, ctx, namespacePath)
	}

	// Handle other types
	return SchemaToZodType(schema, ctx, namespacePath)
}

// BuildObjectExpr creates a z.object({...}) expression from an object schema.
//
// Each property is converted to a Zod expression with appropriate modifiers
// for optionality and nullability.
func BuildObjectExpr(schema spec.Schema, ctx *BuildContext, namespacePath []string) ZodExpr {
	// Build required lookup set
	requiredSet := make(map[string]bool)
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	properties := make([]ZProperty, len(schema.Properties))
	for i, prop := range schema.Properties {
		propExpr := SchemaToZodType(prop.Schema, ctx, namespacePath)

		// Add modifiers for optionality and nullability
		var modifiers []Modifier
		isRequired := prop.Required || requiredSet[prop.Name]
		if !isRequired {
			modifiers = append(modifiers, Optional{})
		}
		if prop.Schema.Nullable {
			modifiers = append(modifiers, Nullable{})
		}

		if len(modifiers) > 0 {
			propExpr = ZWithModifiers{Expr: propExpr, Modifiers: modifiers}
		}

		properties[i] = ZProperty{
			Name:   prop.Name,
			Schema: propExpr,
		}
	}

	return ZObject{Properties: properties}
}

// SchemaToZodType converts a schema into a Zod type expression.
//
// This handles all type expression forms:
//   - References ($ref) → z.lazy(() => schemaRef)
//   - Composition (allOf) → schema1.and(schema2)
//   - Composition (oneOf/anyOf) → z.union([...])
//   - Arrays → z.array(element)
//   - Records → z.record(key, value)
//   - Primitives → z.string(), z.number(), etc.
//
// References use z.lazy() to handle circular dependencies between schemas.
func SchemaToZodType(schema spec.Schema, ctx *BuildContext, namespacePath []string) ZodExpr {
	// Handle $ref - use z.lazy for forward/circular references
	if schema.Ref != "" {
		refName := extractRefName(schema.Ref)
		if ctx.schemaNames[refName] {
			refPath := ctx.tree.GetRelativeRef(namespacePath, refName)
			parts := strings.Split(refPath, ".")
			parts[len(parts)-1] = toCamelCase(parts[len(parts)-1]) + "Schema"
			schemaRef := strings.Join(parts, ".")
			return ZLazy{Ref: schemaRef}
		}
		return ZUnknown{}
	}

	// Handle composition - allOf
	if len(schema.AllOf) > 0 {
		schemas := make([]ZodExpr, len(schema.AllOf))
		for i, s := range schema.AllOf {
			schemas[i] = SchemaToZodType(s, ctx, namespacePath)
		}
		return ZIntersection{Schemas: schemas}
	}

	// Handle composition - oneOf/anyOf
	if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 {
		var subSchemas []spec.Schema
		if len(schema.OneOf) > 0 {
			subSchemas = schema.OneOf
		} else {
			subSchemas = schema.AnyOf
		}
		schemas := make([]ZodExpr, len(subSchemas))
		for i, s := range subSchemas {
			schemas[i] = SchemaToZodType(s, ctx, namespacePath)
		}
		return ZUnion{Schemas: schemas}
	}

	// Handle arrays
	if schema.IsArray() {
		var element ZodExpr = ZUnknown{}
		if schema.Items != nil {
			element = SchemaToZodType(*schema.Items, ctx, namespacePath)
		}
		var modifiers []Modifier
		if schema.MinItems != nil {
			modifiers = append(modifiers, Min{V: *schema.MinItems})
		}
		if schema.MaxItems != nil {
			modifiers = append(modifiers, Max{V: *schema.MaxItems})
		}
		return ZArray{Element: element, Modifiers: modifiers}
	}

	// Handle objects without named properties (record type)
	if schema.Type == "object" && len(schema.Properties) == 0 {
		return ZRecord{Key: ZString{}, Value: ZUnknown{}}
	}

	// Handle primitives
	switch schema.Type {
	case "string":
		return BuildStringExpr(schema)
	case "integer":
		return BuildIntegerExpr(schema)
	case "number":
		return BuildNumberExpr(schema)
	case "boolean":
		return ZBoolean{}
	case "null":
		return ZNull{}
	default:
		return ZUnknown{}
	}
}

// BuildStringExpr creates a z.string() expression with appropriate modifiers.
//
// Adds modifiers for:
//   - Format validators: .email(), .url(), .uuid(), .datetime(), .date()
//   - Length constraints: .min(n), .max(n)
//   - Pattern matching: .regex(/pattern/)
func BuildStringExpr(schema spec.Schema) ZodExpr {
	var modifiers []Modifier

	// Format-based validators
	switch schema.Format {
	case "email":
		modifiers = append(modifiers, Email{})
	case "uri", "url":
		modifiers = append(modifiers, Url{})
	case "uuid":
		modifiers = append(modifiers, Uuid{})
	case "date-time":
		modifiers = append(modifiers, Datetime{})
	case "date":
		modifiers = append(modifiers, Date{})
	}

	// Length constraints
	if schema.MinLength != nil {
		modifiers = append(modifiers, Min{V: *schema.MinLength})
	}
	if schema.MaxLength != nil {
		modifiers = append(modifiers, Max{V: *schema.MaxLength})
	}
	if schema.Pattern != "" {
		modifiers = append(modifiers, Regex{Pattern: schema.Pattern})
	}

	return ZString{Modifiers: modifiers}
}

// BuildIntegerExpr creates a z.number().int() expression with constraints.
//
// Always includes .int() modifier, plus optional .min() and .max().
func BuildIntegerExpr(schema spec.Schema) ZodExpr {
	modifiers := []Modifier{Int{}}
	if schema.Minimum != nil {
		modifiers = append(modifiers, MinFloat{V: *schema.Minimum})
	}
	if schema.Maximum != nil {
		modifiers = append(modifiers, MaxFloat{V: *schema.Maximum})
	}
	return ZNumber{Modifiers: modifiers}
}

// BuildNumberExpr creates a z.number() expression with optional constraints.
func BuildNumberExpr(schema spec.Schema) ZodExpr {
	var modifiers []Modifier
	if schema.Minimum != nil {
		modifiers = append(modifiers, MinFloat{V: *schema.Minimum})
	}
	if schema.Maximum != nil {
		modifiers = append(modifiers, MaxFloat{V: *schema.Maximum})
	}
	return ZNumber{Modifiers: modifiers}
}

// ============================================================================
// Type Building
// ============================================================================

// BuildTypeNodes creates inferred type definitions from content nodes.
//
// For root-level schemas, produces direct type aliases.
// For namespaced schemas, creates TypeScript namespace blocks containing type exports.
//
// Normal mode output:
//
//	export type User = z.infer<typeof userSchema>;
//	export namespace Models {
//	  export type User = z.infer<typeof userSchema>;
//	}
//
// inferTypes mode output:
//
//	type UserSchemaType = z.infer<typeof userSchema>;
func BuildTypeNodes(nodes []contents.Node, ctx *BuildContext, namespacePath []string) []Node {
	var result []Node

	for _, node := range nodes {
		if node.IsSchema() {
			result = append(result, BuildInferredType(*node.Schema, ctx, namespacePath))
		} else if node.IsNamespace() {
			// Build a TypeScript namespace with type exports
			result = append(result, BuildTypeNamespace(node.Namespace, ctx, namespacePath))
		}
	}

	return result
}

// BuildTypeNamespace creates a TypeScript namespace containing type exports.
//
// The namespace contains type aliases that reference the top-level schemas:
//
//	export namespace Models {
//	  export type User = z.infer<typeof userSchema>;
//	}
func BuildTypeNamespace(ns *contents.NamespaceNode, ctx *BuildContext, parentPath []string) Node {
	currentPath := append(parentPath, ns.Name)

	var children []Node
	for _, node := range ns.Children {
		if node.IsSchema() {
			children = append(children, BuildNamespaceTypeExport(*node.Schema, ctx, currentPath))
		} else if node.IsNamespace() {
			// Handle nested namespaces recursively
			children = append(children, BuildTypeNamespace(node.Namespace, ctx, currentPath))
		}
	}

	return &TSNamespace{
		Name:     ns.Name,
		Children: children,
	}
}

// BuildNamespaceTypeExport creates a type export inside a namespace.
//
// Produces: export type User = z.infer<typeof userSchema>;
// The schema reference is always to the top-level schema.
func BuildNamespaceTypeExport(schema spec.Schema, ctx *BuildContext, namespacePath []string) Node {
	schemaVarName := toCamelCase(schema.Name) + "Schema"
	return &InferredType{
		Name:      schema.Name,
		SchemaRef: schemaVarName,
		Export:    true,
	}
}

// BuildInferredType creates a z.infer type alias for a schema.
//
// In normal mode, exports the type directly:
//
//	export type User = z.infer<typeof userSchema>;
//
// In inferTypes mode, creates internal type for interface indexing:
//
//	type UserSchemaType = z.infer<typeof userSchema>;
func BuildInferredType(schema spec.Schema, ctx *BuildContext, namespacePath []string) Node {
	schemaVarName := toCamelCase(schema.Name) + "Schema"

	if ctx.inferTypes {
		// Internal type for interface indexing
		schemaTypeName := schema.Name + "SchemaType"
		return &InferredType{
			Name:      schemaTypeName,
			SchemaRef: schemaVarName,
			Export:    false,
		}
	}

	// Export type directly (for root-level schemas only)
	return &InferredType{
		Name:      schema.Name,
		SchemaRef: schemaVarName,
		Export:    true,
	}
}

// ============================================================================
// Interface Building (for inferTypes mode)
// ============================================================================

// BuildInterfaceNodes creates TypeScript interfaces from content nodes.
//
// Only called when inferTypes=true. Produces interfaces with:
//   - JSDoc comments from schema descriptions
//   - Properties using indexed types: SchemaType["prop"]
func BuildInterfaceNodes(nodes []contents.Node, ctx *BuildContext, namespacePath []string) []Node {
	var result []Node

	for _, node := range nodes {
		if node.IsSchema() {
			result = append(result, BuildInterface(*node.Schema, ctx, namespacePath))
		} else if node.IsNamespace() {
			result = append(result, BuildTSNamespace(node.Namespace, ctx, namespacePath))
		}
	}

	return result
}

// BuildTSNamespace creates a TypeScript namespace for interface organization.
//
// Produces: export namespace Models { ... }
func BuildTSNamespace(ns *contents.NamespaceNode, ctx *BuildContext, parentPath []string) Node {
	currentPath := append(parentPath, ns.Name)
	return &TSNamespace{
		Name:     ns.Name,
		Children: BuildInterfaceNodes(ns.Children, ctx, currentPath),
	}
}

// BuildInterface creates a TypeScript interface from an object schema.
//
// Produces interfaces with indexed property types that transfer JSDoc:
//
//	export interface User {
//	  /** User's email address */
//	  email: UserSchemaType["email"];
//	}
//
// Non-object schemas produce type aliases instead.
func BuildInterface(schema spec.Schema, ctx *BuildContext, namespacePath []string) Node {
	// Get the schema type name for indexing
	var schemaTypeName string
	if len(namespacePath) == 0 {
		schemaTypeName = schema.Name + "SchemaType"
	} else {
		schemaTypeName = strings.Join(append(namespacePath, schema.Name), "_") + "SchemaType"
	}

	// Handle enums - just a type alias
	if schema.IsEnum() {
		return &TypeAlias{
			Name:        schema.Name,
			Description: Description(schema.Description),
			Type:        schemaTypeName,
			Export:      true,
		}
	}

	// Handle object types
	if schema.IsObject() {
		requiredSet := make(map[string]bool)
		for _, r := range schema.Required {
			requiredSet[r] = true
		}

		properties := make([]InterfaceProperty, len(schema.Properties))
		for i, prop := range schema.Properties {
			isRequired := prop.Required || requiredSet[prop.Name]
			properties[i] = InterfaceProperty{
				Name:        prop.Name,
				IndexType:   schemaTypeName,
				IndexKey:    prop.Name,
				Optional:    !isRequired,
				ReadOnly:    prop.Schema.ReadOnly,
				Description: Description(prop.Schema.Description),
			}
		}

		return &Interface{
			Name:        schema.Name,
			Description: Description(schema.Description),
			Properties:  properties,
		}
	}

	// Handle type aliases
	return &TypeAlias{
		Name:        schema.Name,
		Description: Description(schema.Description),
		Type:        schemaTypeName,
		Export:      true,
	}
}

// ============================================================================
// Helpers
// ============================================================================

// toCamelCase converts a PascalCase name to camelCase.
//
// Transforms: "UserSchema" → "userSchema"
func toCamelCase(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// allStrings returns true if all values in the slice are strings.
// Used to determine if an enum can use z.enum() vs z.union().
func allStrings(values []any) bool {
	for _, v := range values {
		if _, ok := v.(string); !ok {
			return false
		}
	}
	return true
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
