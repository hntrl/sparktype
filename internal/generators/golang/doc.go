// Package golang generates Go struct definitions from OpenAPI schemas.
//
// # Output Format
//
// The generator produces Go code with:
//
//   - Struct types for object schemas with JSON struct tags
//   - Type aliases for non-object schemas
//   - Const blocks with typed constants for string enumerations
//   - Proper import statements (e.g., "time" for date-time formats)
//
// # AST Architecture
//
// The generator uses an AST-based approach:
//
//  1. BuildFile constructs a File AST from the ResolvedTree
//  2. Each AST node has a Serialize() method producing Go code
//  3. The File tracks required imports and includes them automatically
//
// # AST Node Types
//
//   - File: Root node with package declaration, imports, and type definitions
//   - Struct: type X struct { ... } with field definitions
//   - TypeDef: type X Y (type alias preserving underlying type)
//   - TypeAlias: type X = Y (true type alias)
//   - EnumDef: Const block with typed string constants
//   - StructField: Field with name, type, JSON tag, and optional comment
//   - EnumValue: Single const value in an enum block
//
// Type expression nodes:
//   - StringType, IntType, Int64Type, Float64Type, BoolType: Primitive types
//   - SliceType: []T
//   - MapType: map[K]V
//   - PointerType: *T (used for optional fields)
//   - TimeType: time.Time (for date-time formats)
//   - ReferenceType: Reference to another defined type
//
// # JSON Tags
//
// All struct fields include JSON tags for serialization:
//   - Required fields: `json:"fieldName"`
//   - Optional fields: `json:"fieldName,omitempty"`
//
// # Pointer Types
//
// Optional and nullable fields use pointer types (*T) to distinguish between
// zero values and missing/null values. This is important for proper JSON
// marshaling behavior.
//
// # Namespace Handling
//
// Go doesn't have namespaces within a package. When namespaces are used in the
// configuration, the generator will raise an error since Go's package system
// should be used for organization instead.
package golang
