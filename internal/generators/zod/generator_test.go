package zod

import (
	"testing"

	"github.com/sebdah/goldie/v2"

	"github.com/hntrl/sparktype/internal/contents"
	"github.com/hntrl/sparktype/internal/generators"
	"github.com/hntrl/sparktype/internal/spec"
)

func TestGenerate_SimpleObject(t *testing.T) {
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Schema: &spec.Schema{
				Name: "User",
				Type: "object",
				Properties: []spec.Property{
					{Name: "id", Schema: spec.Schema{Type: "integer"}, Required: true},
					{Name: "name", Schema: spec.Schema{Type: "string"}, Required: true},
				},
			}},
		},
		SchemaNamespaces: map[string][]string{},
	}

	result, err := Generate(tree, generators.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	g.Assert(t, "simple_object", result)
}

func TestGenerate_WithDescription(t *testing.T) {
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Schema: &spec.Schema{
				Name:        "User",
				Type:        "object",
				Description: "A user account in the system",
				Properties: []spec.Property{
					{Name: "id", Schema: spec.Schema{Type: "integer", Description: "Unique identifier"}, Required: true},
					{Name: "name", Schema: spec.Schema{Type: "string", Description: "User's display name"}, Required: true},
				},
			}},
		},
		SchemaNamespaces: map[string][]string{},
	}

	result, err := Generate(tree, generators.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	g.Assert(t, "with_description", result)
}

func TestGenerate_OptionalFields(t *testing.T) {
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Schema: &spec.Schema{
				Name: "User",
				Type: "object",
				Properties: []spec.Property{
					{Name: "id", Schema: spec.Schema{Type: "integer"}, Required: true},
					{Name: "name", Schema: spec.Schema{Type: "string"}, Required: false},
					{Name: "email", Schema: spec.Schema{Type: "string"}, Required: false},
				},
			}},
		},
		SchemaNamespaces: map[string][]string{},
	}

	result, err := Generate(tree, generators.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	g.Assert(t, "optional_fields", result)
}

func TestGenerate_StringEnum(t *testing.T) {
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Schema: &spec.Schema{
				Name: "Status",
				Type: "string",
				Enum: []any{"active", "inactive", "pending"},
			}},
		},
		SchemaNamespaces: map[string][]string{},
	}

	result, err := Generate(tree, generators.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	g.Assert(t, "string_enum", result)
}

func TestGenerate_ArrayType(t *testing.T) {
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Schema: &spec.Schema{
				Name: "Tags",
				Type: "array",
				Items: &spec.Schema{
					Type: "string",
				},
			}},
		},
		SchemaNamespaces: map[string][]string{},
	}

	result, err := Generate(tree, generators.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	g.Assert(t, "array_type", result)
}

func TestGenerate_Namespace(t *testing.T) {
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Namespace: &contents.NamespaceNode{
				Name: "Models",
				Children: []contents.Node{
					{Schema: &spec.Schema{
						Name: "User",
						Type: "object",
						Properties: []spec.Property{
							{Name: "id", Schema: spec.Schema{Type: "integer"}, Required: true},
							{Name: "name", Schema: spec.Schema{Type: "string"}, Required: true},
						},
					}},
				},
			}},
		},
		SchemaNamespaces: map[string][]string{
			"User": {"Models"},
		},
	}

	result, err := Generate(tree, generators.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	g.Assert(t, "namespace", result)
}

func TestGenerate_Nullable(t *testing.T) {
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Schema: &spec.Schema{
				Name: "User",
				Type: "object",
				Properties: []spec.Property{
					{Name: "id", Schema: spec.Schema{Type: "integer"}, Required: true},
					{Name: "name", Schema: spec.Schema{Type: "string", Nullable: true}, Required: true},
				},
			}},
		},
		SchemaNamespaces: map[string][]string{},
	}

	result, err := Generate(tree, generators.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	g.Assert(t, "nullable", result)
}

