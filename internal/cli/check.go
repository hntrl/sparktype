package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/hntrl/sparktype/internal/config"
	"github.com/hntrl/sparktype/internal/contents"
	"github.com/hntrl/sparktype/internal/generators"
	"github.com/hntrl/sparktype/internal/spec"
)

var checkConfigPath string

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check if generated types are in sync with OpenAPI specs",
	Long: `Check compares existing generated files with what would be generated
from the current OpenAPI specifications. Exits with code 1 if any files
are out of sync.

This is useful for CI/CD pipelines to ensure generated types are kept
up to date with the source specifications.`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().StringVarP(&checkConfigPath, "config", "c", "", "path to typegen.jsonc config file")
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Find config file
	cfgPath := checkConfigPath
	if cfgPath == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		cfgPath = filepath.Join(cwd, "typegen.jsonc")
	}

	// Load configuration
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Create spec registry
	registry := spec.NewRegistry(cfg.Specs, filepath.Dir(cfgPath))

	// Track mismatches
	var mismatches int

	// Process each output
	for _, output := range cfg.Outputs {
		result, err := checkOutput(registry, output, cfg.Options)
		if err != nil {
			fmt.Printf("Checking %s... ERROR\n", output.Path)
			fmt.Printf("  %v\n", err)
			mismatches++
			continue
		}

		if result == nil {
			// File doesn't exist
			fmt.Printf("Checking %s... MISSING\n", output.Path)
			mismatches++
			continue
		}

		if result.Match {
			fmt.Printf("Checking %s... OK\n", output.Path)
		} else {
			fmt.Printf("Checking %s... MISMATCH\n", output.Path)
			fmt.Print(formatDiff(result))
			mismatches++
		}
	}

	// Print summary
	fmt.Println()
	if mismatches > 0 {
		fmt.Printf("%d file(s) out of sync. Run 'sparktype generate' to update.\n", mismatches)
		os.Exit(1)
	}

	fmt.Println("All files are in sync.")
	return nil
}

func checkOutput(registry *spec.Registry, output config.Output, globalOpts config.Options) (*generators.CompareResult, error) {
	// Read existing file
	existing, err := os.ReadFile(output.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // File doesn't exist
		}
		return nil, fmt.Errorf("reading %s: %w", output.Path, err)
	}

	// Create content resolver
	resolver := contents.NewResolver(registry)

	// Resolve content tree
	tree, err := resolver.Resolve(output.Contents)
	if err != nil {
		return nil, fmt.Errorf("resolving contents for %s: %w", output.Path, err)
	}

	// Get generator
	generator, err := generators.Get(output.Format)
	if err != nil {
		return nil, fmt.Errorf("getting generator for %s (format: %s): %w", output.Path, output.Format, err)
	}

	// Merge options
	opts := generators.Options{
		GlobalOptions: globalOpts,
		OutputOptions: output.Options,
	}

	// Compare
	return generator.Compare(existing, tree, opts)
}

func formatDiff(result *generators.CompareResult) string {
	var sb strings.Builder

	for _, td := range result.Types {
		switch td.Status {
		case generators.Added:
			sb.WriteString("  + ")
			sb.WriteString(td.Name)
			sb.WriteString(" (added)\n")
		case generators.Removed:
			sb.WriteString("  - ")
			sb.WriteString(td.Name)
			sb.WriteString(" (removed)\n")
		case generators.Changed:
			sb.WriteString("  ")
			sb.WriteString(td.Name)
			sb.WriteString(":\n")
			for _, pd := range td.Properties {
				switch pd.Status {
				case generators.Added:
					sb.WriteString("    + ")
					sb.WriteString(pd.Name)
					sb.WriteString(": ")
					sb.WriteString(pd.NewValue)
					sb.WriteString(" (added)\n")
				case generators.Removed:
					sb.WriteString("    - ")
					sb.WriteString(pd.Name)
					sb.WriteString(": ")
					sb.WriteString(pd.OldValue)
					sb.WriteString(" (removed)\n")
				case generators.Changed:
					sb.WriteString("    ~ ")
					sb.WriteString(pd.Name)
					sb.WriteString(": ")
					sb.WriteString(pd.OldValue)
					sb.WriteString(" -> ")
					sb.WriteString(pd.NewValue)
					sb.WriteString("\n")
				}
			}
		}
	}

	return sb.String()
}
