package golang

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/hntrl/sparktype/internal/contents"
	"github.com/hntrl/sparktype/internal/generators"
	"github.com/hntrl/sparktype/internal/spec"
)

func init() {
	generators.Register("go", Generate, Compare)
}

// BuildContext holds state during AST construction.
//
// Currently empty but reserved for future options.
type BuildContext struct{}

// Generate produces Go struct definitions from a resolved content tree.
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
//   - Package declaration (configurable via "package" option)
//   - Import statements (automatically includes "time" when needed)
//   - Type definitions (struct, type alias, enum) in alphabetical order
//
// Note: Go generator does not support namespaces. If namespaces are present
// in the content tree, schemas are flattened and namespaces are ignored. Consider
// using multiple output files with different packages for organization instead.
//
// This function is also used by Compare to build the expected AST.
func BuildFile(tree *contents.ResolvedTree, opts generators.Options) *File {
	packageName := opts.GetStringOption("package", "types")

	// Collect and sort all schemas alphabetically
	schemas := tree.CollectAllSchemas()
	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].Name < schemas[j].Name
	})

	ctx := &BuildContext{}

	var types []Node
	for _, schema := range schemas {
		types = append(types, BuildSchemaNode(schema, ctx))
	}

	// Check if any type needs the time package
	needsTime := false
	for _, t := range types {
		if checkNeedsTime(t) {
			needsTime = true
			break
		}
	}

	file := &File{
		Header:  "// " + generators.GeneratedFileHeader,
		Package: packageName,
		Types:   types,
	}

	if needsTime {
		file.Imports = []string{"time"}
	}

	return file
}

// BuildSchemaNode converts a schema into the appropriate Go AST node type.
//
// The node type is determined by the schema's characteristics:
//   - Enum schemas → EnumDef (const block with typed values)
//   - Object schemas → Struct
//   - Other schemas → TypeAlias
func BuildSchemaNode(schema spec.Schema, ctx *BuildContext) Node {
	// Handle enums
	if schema.IsEnum() {
		return BuildEnum(schema, ctx)
	}

	// Handle object types
	if schema.IsObject() {
		return BuildStruct(schema, ctx)
	}

	// Handle type aliases
	return BuildTypeAlias(schema, ctx)
}

// BuildEnum converts an enum schema into a Go const block with typed values.
//
// For string enums, produces:
//
//	type Status string
//	const (
//	    StatusActive Status = "active"
//	    StatusInactive Status = "inactive"
//	)
//
// Mixed-type enums produce a comment listing possible values.
func BuildEnum(schema spec.Schema, ctx *BuildContext) Node {
	typeName := toExportedName(schema.Name)

	if allStrings(schema.Enum) {
		values := make([]EnumValue, len(schema.Enum))
		for i, val := range schema.Enum {
			strVal := fmt.Sprintf("%v", val)
			values[i] = EnumValue{
				Name:  toExportedName(strVal),
				Value: strVal,
			}
		}
		return &EnumDef{
			Name:    typeName,
			Comment: Comment(fmt.Sprintf("%s %s", typeName, schema.Description)),
			Values:  values,
		}
	}

	// Mixed type enum - use interface{} with documentation
	strValues := make([]string, len(schema.Enum))
	for i, val := range schema.Enum {
		strValues[i] = fmt.Sprintf("%v", val)
	}
	return &InterfaceComment{
		Name:    typeName,
		Comment: Comment(fmt.Sprintf("%s %s", typeName, schema.Description)),
		Values:  strValues,
	}
}

// BuildStruct converts an object schema into a Go struct.
//
// The struct includes:
//   - GoDoc comment from schema description
//   - Fields with exported names (PascalCase)
//   - JSON struct tags with original field names
//   - Pointer types for optional fields (to distinguish zero values from absent values)
func BuildStruct(schema spec.Schema, ctx *BuildContext) Node {
	typeName := toExportedName(schema.Name)

	// Build required lookup set
	requiredSet := make(map[string]bool)
	for _, r := range schema.Required {
		requiredSet[r] = true
	}

	fields := make([]StructField, len(schema.Properties))
	for i, prop := range schema.Properties {
		fieldName := toExportedName(prop.Name)
		goType := SchemaToGoTypeExpr(prop.Schema, ctx)

		// Handle optional fields with pointers
		isRequired := prop.Required || requiredSet[prop.Name]
		if !isRequired && !IsPointerType(goType) {
			goType = GoPointer{Inner: goType}
		}

		// Build JSON tag with omitempty for optional fields
		jsonTag := prop.Name
		if !isRequired {
			jsonTag += ",omitempty"
		}

		fields[i] = StructField{
			Name:    fieldName,
			Type:    goType,
			JSONTag: jsonTag,
			Comment: Comment(prop.Schema.Description),
		}
	}

	var comment Comment
	if schema.Description != "" {
		comment = Comment(fmt.Sprintf("%s %s", typeName, schema.Description))
	}

	return &Struct{
		Name:    typeName,
		Comment: comment,
		Fields:  fields,
	}
}

// BuildTypeAlias converts a non-object schema into a Go type alias.
//
// Produces: type Name = underlyingType
func BuildTypeAlias(schema spec.Schema, ctx *BuildContext) Node {
	typeName := toExportedName(schema.Name)
	goType := SchemaToGoTypeExpr(schema, ctx)

	var comment Comment
	if schema.Description != "" {
		comment = Comment(fmt.Sprintf("%s %s", typeName, schema.Description))
	}

	return &TypeAlias{
		Name:    typeName,
		Comment: comment,
		Type:    goType,
	}
}

