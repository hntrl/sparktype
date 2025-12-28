package python

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hntrl/sparktype/internal/contents"
	"github.com/hntrl/sparktype/internal/generators"
	"github.com/hntrl/sparktype/internal/spec"
)

func init() {
	generators.Register("python", Generate, Compare)
}

// BuildContext holds state during AST construction.
//
// The context is passed through all Build* functions to provide access to
// generation options that affect output style.
type BuildContext struct {
	// generateEnums produces Python Enum classes instead of Literal unions.
	generateEnums bool
}

// Generate produces Python TypedDict definitions from a resolved content tree.
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
//   - Import statements for typing constructs
//   - Class definitions (TypedDict, Enum, TypeAlias) in alphabetical order
//
// Note: Python generator does not support namespaces. If namespaces are present
// in the content tree, schemas are flattened and namespaces are ignored. Consider
// using multiple output files for organization instead.
//
// This function is also used by Compare to build the expected AST.
func BuildFile(tree *contents.ResolvedTree, opts generators.Options) *File {
	ctx := &BuildContext{
		generateEnums: opts.GlobalOptions.GenerateEnums,
	}

	// Collect and sort all schemas alphabetically
	schemas := tree.CollectAllSchemas()
	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].Name < schemas[j].Name
	})

	file := &File{
		Header: "# " + generators.GeneratedFileHeader,
		Imports: []string{
			"from typing import TypedDict, NotRequired, Literal, Union, List, Dict, Any",
			"from enum import Enum",
		},
	}

	for _, schema := range schemas {
		file.Classes = append(file.Classes, BuildSchemaNode(schema, ctx))
	}

	return file
}

// BuildSchemaNode converts a schema into the appropriate Python AST node type.
//
// The node type is determined by the schema's characteristics:
//   - Enum schemas → EnumClass or TypeAlias (Literal union)
//   - Object schemas → TypedDict
//   - Other schemas → TypeAlias
func BuildSchemaNode(schema spec.Schema, ctx *BuildContext) Node {
	// Handle enums
	if schema.IsEnum() {
		return BuildEnum(schema, ctx)
	}

	// Handle object types
	if schema.IsObject() {
		return BuildTypedDict(schema, ctx)
	}

	// Handle type aliases
	return BuildTypeAlias(schema, ctx)
}

// BuildEnum converts an enum schema into either an EnumClass or TypeAlias.
//
// When generateEnums is true and all values are strings, produces a Python Enum:
//
//	class Status(str, Enum):
//	    ACTIVE = "active"
//	    INACTIVE = "inactive"
//
// Otherwise produces a Literal type alias:
//
//	Status = Literal["active", "inactive"]
func BuildEnum(schema spec.Schema, ctx *BuildContext) Node {
	if ctx.generateEnums && allStrings(schema.Enum) {
		// Generate Python Enum class
		members := make([]EnumMember, len(schema.Enum))
		for i, val := range schema.Enum {
			strVal := fmt.Sprintf("%v", val)
			members[i] = EnumMember{
				Key:   strings.ToUpper(toSnakeCase(strVal)),
				Value: strVal,
			}
		}
		return &EnumClass{
			Name:      schema.Name,
			Docstring: Docstring(schema.Description),
			Members:   members,
		}
	}

	// Generate Literal type alias
	return &TypeAlias{
		Name:    schema.Name,
		Comment: Comment(schema.Description),
		Type:    LiteralType{Values: schema.Enum},
	}
}

// BuildTypedDict converts an object schema into a TypedDict class.
//
// The function handles Python's TypedDict requirements:
//   - All-required or all-optional TypedDicts use total=True/False
//   - Mixed TypedDicts use NotRequired[] wrapper for optional fields
//
// Field names are converted to snake_case to follow Python conventions.
func BuildTypedDict(schema spec.Schema, ctx *BuildContext) Node {
	// Build required lookup set
	requiredSet := make(map[string]bool)
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	// Determine if all properties are required, all optional, or mixed
	allRequired := true
	allOptional := true
	for _, prop := range schema.Properties {
		isRequired := prop.Required || requiredSet[prop.Name]
		if isRequired {
			allOptional = false
		} else {
			allRequired = false
		}
	}

	fields := make([]Field, len(schema.Properties))
	for i, prop := range schema.Properties {
		pyType := SchemaToPyTypeExpr(prop.Schema, ctx)
		isRequired := prop.Required || requiredSet[prop.Name]

		// Use NotRequired wrapper for optional fields in mixed TypedDicts
		if !allRequired && !allOptional && !isRequired {
			pyType = NotRequiredType{Inner: pyType}
		}

		fields[i] = Field{
			Name:      toSnakeCase(prop.Name),
			Type:      pyType,
			Docstring: Docstring(prop.Schema.Description),
		}
	}

	return &TypedDict{
		Name:      schema.Name,
		Docstring: Docstring(schema.Description),
		Fields:    fields,
		Total:     allRequired || (!allOptional && !allRequired), // Total is true unless all optional
	}
}

