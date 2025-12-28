// Package python generates Python TypedDict class definitions from OpenAPI schemas.
//
// # Output Format
//
// The generator produces Python code with:
//
//   - TypedDict classes for object schemas
//   - Enum classes (using StrEnum) for string enumerations
//   - Type aliases for non-object schemas
//   - Proper import statements for typing constructs
//
// # AST Architecture
//
// The generator uses an AST-based approach:
//
//  1. BuildFile constructs a File AST from the ResolvedTree
//  2. Each AST node has a Serialize() method producing Python code
//  3. Indentation is handled by parent nodes when serializing children
//
// # AST Node Types
//
//   - File: Root node with imports and class definitions
//   - TypedDict: class X(TypedDict): with field definitions
//   - EnumClass: class X(StrEnum): with enum members
//   - TypeAlias: X = Union[...] or X = List[...] etc.
//   - Field: TypedDict field with name, type, and optional description
//   - EnumMember: Enum value with name and string value
//   - Docstring: Python docstring (triple-quoted) for class documentation
//
// Type expression nodes:
//   - StrType, IntType, FloatType, BoolType, NoneType: Primitive types
//   - ListType: List[T]
//   - DictType: Dict[K, V]
//   - UnionType: Union[T, U, ...]
//   - OptionalType: Optional[T] (sugar for Union[T, None])
//   - LiteralType: Literal["value"]
//   - ReferenceType: Reference to another defined type
//
// # Documentation
//
// The generator produces Python docstrings for class-level documentation
// (TypedDict and EnumClass) and inline comments for field descriptions.
// This ensures documentation appears in IDE tooltips and generated API docs.
//
// # Namespace Handling
//
// Python doesn't have a direct namespace equivalent. When namespaces are used
// in the configuration, the generator will raise an error since Python's module
// system should be used for organization instead.
package python
