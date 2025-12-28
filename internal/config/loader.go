package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Load reads and parses a typegen.jsonc configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Strip comments from JSONC
	jsonData := stripJSONComments(data)

	// Expand environment variables
	jsonData = expandEnvVars(jsonData)

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(jsonData, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &cfg, nil
}

// stripJSONComments removes single-line (//) and multi-line (/* */) comments from JSONC
func stripJSONComments(data []byte) []byte {
	input := string(data)
	var result strings.Builder
	inString := false
	inSingleComment := false
	inMultiComment := false
	i := 0

	for i < len(input) {
		// Handle string literals (don't process comments inside strings)
		if !inSingleComment && !inMultiComment && input[i] == '"' {
			// Check if it's an escaped quote
			escapeCount := 0
			for j := i - 1; j >= 0 && input[j] == '\\'; j-- {
				escapeCount++
			}
			if escapeCount%2 == 0 {
				inString = !inString
			}
			result.WriteByte(input[i])
			i++
			continue
		}

		if inString {
			result.WriteByte(input[i])
			i++
			continue
		}

		// Handle single-line comment start
		if !inMultiComment && i+1 < len(input) && input[i] == '/' && input[i+1] == '/' {
			inSingleComment = true
			i += 2
			continue
		}

		// Handle single-line comment end
		if inSingleComment && input[i] == '\n' {
			inSingleComment = false
			result.WriteByte(input[i])
			i++
			continue
		}

		// Handle multi-line comment start
		if !inSingleComment && i+1 < len(input) && input[i] == '/' && input[i+1] == '*' {
			inMultiComment = true
			i += 2
			continue
		}

		// Handle multi-line comment end
		if inMultiComment && i+1 < len(input) && input[i] == '*' && input[i+1] == '/' {
			inMultiComment = false
			i += 2
			continue
		}

		// Skip content inside comments
		if inSingleComment || inMultiComment {
			i++
			continue
		}

		result.WriteByte(input[i])
		i++
	}

	return []byte(result.String())
}

// expandEnvVars replaces ${VAR_NAME} patterns with environment variable values
func expandEnvVars(data []byte) []byte {
	envPattern := regexp.MustCompile(`\$\{([^}]+)\}`)
	return envPattern.ReplaceAllFunc(data, func(match []byte) []byte {
		// Extract variable name from ${VAR_NAME}
		varName := string(match[2 : len(match)-1])
		value := os.Getenv(varName)
		return []byte(value)
	})
}

// UnmarshalJSON implements custom unmarshaling for ContentItem
// to handle both string (pattern) and object (namespace) formats
func (c *ContentItem) UnmarshalJSON(data []byte) error {
	// Try as string first (schema pattern)
	var pattern string
	if err := json.Unmarshal(data, &pattern); err == nil {
		c.Pattern = pattern
		c.Namespace = nil
		return nil
	}

	// Try as namespace object
	var ns NamespaceDef
	if err := json.Unmarshal(data, &ns); err == nil {
		c.Pattern = ""
		c.Namespace = &ns
		return nil
	}

	return fmt.Errorf("content item must be a string (schema pattern) or object (namespace definition)")
}

// MarshalJSON implements custom marshaling for ContentItem
func (c ContentItem) MarshalJSON() ([]byte, error) {
	if c.Pattern != "" {
		return json.Marshal(c.Pattern)
	}
	if c.Namespace != nil {
		return json.Marshal(c.Namespace)
	}
	return []byte("null"), nil
}