// BuildTypeAlias converts a non-object schema into a TypeAlias.
//
// Used for primitive types, arrays, unions, and other non-object schemas
// that can't be represented as TypedDict classes.
func BuildTypeAlias(schema spec.Schema, ctx *BuildContext) Node {
	pyType := SchemaToPyTypeExpr(schema, ctx)
	if schema.Nullable {
		pyType = OptionalType{Inner: pyType}
	}
	return &TypeAlias{
		Name:    schema.Name,
		Comment: Comment(schema.Description),
		Type:    pyType,
	}
}

// SchemaToPyTypeExpr converts a schema into a Python type expression.
//
// This handles all type expression forms:
//   - References ($ref) → ReferenceType (uses the schema name directly)
//   - Composition (allOf) → First schema (Python lacks intersection types)
//   - Composition (oneOf/anyOf) → UnionType
//   - Nullable → OptionalType wrapper
//   - Primitives/Arrays → delegated to SchemaToPyTypeExprNonNullable
//
// Python does not have a direct equivalent to TypeScript's intersection types,
// so allOf compositions use only the first schema.
func SchemaToPyTypeExpr(schema spec.Schema, ctx *BuildContext) PyTypeExpr {
	// Handle $ref
	if schema.Ref != "" {
		refName := extractRefName(schema.Ref)
		return ReferenceType{Name: refName}
	}

	// Handle composition
	if len(schema.AllOf) > 0 {
		// Python doesn't have intersection types, use the first one
		if len(schema.AllOf) > 0 {
			return SchemaToPyTypeExpr(schema.AllOf[0], ctx)
		}
	}

	if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 {
		var schemas []spec.Schema
		if len(schema.OneOf) > 0 {
			schemas = schema.OneOf
		} else {
			schemas = schema.AnyOf
		}
		types := make([]PyTypeExpr, len(schemas))
		for i, s := range schemas {
			types[i] = SchemaToPyTypeExpr(s, ctx)
		}
		return UnionType{Types: types}
	}

	// Handle nullable
	if schema.Nullable {
		return OptionalType{Inner: SchemaToPyTypeExprNonNullable(schema, ctx)}
	}

	return SchemaToPyTypeExprNonNullable(schema, ctx)
}

// SchemaToPyTypeExprNonNullable converts a schema to a Python type, excluding nullability.
//
// This handles:
//   - Arrays → List[T]
//   - Empty objects → Dict[str, Any]
//   - Primitives → str, int, float, bool, None
//   - Named schemas → Reference to the type
func SchemaToPyTypeExprNonNullable(schema spec.Schema, ctx *BuildContext) PyTypeExpr {
	// Handle arrays
	if schema.IsArray() {
		if schema.Items != nil {
			return ListType{Element: SchemaToPyTypeExpr(*schema.Items, ctx)}
		}
		return ListType{Element: AnyType{}}
	}

	// Handle objects without named properties (dict type)
	if schema.Type == "object" && len(schema.Properties) == 0 {
		return DictType{Key: StrType{}, Value: AnyType{}}
	}

	// Handle primitives
	switch schema.Type {
	case "string":
		return StrType{}
	case "integer":
		return IntType{}
	case "number":
		return FloatType{}
	case "boolean":
		return BoolType{}
	case "null":
		return NoneType{}
	default:
		if schema.Name != "" {
			return ReferenceType{Name: schema.Name}
		}
		return AnyType{}
	}
}

// ============================================================================
// Helpers
// ============================================================================

// toSnakeCase converts a camelCase or PascalCase string to snake_case.
//
// Transforms: "userName" → "user_name", "HTTPRequest" → "h_t_t_p_request"
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// allStrings returns true if all values in the slice are strings.
// Used to determine if an enum can use Python's Enum class.
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
