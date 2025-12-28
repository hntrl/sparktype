package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name       string
		file       string
		wantErr    bool
		errContain string
	}{
		{
			name: "valid simple config",
			file: "valid_simple.jsonc",
		},
		{
			name: "valid complex config",
			file: "valid_complex.jsonc",
		},
		{
			name:       "no specs",
			file:       "invalid_no_specs.jsonc",
			wantErr:    true,
			errContain: "at least one spec must be defined",
		},
		{
			name:       "no outputs",
			file:       "invalid_no_outputs.jsonc",
			wantErr:    true,
			errContain: "at least one output must be defined",
		},
		{
			name:       "spec with multiple sources",
			file:       "invalid_spec_multi_source.jsonc",
			wantErr:    true,
			errContain: "can only have one of",
		},
		{
			name:       "spec with no source",
			file:       "invalid_spec_no_source.jsonc",
			wantErr:    true,
			errContain: "must have one of",
		},
		{
			name:       "invalid format",
			file:       "invalid_format.jsonc",
			wantErr:    true,
			errContain: "invalid format",
		},
		{
			name:       "invalid pattern format",
			file:       "invalid_pattern.jsonc",
			wantErr:    true,
			errContain: "must be in format 'spec:pattern'",
		},
		{
			name:       "namespace in python",
			file:       "invalid_namespace_python.jsonc",
			wantErr:    true,
			errContain: "does not support namespaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Load(filepath.Join("testdata", tt.file))
			if err != nil {
				t.Fatalf("failed to load config: %v", err)
			}

			err = Validate(cfg)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateSpec(t *testing.T) {
	tests := []struct {
		name       string
		specName   string
		spec       Spec
		wantErr    bool
		errContain string
	}{
		{
			name:     "local spec",
			specName: "api",
			spec:     Spec{Path: "./api.yaml"},
		},
		{
			name:     "remote spec",
			specName: "api",
			spec:     Spec{URL: "https://example.com/api.json"},
		},
		{
			name:     "inline spec",
			specName: "api",
			spec:     Spec{Schemas: map[string]any{"User": nil}},
		},
		{
			name:       "no source",
			specName:   "api",
			spec:       Spec{},
			wantErr:    true,
			errContain: "must have one of",
		},
		{
			name:     "path and url",
			specName: "api",
			spec: Spec{
				Path: "./api.yaml",
				URL:  "https://example.com/api.json",
			},
			wantErr:    true,
			errContain: "can only have one of",
		},
		{
			name:     "path and schemas",
			specName: "api",
			spec: Spec{
				Path:    "./api.yaml",
				Schemas: map[string]any{"User": nil},
			},
			wantErr:    true,
			errContain: "can only have one of",
		},
		{
			name:     "url and schemas",
			specName: "api",
			spec: Spec{
				URL:     "https://example.com/api.json",
				Schemas: map[string]any{"User": nil},
			},
			wantErr:    true,
			errContain: "can only have one of",
		},
		{
			name:     "all three",
			specName: "api",
			spec: Spec{
				Path:    "./api.yaml",
				URL:     "https://example.com/api.json",
				Schemas: map[string]any{"User": nil},
			},
			wantErr:    true,
			errContain: "can only have one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSpec(tt.specName, tt.spec)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateOutput(t *testing.T) {
	specs := map[string]Spec{
		"users":    {Path: "./users.yaml"},
		"products": {Path: "./products.yaml"},
	}

	tests := []struct {
		name       string
		output     Output
		wantErr    bool
		errContain string
	}{
		{
			name: "valid typescript output",
			output: Output{
				Path:     "./types.ts",
				Format:   "typescript",
				Contents: []ContentItem{{Pattern: "users:*"}},
			},
		},
		{
			name: "valid zod output",
			output: Output{
				Path:     "./schemas.ts",
				Format:   "zod",
				Contents: []ContentItem{{Pattern: "users:*"}},
			},
		},
		{
			name: "valid python output",
			output: Output{
				Path:     "./types.py",
				Format:   "python",
				Contents: []ContentItem{{Pattern: "users:*"}},
			},
		},
		{
			name: "valid go output",
			output: Output{
				Path:     "./types.go",
				Format:   "go",
				Contents: []ContentItem{{Pattern: "users:*"}},
			},
		},
		{
			name: "missing path",
			output: Output{
				Format:   "typescript",
				Contents: []ContentItem{{Pattern: "users:*"}},
			},
			wantErr:    true,
			errContain: "'path' is required",
		},
		{
			name: "missing format",
			output: Output{
				Path:     "./types.ts",
				Contents: []ContentItem{{Pattern: "users:*"}},
			},
			wantErr:    true,
			errContain: "'format' is required",
		},
		{
			name: "missing contents",
			output: Output{
				Path:   "./types.ts",
				Format: "typescript",
			},
			wantErr:    true,
			errContain: "'contents' is required",
		},
		{
			name: "empty contents",
			output: Output{
				Path:     "./types.ts",
				Format:   "typescript",
				Contents: []ContentItem{},
			},
			wantErr:    true,
			errContain: "'contents' is required",
		},
		{
			name: "typescript with namespace",
			output: Output{
				Path:   "./types.ts",
				Format: "typescript",
				Contents: []ContentItem{
					{Namespace: &NamespaceDef{Name: "Models", Contents: []ContentItem{{Pattern: "users:*"}}}},
				},
			},
		},
		{
			name: "zod with namespace",
			output: Output{
				Path:   "./schemas.ts",
				Format: "zod",
				Contents: []ContentItem{
					{Namespace: &NamespaceDef{Name: "Models", Contents: []ContentItem{{Pattern: "users:*"}}}},
				},
			},
		},
		{
			name: "python with namespace",
			output: Output{
				Path:   "./types.py",
				Format: "python",
				Contents: []ContentItem{
					{Namespace: &NamespaceDef{Name: "Models", Contents: []ContentItem{{Pattern: "users:*"}}}},
				},
			},
			wantErr:    true,
			errContain: "does not support namespaces",
		},
		{
			name: "go with namespace",
			output: Output{
				Path:   "./types.go",
				Format: "go",
				Contents: []ContentItem{
					{Namespace: &NamespaceDef{Name: "Models", Contents: []ContentItem{{Pattern: "users:*"}}}},
				},
			},
			wantErr:    true,
			errContain: "does not support namespaces",
		},
		{
			name: "undefined spec reference",
			output: Output{
				Path:     "./types.ts",
				Format:   "typescript",
				Contents: []ContentItem{{Pattern: "undefined:*"}},
			},
			wantErr:    true,
			errContain: "is not defined",
		},
		{
			name: "nested namespace with valid pattern",
			output: Output{
				Path:   "./types.ts",
				Format: "typescript",
				Contents: []ContentItem{
					{
						Namespace: &NamespaceDef{
							Name: "Outer",
							Contents: []ContentItem{
								{
									Namespace: &NamespaceDef{
										Name:     "Inner",
										Contents: []ContentItem{{Pattern: "users:*"}},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOutput(0, tt.output, specs)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidatePattern(t *testing.T) {
	specs := map[string]Spec{
		"users":    {Path: "./users.yaml"},
		"products": {Path: "./products.yaml"},
	}

	tests := []struct {
		name       string
		pattern    string
		wantErr    bool
		errContain string
	}{
		{
			name:    "simple pattern",
			pattern: "users:User",
		},
		{
			name:    "wildcard pattern",
			pattern: "users:*",
		},
		{
			name:    "prefix pattern",
			pattern: "users:*Request",
		},
		{
			name:    "suffix pattern",
			pattern: "products:Product*",
		},
		{
			name:       "missing colon",
			pattern:    "users",
			wantErr:    true,
			errContain: "must be in format 'spec:pattern'",
		},
		{
			name:       "undefined spec",
			pattern:    "orders:*",
			wantErr:    true,
			errContain: "is not defined",
		},
		{
			name:    "pattern with colon in glob",
			pattern: "users:Type:A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePattern(0, tt.pattern, specs, []string{})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errContain) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestContainsNamespace(t *testing.T) {
	tests := []struct {
		name     string
		contents []ContentItem
		expected bool
	}{
		{
			name:     "empty contents",
			contents: []ContentItem{},
			expected: false,
		},
		{
			name: "only patterns",
			contents: []ContentItem{
				{Pattern: "users:*"},
				{Pattern: "products:*"},
			},
			expected: false,
		},
		{
			name: "contains namespace",
			contents: []ContentItem{
				{Pattern: "users:*"},
				{Namespace: &NamespaceDef{Name: "Models"}},
			},
			expected: true,
		},
		{
			name: "only namespace",
			contents: []ContentItem{
				{Namespace: &NamespaceDef{Name: "Models"}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsNamespace(tt.contents)
			if result != tt.expected {
				t.Errorf("containsNamespace() = %v, want %v", result, tt.expected)
			}
		})
	}
}

