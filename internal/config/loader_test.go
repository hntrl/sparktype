package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		file     string
		wantErr  bool
		validate func(*testing.T, *Config)
	}{
		{
			name: "simple valid config",
			file: "valid_simple.jsonc",
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Specs) != 1 {
					t.Errorf("expected 1 spec, got %d", len(cfg.Specs))
				}
				if _, ok := cfg.Specs["users"]; !ok {
					t.Error("expected 'users' spec")
				}
				if len(cfg.Outputs) != 1 {
					t.Errorf("expected 1 output, got %d", len(cfg.Outputs))
				}
			},
		},
		{
			name: "complex valid config",
			file: "valid_complex.jsonc",
			validate: func(t *testing.T, cfg *Config) {
				if len(cfg.Specs) != 3 {
					t.Errorf("expected 3 specs, got %d", len(cfg.Specs))
				}
				if len(cfg.Outputs) != 2 {
					t.Errorf("expected 2 outputs, got %d", len(cfg.Outputs))
				}
				// Check nested namespace structure
				output := cfg.Outputs[0]
				if len(output.Contents) != 3 {
					t.Errorf("expected 3 content items, got %d", len(output.Contents))
				}
				// Third item should be a namespace
				if !output.Contents[2].IsNamespace() {
					t.Error("expected third content item to be a namespace")
				}
				ns := output.Contents[2].Namespace
				if ns.Name != "Products" {
					t.Errorf("expected namespace name 'Products', got %q", ns.Name)
				}
			},
		},
		{
			name:    "nonexistent file",
			file:    "nonexistent.jsonc",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.file)
			cfg, err := Load(path)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables for the test
	os.Setenv("API_BASE_URL", "https://api.example.com")
	os.Setenv("API_KEY", "secret-key-123")
	os.Setenv("OUTPUT_DIR", "./generated")
	defer func() {
		os.Unsetenv("API_BASE_URL")
		os.Unsetenv("API_KEY")
		os.Unsetenv("OUTPUT_DIR")
	}()

	cfg, err := Load(filepath.Join("testdata", "valid_env_vars.jsonc"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check URL expansion
	spec := cfg.Specs["api"]
	expectedURL := "https://api.example.com/openapi.json"
	if spec.URL != expectedURL {
		t.Errorf("expected URL %q, got %q", expectedURL, spec.URL)
	}

	// Check header expansion
	expectedKey := "secret-key-123"
	if spec.Headers["X-Api-Key"] != expectedKey {
		t.Errorf("expected header value %q, got %q", expectedKey, spec.Headers["X-Api-Key"])
	}

	// Check output path expansion
	expectedPath := "./generated/types.ts"
	if cfg.Outputs[0].Path != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, cfg.Outputs[0].Path)
	}
}

func TestStripJSONComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no comments",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "single line comment",
			input:    "{\n  // comment\n  \"key\": \"value\"\n}",
			expected: "{\n  \n  \"key\": \"value\"\n}",
		},
		{
			name:     "multi-line comment",
			input:    "{\n  /* multi\n     line */\n  \"key\": \"value\"\n}",
			expected: "{\n  \n  \"key\": \"value\"\n}",
		},
		{
			name:     "comment-like content in string",
			input:    `{"url": "https://example.com"}`,
			expected: `{"url": "https://example.com"}`,
		},
		{
			name:     "double slash in string preserved",
			input:    `{"comment": "// not a comment"}`,
			expected: `{"comment": "// not a comment"}`,
		},
		{
			name:     "inline comment at end of line",
			input:    "{\"key\": \"value\" // inline comment\n}",
			expected: "{\"key\": \"value\" \n}",
		},
		{
			name:     "trailing comma support prep",
			input:    "{\n  \"a\": 1,\n  // comment\n}",
			expected: "{\n  \"a\": 1,\n  \n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripJSONComments([]byte(tt.input))
			if string(result) != tt.expected {
				t.Errorf("stripJSONComments(%q)\ngot:  %q\nwant: %q", tt.input, string(result), tt.expected)
			}
		})
	}
}

