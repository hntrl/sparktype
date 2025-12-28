package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hntrl/sparktype/internal/config"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a typegen.jsonc configuration file",
	Long:  `Validate a typegen.jsonc configuration file for syntax and semantic errors.`,
	RunE:  runValidate,
}

func init() {
	validateCmd.Flags().StringVarP(&configPath, "config", "c", "", "path to typegen.jsonc config file")
}

func runValidate(cmd *cobra.Command, args []string) error {
	cfgPath := configPath
	if cfgPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		cfgPath = filepath.Join(cwd, "typegen.jsonc")
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("configuration is invalid: %w", err)
	}

	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	fmt.Printf("Configuration is valid: %s\n", cfgPath)
	return nil
}
