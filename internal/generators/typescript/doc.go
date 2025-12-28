// Package typescript generates TypeScript interface and type alias definitions
// from OpenAPI schemas.
//
// # Output Format
//
// The generator produces TypeScript code with:
//
//   - Interface declarations for object schemas
//   - Type aliases for non-object schemas (primitives, unions, arrays)
//   - Enum declarations (when generateEnums option is enabled)
//   - Namespace blocks for hierarchical organization
//
// # AST Architecture
//
// The generator uses an AST-based approach:
//
//  1. BuildFile constructs a File AST from the ResolvedTree
//  2. Each AST node (Interface, TypeAlias, Namespace, etc.) has a Serialize() method
//  3. The File.Serialize() method produces the final TypeScript output
//
// This architecture enables:
//   - Clean separation between schema analysis and code generation
//   - Consistent formatting without string manipulation
//   - Parsing existing files for comparison
//
// # AST Node Types
//
//   - File: Root node containing header and child nodes
//   - Namespace: export namespace X { ... }
//   - Interface: export interface X { ... }
//   - TypeAlias: export type X = ...
//   - Enum: export enum X { ... }
//   - Property: Interface property with type, optionality, and description
//   - TypeExpr: Type expressions (StringType, ArrayType, UnionType, etc.)
//
// # Options
//
//   - exportType: "interface" (default) or "type" - controls object export style
//   - readonlyProperties: Add readonly modifier to all properties
//
// # Cross-References
//
// When schemas reference other schemas, the generator uses the ResolvedTree's
// namespace tracking to produce correct reference paths. A schema in namespace
// "Models" referencing "User" in the root will generate "User", while referencing
// "Address" in namespace "Common" will generate "Common.Address".
package typescript