func TestGenerate_InferTypes(t *testing.T) {
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Schema: &spec.Schema{
				Name:        "User",
				Type:        "object",
				Description: "A user account",
				Properties: []spec.Property{
					{Name: "id", Schema: spec.Schema{Type: "integer", Description: "Unique identifier"}, Required: true},
					{Name: "name", Schema: spec.Schema{Type: "string", Description: "User's display name"}, Required: true},
				},
			}},
		},
		SchemaNamespaces: map[string][]string{},
	}

	result, err := Generate(tree, generators.Options{
		OutputOptions: map[string]any{
			"inferTypes": true,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	g.Assert(t, "infer_types", result)
}

func TestGenerate_StringConstraints(t *testing.T) {
	minLen := 3
	maxLen := 50
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Schema: &spec.Schema{
				Name:      "Email",
				Type:      "string",
				Format:    "email",
				MinLength: &minLen,
				MaxLength: &maxLen,
			}},
		},
		SchemaNamespaces: map[string][]string{},
	}

	result, err := Generate(tree, generators.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	g.Assert(t, "string_constraints", result)
}

// Test utility functions
func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"User", "user"},
		{"UserProfile", "userProfile"},
		{"ABC", "aBC"},
		{"ID", "iD"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toCamelCase(tt.input)
			if result != tt.expected {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExtractRefName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"#/components/schemas/User", "User"},
		{"User", "User"},
		{"models/User", "User"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := extractRefName(tt.input)
			if result != tt.expected {
				t.Errorf("extractRefName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAllStrings(t *testing.T) {
	tests := []struct {
		values   []any
		expected bool
	}{
		{[]any{"a", "b", "c"}, true},
		{[]any{"a", 1, "c"}, false},
		{[]any{}, true},
		{[]any{1, 2, 3}, false},
	}

	for _, tt := range tests {
		result := allStrings(tt.values)
		if result != tt.expected {
			t.Errorf("allStrings(%v) = %v, want %v", tt.values, result, tt.expected)
		}
	}
}

// Parser tests for new features
func TestParser_RegexModifier(t *testing.T) {
	input := `export const slugSchema = z.string().regex(/^[a-z]+$/);`
	parser := NewParser(input)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Schemas) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(file.Schemas))
	}

	schema, ok := file.Schemas[0].(*SchemaDecl)
	if !ok {
		t.Fatalf("expected SchemaDecl, got %T", file.Schemas[0])
	}

	withMods, ok := schema.Schema.(ZWithModifiers)
	if !ok {
		t.Fatalf("expected ZWithModifiers, got %T", schema.Schema)
	}

	// Find the regex modifier
	found := false
	for _, mod := range withMods.Modifiers {
		if regex, ok := mod.(Regex); ok {
			found = true
			if regex.Pattern != "^[a-z]+$" {
				t.Errorf("expected pattern '^[a-z]+$', got %q", regex.Pattern)
			}
		}
	}

	if !found {
		t.Error("expected to find Regex modifier")
	}
}

func TestParser_RegexWithNestedParens(t *testing.T) {
	// Test regex with parentheses inside like (?:...)
	input := `export const slugSchema = z.string().regex(/^[a-z0-9]+(?:-[a-z0-9]+)*$/);`
	parser := NewParser(input)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Schemas) != 1 {
		t.Fatalf("expected 1 schema, got %d", len(file.Schemas))
	}

	schema, ok := file.Schemas[0].(*SchemaDecl)
	if !ok {
		t.Fatalf("expected SchemaDecl, got %T", file.Schemas[0])
	}

	withMods, ok := schema.Schema.(ZWithModifiers)
	if !ok {
		t.Fatalf("expected ZWithModifiers, got %T", schema.Schema)
	}

	// Find the regex modifier
	found := false
	for _, mod := range withMods.Modifiers {
		if regex, ok := mod.(Regex); ok {
			found = true
			expected := "^[a-z0-9]+(?:-[a-z0-9]+)*$"
			if regex.Pattern != expected {
				t.Errorf("expected pattern %q, got %q", expected, regex.Pattern)
			}
		}
	}

	if !found {
		t.Error("expected to find Regex modifier")
	}
}
