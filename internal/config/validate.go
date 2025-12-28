package config

import (
	"fmt"
	"strings"
)

// Validate performs semantic validation on the configuration
func Validate(cfg *Config) error {
	if len(cfg.Specs) == 0 {
		return fmt.Errorf("at least one spec must be defined")
	}

	if len(cfg.Outputs) == 0 {
		return fmt.Errorf("at least one output must be defined")
	}

	// Validate specs
	for name, spec := range cfg.Specs {
		if err := validateSpec(name, spec); err != nil {
			return err
		}
	}

	// Validate outputs
	for i, output := range cfg.Outputs {
		if err := validateOutput(i, output, cfg.Specs); err != nil {
			return err
		}
	}

	return nil
}

func validateSpec(name string, spec Spec) error {
	modes := 0
	if spec.Path != "" {
		modes++
	}
	if spec.URL != "" {
		modes++
	}
	if len(spec.Schemas) > 0 {
		modes++
	}

	if modes == 0 {
		return fmt.Errorf("spec %q must have one of 'path', 'url', or 'schemas' defined", name)
	}
	if modes > 1 {
		return fmt.Errorf("spec %q can only have one of 'path', 'url', or 'schemas' defined", name)
	}

	return nil
}

func validateOutput(index int, output Output, specs map[string]Spec) error {
	if output.Path == "" {
		return fmt.Errorf("output[%d]: 'path' is required", index)
	}

	if output.Format == "" {
		return fmt.Errorf("output[%d]: 'format' is required", index)
	}

	validFormats := []string{"typescript", "zod", "python", "go"}
	formatValid := false
	for _, f := range validFormats {
		if output.Format == f {
			formatValid = true
			break
		}
	}
	if !formatValid {
		return fmt.Errorf("output[%d]: invalid format %q, must be one of: %s",
			index, output.Format, strings.Join(validFormats, ", "))
	}

	if len(output.Contents) == 0 {
		return fmt.Errorf("output[%d]: 'contents' is required and must not be empty", index)
	}

	// Check if namespaces are used with unsupported formats
	supportsNamespaces := output.Format == "typescript" || output.Format == "zod"
	hasNamespaces := containsNamespace(output.Contents)

	if hasNamespaces && !supportsNamespaces {
		return fmt.Errorf("output[%d]: format %q does not support namespaces, use flat contents with only schema patterns",
			index, output.Format)
	}

	// Validate all content items
	if err := validateContents(index, output.Contents, specs, []string{}); err != nil {
		return err
	}

	return nil
}

func containsNamespace(contents []ContentItem) bool {
	for _, item := range contents {
		if item.IsNamespace() {
			return true
		}
	}
	return false
}

func validateContents(outputIndex int, contents []ContentItem, specs map[string]Spec, path []string) error {
	for i, item := range contents {
		itemPath := append(path, fmt.Sprintf("[%d]", i))

		if item.IsPattern() {
			// Validate pattern format: spec:pattern
			if err := validatePattern(outputIndex, item.Pattern, specs, itemPath); err != nil {
				return err
			}
		} else if item.IsNamespace() {
			if item.Namespace.Name == "" {
				return fmt.Errorf("output[%d].contents%s: namespace name is required",
					outputIndex, strings.Join(itemPath, ""))
			}
			// Recursively validate namespace contents
			nsPath := append(itemPath, fmt.Sprintf(".%s", item.Namespace.Name))
			if err := validateContents(outputIndex, item.Namespace.Contents, specs, nsPath); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("output[%d].contents%s: invalid content item",
				outputIndex, strings.Join(itemPath, ""))
		}
	}
	return nil
}

func validatePattern(outputIndex int, pattern string, specs map[string]Spec, path []string) error {
	// Pattern must be in format "spec:pattern"
	parts := strings.SplitN(pattern, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("output[%d].contents%s: invalid pattern %q, must be in format 'spec:pattern' (e.g., 'users:User' or 'users:*Request')",
			outputIndex, strings.Join(path, ""), pattern)
	}

	specName := parts[0]
	if _, exists := specs[specName]; !exists {
		return fmt.Errorf("output[%d].contents%s: spec %q is not defined",
			outputIndex, strings.Join(path, ""), specName)
	}

	return nil
}
