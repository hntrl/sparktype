// Package contents resolves configuration content items into a hierarchical tree
// of schemas organized by namespace.
//
// This package is the bridge between configuration (which uses spec:pattern syntax
// to select schemas) and generators (which need a resolved tree of actual schemas).
//
// # Content Resolution
//
// The Resolver takes ContentItem configurations and produces a ResolvedTree by:
//
//  1. Parsing spec:pattern strings to identify which spec and which schemas to include
//  2. Using glob matching to filter schemas by name pattern (e.g., "*Request", "User*")
//  3. Building a hierarchical tree that preserves namespace structure from the config
//  4. Tracking namespace paths for each schema to enable cross-reference resolution
//
// # Pattern Syntax
//
// Schema patterns use the format "specName:globPattern":
//
//   - "users:*" - All schemas from the "users" spec
//   - "products:*Request" - Schemas ending in "Request" from "products" spec
//   - "api:User" - Exact match for "User" schema from "api" spec
//
// # Namespace Tracking
//
// The ResolvedTree maintains a SchemaNamespaces map that records the namespace path
// for each schema. This enables generators to produce correct cross-references when
// one schema in a namespace references another schema in a different namespace.
//
// For example, if "User" is in namespace "Models" and "Order" references "User",
// the generator can produce "Models.User" as the reference path.
package contents
