package spec

// Schema represents a normalized OpenAPI schema definition.
//
// This struct captures all relevant information from an OpenAPI schema,
// normalized into a consistent format that generators can work with.
// It supports objects, arrays, primitives, enums, and composition types.
type Schema struct {
	// Name is the schema's identifier, typically from components/schemas.
	// For inline schemas (e.g., array items), this may be empty.
	Name string

	// Description is human-readable documentation for this schema.
	// Generators typically convert this to language-appropriate comments.
	Description string

	// Type is the JSON Schema type: "object", "array", "string", "number",
	// "integer", "boolean", or "null". May be empty for composed types.
	Type string

	// Format provides additional type information for string/number types.
	// Common values: "date-time", "date", "email", "uri", "uuid", "int32", "int64".
	Format string

	// Properties contains the object's property definitions.
	// Only populated when Type is "object" or when the schema has properties.
	// Properties are ordered consistently (alphabetically by name).
	Properties []Property

	// Required lists property names that must be present.
	// Cross-reference with Properties[i].Name to determine if a property is required.
	// Individual Property.Required fields are also set for convenience.
	Required []string

	// Items defines the element type for array schemas.
	// Only populated when Type is "array".
	Items *Schema

	// Enum constrains the value to a fixed set of values.
	// Values are typically strings but may be numbers or mixed types.
	// When present, generators may produce enum types or union literals.
	Enum []any

	// AllOf combines schemas using logical AND (intersection).
	// The resulting type must satisfy all listed schemas.
	// Generators typically produce intersection types or merged interfaces.
	AllOf []Schema

	// OneOf combines schemas using exclusive OR (discriminated union).
	// The value must match exactly one of the listed schemas.
	// Generators typically produce union types.
	OneOf []Schema

	// AnyOf combines schemas using logical OR (union).
	// The value must match at least one of the listed schemas.
	// Generators typically produce union types (same as OneOf in most languages).
	AnyOf []Schema

	// Ref holds the original $ref string (e.g., "#/components/schemas/User").
	// When set, this schema is a reference to another schema definition.
	// Generators extract the referenced type name for cross-references.
	Ref string

	// Nullable indicates the value may be null in addition to the base type.
	// Generators typically produce union with null (T | null) or optional types.
	Nullable bool

	// ReadOnly indicates the property is only returned in responses, not accepted in requests.
	// Generators may add readonly modifiers to such properties.
	ReadOnly bool

	// WriteOnly indicates the property is only accepted in requests, not returned in responses.
	// This is less commonly used in type generation.
	WriteOnly bool

	// Default is the default value when the property is not provided.
	// Primarily informational; not typically used in type generation.
	Default any

	// Example provides an example value for documentation.
	// Primarily informational; not typically used in type generation.
	Example any

	// Minimum constrains numeric values to be >= this value.
	// Generators that support runtime validation (e.g., Zod) use this.
	Minimum *float64

	// Maximum constrains numeric values to be <= this value.
	// Generators that support runtime validation (e.g., Zod) use this.
	Maximum *float64

	// MinLength constrains string length to be >= this value.
	// Generators that support runtime validation (e.g., Zod) use this.
	MinLength *int

	// MaxLength constrains string length to be <= this value.
	// Generators that support runtime validation (e.g., Zod) use this.
	MaxLength *int

	// Pattern is a regex pattern that string values must match.
	// Generators that support runtime validation (e.g., Zod) use this.
	Pattern string

	// MinItems constrains array length to be >= this value.
	// Generators that support runtime validation (e.g., Zod) use this.
	MinItems *int

	// MaxItems constrains array length to be <= this value.
	// Generators that support runtime validation (e.g., Zod) use this.
	MaxItems *int

	// UniqueItems indicates array elements must be unique.
	// Primarily informational; not commonly used in type generation.
	UniqueItems bool

	// SourceSpec is the name of the spec this schema was loaded from.
	// Used for error messages and debugging. Set by LoadSpec().
	SourceSpec string
}

// Property represents a single property within an object schema.
//
// Properties combine a name with a schema definition and track whether
// the property is required. The Schema field contains the full type
// information for the property's value.
type Property struct {
	// Name is the property's key in the object.
	// This becomes the field/property name in generated code.
	Name string

	// Schema defines the property's value type.
	// May be a primitive, reference, array, or nested object.
	Schema Schema

	// Required indicates this property must be present.
	// When false, generators produce optional types (e.g., T | undefined, *T).
	Required bool
}

// IsObject returns true if the schema represents an object type.
// A schema is considered an object if its Type is "object" or if it has Properties.
func (s *Schema) IsObject() bool {
	return s.Type == "object" || len(s.Properties) > 0
}

// IsArray returns true if the schema represents an array type.
// Check Items for the element type definition.
func (s *Schema) IsArray() bool {
	return s.Type == "array"
}

// IsEnum returns true if the schema constrains values to an enumeration.
// Check Enum for the list of allowed values.
func (s *Schema) IsEnum() bool {
	return len(s.Enum) > 0
}

// IsPrimitive returns true if the schema is a primitive (non-composite) type.
// Primitive types are: string, number, integer, boolean.
func (s *Schema) IsPrimitive() bool {
	switch s.Type {
	case "string", "number", "integer", "boolean":
		return true
	}
	return false
}

// HasComposition returns true if the schema uses allOf, oneOf, or anyOf.
// Composed schemas combine multiple sub-schemas using set operations.
func (s *Schema) HasComposition() bool {
	return len(s.AllOf) > 0 || len(s.OneOf) > 0 || len(s.AnyOf) > 0
}