func TestExpandEnvVars(t *testing.T) {
	// Set test environment variables
	os.Setenv("TEST_VAR", "test-value")
	os.Setenv("ANOTHER_VAR", "another-value")
	defer func() {
		os.Unsetenv("TEST_VAR")
		os.Unsetenv("ANOTHER_VAR")
	}()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no env vars",
			input:    `{"key": "value"}`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "single env var",
			input:    `{"key": "${TEST_VAR}"}`,
			expected: `{"key": "test-value"}`,
		},
		{
			name:     "multiple env vars",
			input:    `{"a": "${TEST_VAR}", "b": "${ANOTHER_VAR}"}`,
			expected: `{"a": "test-value", "b": "another-value"}`,
		},
		{
			name:     "env var in middle of string",
			input:    `{"url": "https://${TEST_VAR}.example.com"}`,
			expected: `{"url": "https://test-value.example.com"}`,
		},
		{
			name:     "undefined env var becomes empty",
			input:    `{"key": "${UNDEFINED_VAR}"}`,
			expected: `{"key": ""}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandEnvVars([]byte(tt.input))
			if string(result) != tt.expected {
				t.Errorf("expandEnvVars(%q)\ngot:  %q\nwant: %q", tt.input, string(result), tt.expected)
			}
		})
	}
}

func TestContentItemUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantPattern string
		wantNS      bool
		wantNSName  string
		wantErr     bool
	}{
		{
			name:        "string pattern",
			input:       `"users:*"`,
			wantPattern: "users:*",
			wantNS:      false,
		},
		{
			name:       "namespace object",
			input:      `{"namespace": "Models", "contents": ["api:*"]}`,
			wantNS:     true,
			wantNSName: "Models",
		},
		{
			name:       "nested namespace",
			input:      `{"namespace": "Outer", "contents": [{"namespace": "Inner", "contents": ["api:*"]}]}`,
			wantNS:     true,
			wantNSName: "Outer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var item ContentItem
			err := json.Unmarshal([]byte(tt.input), &item)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.wantPattern != "" {
				if item.Pattern != tt.wantPattern {
					t.Errorf("expected pattern %q, got %q", tt.wantPattern, item.Pattern)
				}
				if item.Namespace != nil {
					t.Error("expected Namespace to be nil for pattern")
				}
			}

			if tt.wantNS {
				if item.Namespace == nil {
					t.Fatal("expected Namespace to be non-nil")
				}
				if item.Namespace.Name != tt.wantNSName {
					t.Errorf("expected namespace name %q, got %q", tt.wantNSName, item.Namespace.Name)
				}
				if item.Pattern != "" {
					t.Error("expected Pattern to be empty for namespace")
				}
			}
		})
	}
}

func TestContentItemMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		item     ContentItem
		expected string
	}{
		{
			name:     "pattern",
			item:     ContentItem{Pattern: "users:*"},
			expected: `"users:*"`,
		},
		{
			name: "namespace",
			item: ContentItem{
				Namespace: &NamespaceDef{
					Name:     "Models",
					Contents: []ContentItem{{Pattern: "api:*"}},
				},
			},
			expected: `{"namespace":"Models","contents":["api:*"]}`,
		},
		{
			name:     "empty item",
			item:     ContentItem{},
			expected: `null`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := json.Marshal(tt.item)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("MarshalJSON()\ngot:  %s\nwant: %s", string(result), tt.expected)
			}
		})
	}
}

func TestSpecMethods(t *testing.T) {
	t.Run("IsLocal", func(t *testing.T) {
		spec := Spec{Path: "./api.yaml"}
		if !spec.IsLocal() {
			t.Error("expected IsLocal() to return true")
		}
		if spec.IsRemote() {
			t.Error("expected IsRemote() to return false")
		}
		if spec.IsInline() {
			t.Error("expected IsInline() to return false")
		}
	})

	t.Run("IsRemote", func(t *testing.T) {
		spec := Spec{URL: "https://example.com/api.json"}
		if spec.IsLocal() {
			t.Error("expected IsLocal() to return false")
		}
		if !spec.IsRemote() {
			t.Error("expected IsRemote() to return true")
		}
		if spec.IsInline() {
			t.Error("expected IsInline() to return false")
		}
	})

	t.Run("IsInline", func(t *testing.T) {
		spec := Spec{Schemas: map[string]any{"User": map[string]any{"type": "object"}}}
		if spec.IsLocal() {
			t.Error("expected IsLocal() to return false")
		}
		if spec.IsRemote() {
			t.Error("expected IsRemote() to return false")
		}
		if !spec.IsInline() {
			t.Error("expected IsInline() to return true")
		}
	})

	t.Run("GetSource", func(t *testing.T) {
		tests := []struct {
			spec     Spec
			expected string
		}{
			{Spec{Path: "./api.yaml"}, "./api.yaml"},
			{Spec{URL: "https://example.com/api.json"}, "https://example.com/api.json"},
			{Spec{Schemas: map[string]any{"User": nil}}, "inline"},
		}
		for _, tt := range tests {
			if got := tt.spec.GetSource(); got != tt.expected {
				t.Errorf("GetSource() = %q, want %q", got, tt.expected)
			}
		}
	})
}

func TestContentItemMethods(t *testing.T) {
	t.Run("IsPattern", func(t *testing.T) {
		item := ContentItem{Pattern: "users:*"}
		if !item.IsPattern() {
			t.Error("expected IsPattern() to return true")
		}
		if item.IsNamespace() {
			t.Error("expected IsNamespace() to return false")
		}
	})

	t.Run("IsNamespace", func(t *testing.T) {
		item := ContentItem{Namespace: &NamespaceDef{Name: "Models"}}
		if item.IsPattern() {
			t.Error("expected IsPattern() to return false")
		}
		if !item.IsNamespace() {
			t.Error("expected IsNamespace() to return true")
		}
	})
}
