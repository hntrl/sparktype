package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/hntrl/sparktype/internal/config"
	"github.com/hntrl/sparktype/internal/contents"
	"github.com/hntrl/sparktype/internal/generators"
	"github.com/hntrl/sparktype/internal/spec"

	// Register generators
	_ "github.com/hntrl/sparktype/internal/generators/golang"
	_ "github.com/hntrl/sparktype/internal/generators/python"
	_ "github.com/hntrl/sparktype/internal/generators/typescript"
	_ "github.com/hntrl/sparktype/internal/generators/zod"
)

var (
	configPath string
	watchMode  bool
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate types from OpenAPI specifications",
	Long:  `Generate static types from OpenAPI specifications based on the typegen.jsonc configuration file.`,
	RunE:  runGenerate,
}

func init() {
	generateCmd.Flags().StringVarP(&configPath, "config", "c", "", "path to typegen.jsonc config file")
	generateCmd.Flags().BoolVarP(&watchMode, "watch", "w", false, "watch for changes and regenerate")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Find config file
	cfgPath := configPath
	if cfgPath == "" {
		// Look for default config file in current directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		cfgPath = filepath.Join(cwd, DefaultConfigFileName)
	}

	// Run initial generation
	if err := runGenerateOnce(cfgPath); err != nil {
		if !watchMode {
			return err
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	// If watch mode, start watching
	if watchMode {
		return runWatch(cfgPath)
	}

	return nil
}

func runGenerateOnce(cfgPath string) error {
	// Load configuration
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(cfg); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Get base path for resolving relative paths
	basePath := filepath.Dir(cfgPath)

	// Create spec registry
	registry := spec.NewRegistry(cfg.Specs, basePath)

	// Process each output
	for _, output := range cfg.Outputs {
		if err := processOutput(registry, output, cfg.Options, basePath); err != nil {
			return fmt.Errorf("failed to process output %s: %w", output.Path, err)
		}
		fmt.Printf("Generated: %s\n", output.Path)
	}

	return nil
}

func runWatch(cfgPath string) error {
	fmt.Println("\nWatching for changes... (Press Ctrl+C to stop)")

	// Collect all files to watch
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get spec file paths (only local files can be watched)
	var specFiles []string
	basePath := filepath.Dir(cfgPath)
	for _, specCfg := range cfg.Specs {
		if specCfg.Path != "" {
			specPath := specCfg.Path
			if !filepath.IsAbs(specPath) {
				specPath = filepath.Join(basePath, specPath)
			}
			specFiles = append(specFiles, specPath)
		}
	}

	// Add config file itself
	specFiles = append(specFiles, cfgPath)

	// Get initial modification times
	modTimes := make(map[string]int64)
	for _, f := range specFiles {
		info, err := os.Stat(f)
		if err == nil {
			modTimes[f] = info.ModTime().UnixNano()
		}
	}

	// Poll for changes
	ticker := time.NewTicker(WatchPollInterval)
	defer ticker.Stop()

	for range ticker.C {
		changed := false
		for _, f := range specFiles {
			info, err := os.Stat(f)
			if err != nil {
				continue
			}
			if info.ModTime().UnixNano() != modTimes[f] {
				modTimes[f] = info.ModTime().UnixNano()
				changed = true
				fmt.Printf("\nDetected change in: %s\n", filepath.Base(f))
			}
		}

		if changed {
			fmt.Println("Regenerating...")
			if err := runGenerateOnce(cfgPath); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
		}
	}

	return nil
}

func processOutput(registry *spec.Registry, output config.Output, globalOpts config.Options, basePath string) error {
	// Create content resolver
	resolver := contents.NewResolver(registry)

	// Resolve content tree
	tree, err := resolver.Resolve(output.Contents)
	if err != nil {
		return fmt.Errorf("resolving contents for %s: %w", output.Path, err)
	}

	// Get generator for format
	generator, err := generators.Get(output.Format)
	if err != nil {
		return fmt.Errorf("getting generator for %s (format: %s): %w", output.Path, output.Format, err)
	}

	// Merge options
	opts := generators.Options{
		GlobalOptions: globalOpts,
		OutputOptions: output.Options,
	}

	// Generate output
	content, err := generator.Generate(tree, opts)
	if err != nil {
		return fmt.Errorf("generating %s: %w", output.Path, err)
	}

	// Resolve output path relative to config directory
	outputPath := output.Path
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(basePath, outputPath)
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("creating directory %s: %w", outputDir, err)
	}

	// Write output file
	if err := os.WriteFile(outputPath, content, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", output.Path, err)
	}

	return nil
}
