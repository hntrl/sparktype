// Package cli provides the command-line interface for sparktype.
//
// The CLI is built using Cobra and provides the following commands:
//
//   - generate: Generate types from OpenAPI specifications based on typegen.jsonc
//   - check: Compare existing generated files with current specs to detect drift
//   - validate: Validate a typegen.jsonc configuration file
//   - init: Create a new typegen.jsonc configuration file
//
// # Usage
//
// The primary workflow is:
//
//  1. Create a typegen.jsonc configuration file (or use 'sparktype init')
//  2. Run 'sparktype generate' to produce type definitions
//  3. Optionally use 'sparktype check' in CI to ensure types stay in sync
//
// # Configuration Discovery
//
// Commands that require a configuration file will look for typegen.jsonc in the
// current directory by default. Use the -c or --config flag to specify an
// alternate path.
//
// # Watch Mode
//
// The generate command supports a --watch flag that monitors spec files and the
// configuration for changes, automatically regenerating output files when
// modifications are detected.
package cli
