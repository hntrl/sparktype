package python

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
				Name: "UserId",
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

// Test utility functions
func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"User", "user"},
		{"UserProfile", "user_profile"},
		{"ABC", "a_b_c"},
		{"someValue", "some_value"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
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
func TestParser_PipeNoneSyntax(t *testing.T) {
	input := `class User(TypedDict):
    name: str | None
`
	parser := NewParser(input)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Classes) != 1 {
		t.Fatalf("expected 1 class, got %d", len(file.Classes))
	}

	td, ok := file.Classes[0].(*TypedDict)
	if !ok {
		t.Fatalf("expected TypedDict, got %T", file.Classes[0])
	}

	if len(td.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(td.Fields))
	}

	// Should be parsed as OptionalType, not UnionType
	opt, ok := td.Fields[0].Type.(OptionalType)
	if !ok {
		t.Fatalf("expected OptionalType, got %T", td.Fields[0].Type)
	}

	if _, ok := opt.Inner.(StrType); !ok {
		t.Errorf("expected OptionalType inner to be StrType, got %T", opt.Inner)
	}
}

func TestParser_UnionNoneSyntax(t *testing.T) {
	input := `class User(TypedDict):
    name: Union[str, None]
`
	parser := NewParser(input)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Classes) != 1 {
		t.Fatalf("expected 1 class, got %d", len(file.Classes))
	}

	td, ok := file.Classes[0].(*TypedDict)
	if !ok {
		t.Fatalf("expected TypedDict, got %T", file.Classes[0])
	}

	if len(td.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(td.Fields))
	}

	// Should be parsed as OptionalType, not UnionType
	opt, ok := td.Fields[0].Type.(OptionalType)
	if !ok {
		t.Fatalf("expected OptionalType for Union[T, None], got %T", td.Fields[0].Type)
	}

	if _, ok := opt.Inner.(StrType); !ok {
		t.Errorf("expected OptionalType inner to be StrType, got %T", opt.Inner)
	}
}
