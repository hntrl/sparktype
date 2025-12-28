// Package generators provides the interface and registry for type generators.
//
// A generator converts a resolved content tree (containing OpenAPI schemas organized
// by namespace) into language-specific type definitions. Each generator follows a
// consistent architecture:
//
//  1. Build an AST representation of the target language's type constructs
//  2. Serialize the AST to produce the final output
//  3. Parse existing files back into AST for comparison (used by 'check' command)
//
// # Generator Interface
//
// Each generator provides two functions:
//
//   - Generate: Produces output bytes from a ResolvedTree
//   - Compare: Compares existing file content against expected output, returning detailed diffs
//
// # Available Generators
//
// The following generators are registered via init():
//
//   - "typescript": TypeScript interfaces or type aliases
//   - "zod": Zod schemas with optional TypeScript interface generation
//   - "python": Python TypedDict classes
//   - "golang": Go struct definitions
//
// # Options
//
// Generators receive both global options (from config.Options) and output-specific
// options (from the output's options map). Output options take precedence, allowing
// per-file customization while maintaining sensible defaults.
//
// # Registration
//
// Generators register themselves using Register() in their init() functions.
// The CLI imports generator packages with blank imports to trigger registration:
//
//	import _ "github.com/hntrl/sparktype/internal/generators/typescript"
//
// # Comparison Results
//
// The Compare function returns a CompareResult containing:
//
//   - Match: Whether the existing and expected outputs are equivalent
//   - Types: Detailed diffs showing added, removed, and changed type definitions
//   - Properties: For changed types, property-level diffs showing what changed
package generators