// SchemaToGoTypeExpr converts a schema into a Go type expression.
//
// This handles all type expression forms:
//   - References ($ref) → GoReference (exported name)
//   - Composition (allOf with single schema) → the schema's type
//   - Composition (complex) → interface{} (Go lacks union/intersection types)
//   - Nullable → GoPointer wrapper
//   - Primitives/Arrays → delegated to SchemaToGoTypeExprNonNullable
func SchemaToGoTypeExpr(schema spec.Schema, ctx *BuildContext) GoTypeExpr {
	// Handle $ref
	if schema.Ref != "" {
		return GoReference{Name: toExportedName(extractRefName(schema.Ref))}
	}

	// Handle composition
	if len(schema.AllOf) > 0 {
		if len(schema.AllOf) == 1 {
			return SchemaToGoTypeExpr(schema.AllOf[0], ctx)
		}
		return GoInterface{}
	}

	if len(schema.OneOf) > 0 || len(schema.AnyOf) > 0 {
		// Go doesn't have union types, use interface{}
		return GoInterface{}
	}

	// Handle nullable with pointer type
	if schema.Nullable {
		baseType := SchemaToGoTypeExprNonNullable(schema, ctx)
		if !IsPointerType(baseType) {
			return GoPointer{Inner: baseType}
		}
		return baseType
	}

	return SchemaToGoTypeExprNonNullable(schema, ctx)
}

// SchemaToGoTypeExprNonNullable converts a schema to a Go type, excluding nullability.
//
// This handles:
//   - Arrays → []T
//   - Empty objects → map[string]interface{}
//   - Strings with formats → time.Time, []byte, or string
//   - Integers with formats → int32, int64, or int
//   - Other primitives → float64, bool, interface{}
func SchemaToGoTypeExprNonNullable(schema spec.Schema, ctx *BuildContext) GoTypeExpr {
	// Handle arrays
	if schema.IsArray() {
		if schema.Items != nil {
			return GoSlice{Element: SchemaToGoTypeExpr(*schema.Items, ctx)}
		}
		return GoSlice{Element: GoInterface{}}
	}

	// Handle objects without named properties (map type)
	if schema.Type == "object" && len(schema.Properties) == 0 {
		return GoMap{Key: GoString{}, Value: GoInterface{}}
	}

	// Handle primitives
	switch schema.Type {
	case "string":
		return StringToGoTypeExpr(schema)
	case "integer":
		return IntegerToGoTypeExpr(schema)
	case "number":
		return GoFloat64{}
	case "boolean":
		return GoBool{}
	case "null":
		return GoInterface{}
	default:
		if schema.Name != "" {
			return GoReference{Name: toExportedName(schema.Name)}
		}
		return GoInterface{}
	}
}

// StringToGoTypeExpr converts a string schema to the appropriate Go type.
//
// Format mappings:
//   - "date-time", "date" → time.Time
//   - "binary" → []byte
//   - default → string
func StringToGoTypeExpr(schema spec.Schema) GoTypeExpr {
	switch schema.Format {
	case "date-time", "date":
		return GoTime{}
	case "binary":
		return GoBytes{}
	default:
		return GoString{}
	}
}

// IntegerToGoTypeExpr converts an integer schema to the appropriate Go type.
//
// Format mappings:
//   - "int32" → int32
//   - "int64" → int64
//   - default → int
func IntegerToGoTypeExpr(schema spec.Schema) GoTypeExpr {
	switch schema.Format {
	case "int32":
		return GoInt32{}
	case "int64":
		return GoInt64{}
	default:
		return GoInt{}
	}
}

// checkNeedsTime checks if a node uses time.Time and thus needs the time import.
func checkNeedsTime(node Node) bool {
	switch n := node.(type) {
	case *Struct:
		for _, f := range n.Fields {
			if f.Type.NeedsTimeImport() {
				return true
			}
		}
	case *TypeAlias:
		return n.Type.NeedsTimeImport()
	}
	return false
}

// ============================================================================
// Helpers
// ============================================================================

// toExportedName converts a name to a valid exported Go identifier.
//
// This function:
//   - Converts snake_case and kebab-case to PascalCase
//   - Preserves common acronyms (ID, URL, HTTP, etc.)
//   - Ensures the first character is a letter (prepends "X" if needed)
//
// Examples:
//   - "user_name" → "UserName"
//   - "user-id" → "UserID"
//   - "123abc" → "X123abc"
func toExportedName(s string) string {
	if s == "" {
		return ""
	}

	// Handle snake_case and kebab-case
	s = strings.ReplaceAll(s, "-", "_")
	parts := strings.Split(s, "_")

	var result strings.Builder
	for _, part := range parts {
		if len(part) > 0 {
			upper := strings.ToUpper(part)
			if isCommonAcronym(upper) {
				result.WriteString(upper)
			} else {
				result.WriteString(strings.ToUpper(part[:1]))
				result.WriteString(part[1:])
			}
		}
	}

	name := result.String()

	// Ensure first character is valid for Go identifier
	if len(name) > 0 && !unicode.IsLetter(rune(name[0])) {
		name = "X" + name
	}

	return name
}

// isCommonAcronym returns true if the string is a common acronym
// that should be fully capitalized in Go (per Go naming conventions).
func isCommonAcronym(s string) bool {
	acronyms := map[string]bool{
		"ID": true, "URL": true, "HTTP": true, "HTTPS": true,
		"API": true, "JSON": true, "XML": true, "HTML": true,
		"UUID": true, "URI": true, "SQL": true, "SSH": true,
		"TCP": true, "UDP": true, "IP": true, "DNS": true,
	}
	return acronyms[s]
}

// allStrings returns true if all values in the slice are strings.
// Used to determine if an enum can use typed string constants.
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
