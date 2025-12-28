// Package zod generates Zod schema definitions with optional TypeScript interfaces
// from OpenAPI schemas.
//
// # Output Format
//
// The generator produces:
//
//   - Zod schema declarations (const userSchema = z.object({...}))
//   - Inferred type exports (export type User = z.infer<typeof userSchema>)
//   - Optional TypeScript interfaces with JSDoc transfer (when inferTypes is enabled)
//   - Namespace objects for hierarchical organization
//
// # AST Architecture
//
// Like the TypeScript generator, this uses an AST-based approach:
//
//  1. BuildFile constructs a File AST from the ResolvedTree
//  2. Schemas section contains Zod schema declarations
//  3. TypeSection contains inferred type definitions
//  4. Interfaces section (optional) contains TypeScript interfaces
//
// # AST Node Types
//
// Schema nodes:
//   - SchemaDecl: Top-level const schema = z.xxx() declaration
//   - SchemaProperty: Property within a namespace object
//   - Namespace: Namespace object containing schema properties
//   - NestedNamespace: Nested namespace within another namespace
//
// Zod expression nodes:
//   - ZObject, ZArray, ZString, ZNumber, ZBoolean, ZEnum, ZUnion, etc.
//   - ZLazy: For circular references (z.lazy(() => schemaRef))
//   - ZWithModifiers: Wraps expressions with .optional(), .nullable(), etc.
//
// Type nodes:
//   - InferredType: z.infer<typeof schema> type alias
//   - Interface: TypeScript interface with indexed property types
//   - TypeAlias: Simple type alias
//   - TSNamespace: TypeScript namespace for interface organization
//
// # Options
//
//   - inferTypes: When true, generates internal schema types and public interfaces
//     with JSDoc comments. Properties use indexed types (SchemaType["prop"]) to
//     reference the schema's inferred type while adding documentation.
//
// # Modifiers
//
// Zod expressions support modifiers that chain method calls:
//   - Optional, Nullable: .optional(), .nullable()
//   - Min, Max: .min(n), .max(n)
//   - Email, Url, Uuid, Datetime, Date: String format validators
//   - Regex: .regex(/pattern/)
//   - Int: .int() for integer numbers
package zod
