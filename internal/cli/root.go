package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sparktype",
	Short: "Generate static types from OpenAPI specifications",
	Long: `Sparktype generates static types from OpenAPI specifications.

It supports multiple output formats including TypeScript interfaces,
Zod schemas, Python TypedDicts, and Go structs.`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(checkCmd)
}
