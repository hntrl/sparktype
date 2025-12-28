package typescript

import (
	"testing"

	"github.com/sebdah/goldie/v2"

	"github.com/hntrl/sparktype/internal/config"
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

func TestGenerate_EnumUnion(t *testing.T) {
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
	g.Assert(t, "enum_union", result)
}

func TestGenerate_EnumTypeScript(t *testing.T) {
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

	result, err := Generate(tree, generators.Options{
		GlobalOptions: config.Options{
			GenerateEnums: true,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	g := goldie.New(t, goldie.WithFixtureDir("testdata/golden"))
	g.Assert(t, "enum_typescript", result)
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
					{Schema: &spec.Schema{
						Name: "Product",
						Type: "object",
						Properties: []spec.Property{
							{Name: "id", Schema: spec.Schema{Type: "string"}, Required: true},
							{Name: "price", Schema: spec.Schema{Type: "number"}, Required: true},
						},
					}},
				},
			}},
		},
		SchemaNamespaces: map[string][]string{
			"User":    {"Models"},
			"Product": {"Models"},
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

// Test AST building functions
func TestBuildNodes(t *testing.T) {
	tree := &contents.ResolvedTree{
		Nodes: []contents.Node{
			{Schema: &spec.Schema{Name: "User", Type: "object"}},
		},
		SchemaNamespaces: map[string][]string{},
	}
	ctx := &BuildContext{tree: tree, exportType: "interface"}

	nodes := BuildNodes(tree.Nodes, ctx, []string{})
	if len(nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(nodes))
	}

	if _, ok := nodes[0].(*Interface); !ok {
		t.Errorf("expected Interface, got %T", nodes[0])
	}
}

func TestBuildSchemaNode_Enum(t *testing.T) {
	schema := spec.Schema{
		Name: "Status",
		Type: "string",
		Enum: []any{"a", "b"},
	}
	tree := &contents.ResolvedTree{SchemaNamespaces: map[string][]string{}}
	ctx := &BuildContext{tree: tree, generateEnums: false}

	node := BuildSchemaNode(schema, ctx, []string{})
	if _, ok := node.(*TypeAlias); !ok {
		t.Errorf("expected TypeAlias for enum without generateEnums, got %T", node)
	}

	ctx.generateEnums = true
	node = BuildSchemaNode(schema, ctx, []string{})
	if _, ok := node.(*Enum); !ok {
		t.Errorf("expected Enum with generateEnums, got %T", node)
	}
}

func TestBuildTypeExpr(t *testing.T) {
	tree := &contents.ResolvedTree{SchemaNamespaces: map[string][]string{}}
	ctx := &BuildContext{tree: tree}

	tests := []struct {
		name     string
		schema   spec.Schema
		expected string
	}{
		{"string", spec.Schema{Type: "string"}, "string"},
		{"integer", spec.Schema{Type: "integer"}, "number"},
		{"number", spec.Schema{Type: "number"}, "number"},
		{"boolean", spec.Schema{Type: "boolean"}, "boolean"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := BuildTypeExpr(tt.schema, ctx, []string{})
			serialized := expr.Serialize()
			if serialized != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, serialized)
			}
		})
	}
}

func TestToEnumKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"active", "Active"},
		{"some-value", "SomeValue"},
		{"SOME_VALUE", "SomeValue"},
		{"some.value", "SomeValue"},
		{"a-b-c", "ABC"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toEnumKey(tt.input)
			if result != tt.expected {
				t.Errorf("toEnumKey(%q) = %q, want %q", tt.input, result, tt.expected)
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
func TestParser_NullableProperty(t *testing.T) {
	input := `export interface User {
  name: string | null;
}`
	parser := NewParser(input)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(file.Nodes))
	}

	iface, ok := file.Nodes[0].(*Interface)
	if !ok {
		t.Fatalf("expected Interface, got %T", file.Nodes[0])
	}

	if len(iface.Properties) != 1 {
		t.Fatalf("expected 1 property, got %d", len(iface.Properties))
	}

	prop := iface.Properties[0]
	if !prop.Nullable {
		t.Error("expected property to be nullable")
	}

	// Type should be string, not union
	if _, ok := prop.Type.(StringType); !ok {
		t.Errorf("expected StringType, got %T", prop.Type)
	}
}

func TestParser_NamespaceQualifiedType(t *testing.T) {
	input := `export interface Member {
  user: Users.User;
}`
	parser := NewParser(input)
	file, err := parser.ParseFile()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(file.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(file.Nodes))
	}

	iface, ok := file.Nodes[0].(*Interface)
	if !ok {
		t.Fatalf("expected Interface, got %T", file.Nodes[0])
	}

	if len(iface.Properties) != 1 {
		t.Fatalf("expected 1 property, got %d", len(iface.Properties))
	}

	prop := iface.Properties[0]
	refType, ok := prop.Type.(ReferenceType)
	if !ok {
		t.Fatalf("expected ReferenceType, got %T", prop.Type)
	}

	if refType.Name != "Users.User" {
		t.Errorf("expected 'Users.User', got %q", refType.Name)
	}
}

func TestCompare_NullableProperty(t *testing.T) {
	// Test that comparing nullable properties works correctly
	expected := &Interface{
		Name: "User",
		Properties: []Property{
			{Name: "name", Type: StringType{}, Nullable: true},
		},
	}
	existing := &Interface{
		Name: "User",
		Properties: []Property{
			{Name: "name", Type: StringType{}, Nullable: true},
		},
	}

	diffs := CompareInterfaces(expected, existing)
	if len(diffs) != 0 {
		t.Errorf("expected no diffs for matching nullable properties, got %v", diffs)
	}

	// Test mismatch
	existing.Properties[0].Nullable = false
	diffs = CompareInterfaces(expected, existing)
	if len(diffs) != 1 {
		t.Errorf("expected 1 diff for nullable mismatch, got %d", len(diffs))
	}
}

