package golang

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

func TestGenerate_TypeAlias(t *testing.T) {
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Schema: &spec.Schema{
				Name: "UserID",
				Type: "string",
			}},
		},
		SchemaNamespaces: map[string][]string{},
	}

	result, err := Generate(tree, generators.Options{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	g.Assert(t, "type_alias", result)
}

func TestGenerate_Reference(t *testing.T) {
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Schema: &spec.Schema{
				Name: "Address",
				Type: "object",
				Properties: []spec.Property{
					{Name: "street", Schema: spec.Schema{Type: "string"}, Required: true},
					{Name: "city", Schema: spec.Schema{Type: "string"}, Required: true},
				},
			}},
			{Schema: &spec.Schema{
				Name: "User",
				Type: "object",
				Properties: []spec.Property{
					{Name: "id", Schema: spec.Schema{Type: "integer"}, Required: true},
					{Name: "address", Schema: spec.Schema{Ref: "#/components/schemas/Address"}, Required: true},
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
	g.Assert(t, "reference", result)
}

func TestGenerate_DateTime(t *testing.T) {
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Schema: &spec.Schema{
				Name: "Event",
				Type: "object",
				Properties: []spec.Property{
					{Name: "id", Schema: spec.Schema{Type: "integer"}, Required: true},
					{Name: "created_at", Schema: spec.Schema{Type: "string", Format: "date-time"}, Required: true},
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
	g.Assert(t, "datetime", result)
}

// Test utility functions
func TestToExportedName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"user", "User"},
		{"user_id", "UserID"},
		{"http_url", "HTTPURL"},
		{"id", "ID"},
		{"api_key", "APIKey"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toExportedName(tt.input)
			if result != tt.expected {
				t.Errorf("toExportedName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsCommonAcronym(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"ID", true},
		{"API", true},
		{"HTTP", true},
		{"URL", true},
		{"JSON", true},
		{"User", false},
		{"Name", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isCommonAcronym(tt.input)
			if result != tt.expected {
				t.Errorf("isCommonAcronym(%q) = %v, want %v", tt.input, result, tt.expected)
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
func TestExtractJSONTag_WithOmitempty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", `json:"name"`, "name"},
		{"with omitempty", `json:"name,omitempty"`, "name,omitempty"},
		{"multiple tags", `json:"name,omitempty" xml:"name"`, "name,omitempty"},
		{"empty", `xml:"name"`, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractJSONTag(tt.input)
			if result != tt.expected {
				t.Errorf("ExtractJSONTag(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParser_StructWithOmitempty(t *testing.T) {
	input := `package types

type User struct {
	Name *string ` + "`json:\"name,omitempty\"`" + `
}`
	parser := NewParser(input)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(file.Types))
	}

	s, ok := file.Types[0].(*Struct)
	if !ok {
		t.Fatalf("expected Struct, got %T", file.Types[0])
	}

	if len(s.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(s.Fields))
	}

	if s.Fields[0].JSONTag != "name,omitempty" {
		t.Errorf("expected JSONTag 'name,omitempty', got %q", s.Fields[0].JSONTag)
	}
}

func TestParser_EnumFromConstBlock(t *testing.T) {
	input := `package types

type Status string

const (
	StatusActive Status = "active"
	StatusInactive Status = "inactive"
)`
	parser := NewParser(input)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Types) != 1 {
		t.Fatalf("expected 1 type, got %d", len(file.Types))
	}

	enum, ok := file.Types[0].(*EnumDef)
	if !ok {
		t.Fatalf("expected EnumDef, got %T", file.Types[0])
	}

	if enum.Name != "Status" {
		t.Errorf("expected name 'Status', got %q", enum.Name)
	}

	if len(enum.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(enum.Values))
	}

	if enum.Values[0].Value != "active" {
		t.Errorf("expected first value 'active', got %q", enum.Values[0].Value)
	}
}
