package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new typegen.jsonc configuration file",
	Long:  `Create a new typegen.jsonc configuration file with example settings.`,
	RunE:  runInit,
}

func runInit(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	cfgPath := filepath.Join(cwd, "typegen.jsonc")

	// Check if file already exists
	if _, err := os.Stat(cfgPath); err == nil {
		return fmt.Errorf("typegen.jsonc already exists in current directory")
	}

	template := `{
  "$schema": "https://hntrl.github.io/sparktype/schema.json",

  // Named spec sources - supports multiple specs
  "specs": {
    "api": {
      "path": "./openapi.yaml"
    }
  },

  // Output configurations
  "outputs": [
    {
      "path": "./src/types/api.ts",
      "format": "typescript",
      "spec": "api"
    }
  ],

  // Global options
  "options": {
    "dereferenceRefs": true,
    "generateEnums": true,
    "nullableHandling": "optional"
  }
}
`

	if err := os.WriteFile(cfgPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("Created: %s\n", cfgPath)
	return nil
}
