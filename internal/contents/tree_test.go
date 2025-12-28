package contents

import (
	"os"
	"testing"

	"github.com/hntrl/sparktype/internal/config"
	"github.com/hntrl/sparktype/internal/spec"
)

// setupTestRegistry creates a registry with test specs
func setupTestRegistry() *spec.Registry {
	cwd, _ := os.Getwd()
	// Navigate up to the root of internal package
	basePath := cwd + "/../spec"

	specs := map[string]config.Spec{
		"users": {
			Schemas: map[string]any{
				"User": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":   map[string]any{"type": "integer"},
						"name": map[string]any{"type": "string"},
					},
				},
				"UserRequest": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"name": map[string]any{"type": "string"},
					},
				},
				"UserResponse": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"user": map[string]any{"$ref": "#/components/schemas/User"},
					},
				},
			},
		},
		"products": {
			Schemas: map[string]any{
				"Product": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id":    map[string]any{"type": "string"},
						"price": map[string]any{"type": "number"},
					},
				},
				"ProductCategory": map[string]any{
					"type": "string",
					"enum": []any{"electronics", "clothing", "food"},
				},
			},
		},
	}

	return spec.NewRegistry(specs, basePath)
}

func TestNewResolver(t *testing.T) {
	registry := setupTestRegistry()
	resolver := NewResolver(registry)

	if resolver == nil {
		t.Fatal("expected non-nil resolver")
	}
}

func TestResolver_Resolve_SimplePattern(t *testing.T) {
	registry := setupTestRegistry()
	resolver := NewResolver(registry)

	contents := []config.ContentItem{
		{Pattern: "users:User"},
	}

	tree, err := resolver.Resolve(contents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tree.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(tree.Nodes))
	}

	if !tree.Nodes[0].IsSchema() {
		t.Error("expected schema node")
	}

	if tree.Nodes[0].Schema.Name != "User" {
		t.Errorf("expected schema name 'User', got %q", tree.Nodes[0].Schema.Name)
	}
}

func TestResolver_Resolve_WildcardPattern(t *testing.T) {
	registry := setupTestRegistry()
	resolver := NewResolver(registry)

	contents := []config.ContentItem{
		{Pattern: "users:*"},
	}

	tree, err := resolver.Resolve(contents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tree.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(tree.Nodes))
	}

	// All should be schemas
	for _, node := range tree.Nodes {
		if !node.IsSchema() {
			t.Error("expected schema node")
		}
	}
}

func TestResolver_Resolve_GlobSuffixPattern(t *testing.T) {
	registry := setupTestRegistry()
	resolver := NewResolver(registry)

	contents := []config.ContentItem{
		{Pattern: "users:*Request"},
	}

	tree, err := resolver.Resolve(contents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tree.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(tree.Nodes))
	}

	if tree.Nodes[0].Schema.Name != "UserRequest" {
		t.Errorf("expected UserRequest, got %q", tree.Nodes[0].Schema.Name)
	}
}

func TestResolver_Resolve_GlobPrefixPattern(t *testing.T) {
	registry := setupTestRegistry()
	resolver := NewResolver(registry)

	contents := []config.ContentItem{
		{Pattern: "users:User*"},
	}

	tree, err := resolver.Resolve(contents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should match User, UserRequest, UserResponse
	if len(tree.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(tree.Nodes))
	}
}

func TestResolver_Resolve_Namespace(t *testing.T) {
	registry := setupTestRegistry()
	resolver := NewResolver(registry)

	contents := []config.ContentItem{
		{
			Namespace: &config.NamespaceDef{
				Name: "Models",
				Contents: []config.ContentItem{
					{Pattern: "users:User"},
				},
			},
		},
	}

	tree, err := resolver.Resolve(contents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tree.Nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(tree.Nodes))
	}

	if !tree.Nodes[0].IsNamespace() {
		t.Fatal("expected namespace node")
	}

	ns := tree.Nodes[0].Namespace
	if ns.Name != "Models" {
		t.Errorf("expected namespace name 'Models', got %q", ns.Name)
	}

	if len(ns.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(ns.Children))
	}

	if !ns.Children[0].IsSchema() {
		t.Error("expected child to be schema")
	}

	if ns.Children[0].Schema.Name != "User" {
		t.Errorf("expected User schema, got %q", ns.Children[0].Schema.Name)
	}

	// Check namespace tracking
	path, ok := tree.SchemaNamespaces["User"]
	if !ok {
		t.Fatal("expected User in SchemaNamespaces")
	}
	if len(path) != 1 || path[0] != "Models" {
		t.Errorf("expected path ['Models'], got %v", path)
	}
}

func TestResolver_Resolve_NestedNamespaces(t *testing.T) {
	registry := setupTestRegistry()
	resolver := NewResolver(registry)

	contents := []config.ContentItem{
		{
			Namespace: &config.NamespaceDef{
				Name: "API",
				Contents: []config.ContentItem{
					{Pattern: "users:User"},
					{
						Namespace: &config.NamespaceDef{
							Name: "Products",
							Contents: []config.ContentItem{
								{Pattern: "products:*"},
							},
						},
					},
				},
			},
		},
	}

	tree, err := resolver.Resolve(contents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check User namespace path
	userPath := tree.SchemaNamespaces["User"]
	if len(userPath) != 1 || userPath[0] != "API" {
		t.Errorf("expected User path ['API'], got %v", userPath)
	}

	// Check Product namespace path
	productPath := tree.SchemaNamespaces["Product"]
	if len(productPath) != 2 || productPath[0] != "API" || productPath[1] != "Products" {
		t.Errorf("expected Product path ['API', 'Products'], got %v", productPath)
	}
}

func TestResolver_Resolve_MixedContent(t *testing.T) {
	registry := setupTestRegistry()
	resolver := NewResolver(registry)

	contents := []config.ContentItem{
		{Pattern: "users:User"},
		{
			Namespace: &config.NamespaceDef{
				Name: "Products",
				Contents: []config.ContentItem{
					{Pattern: "products:*"},
				},
			},
		},
		{Pattern: "users:*Request"},
	}

	tree, err := resolver.Resolve(contents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have: User (schema), Products (namespace), UserRequest (schema)
	if len(tree.Nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(tree.Nodes))
	}

	// First should be User schema
	if !tree.Nodes[0].IsSchema() || tree.Nodes[0].Schema.Name != "User" {
		t.Error("expected first node to be User schema")
	}

	// Second should be Products namespace
	if !tree.Nodes[1].IsNamespace() || tree.Nodes[1].Namespace.Name != "Products" {
		t.Error("expected second node to be Products namespace")
	}

	// Third should be UserRequest schema
	if !tree.Nodes[2].IsSchema() || tree.Nodes[2].Schema.Name != "UserRequest" {
		t.Error("expected third node to be UserRequest schema")
	}

	// Check namespace paths
	if path := tree.SchemaNamespaces["User"]; len(path) != 0 {
		t.Errorf("expected User at root, got path %v", path)
	}

	if path := tree.SchemaNamespaces["Product"]; len(path) != 1 || path[0] != "Products" {
		t.Errorf("expected Product in Products, got path %v", path)
	}
}

func TestResolver_Resolve_InvalidPattern(t *testing.T) {
	registry := setupTestRegistry()
	resolver := NewResolver(registry)

	tests := []struct {
		name    string
		pattern string
	}{
		{"missing colon", "users"},
		{"unknown spec", "unknown:*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents := []config.ContentItem{{Pattern: tt.pattern}}
			_, err := resolver.Resolve(contents)
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestResolvedTree_GetNamespacePath(t *testing.T) {
	tree := &ResolvedTree{
		SchemaNamespaces: map[string][]string{
			"User":    {},
			"Product": {"Models"},
			"Order":   {"Models", "Commerce"},
		},
	}

	tests := []struct {
		schema   string
		expected string
	}{
		{"User", "User"},
		{"Product", "Models.Product"},
		{"Order", "Models.Commerce.Order"},
		{"Unknown", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.schema, func(t *testing.T) {
			result := tree.GetNamespacePath(tt.schema)
			if result != tt.expected {
				t.Errorf("GetNamespacePath(%q) = %q, want %q", tt.schema, result, tt.expected)
			}
		})
	}
}

func TestResolvedTree_GetRelativeRef(t *testing.T) {
	tree := &ResolvedTree{
		SchemaNamespaces: map[string][]string{
			"User":    {},
			"Product": {"Models"},
			"Order":   {"Models", "Commerce"},
			"Item":    {"Models", "Commerce"},
		},
	}

	tests := []struct {
		name     string
		fromPath []string
		toSchema string
		expected string
	}{
		{"same namespace", []string{"Models", "Commerce"}, "Item", "Item"},
		{"root to root", []string{}, "User", "User"},
		{"root to nested", []string{}, "Product", "Models.Product"},
		{"nested to root", []string{"Models"}, "User", "User"},
		{"nested to different nested", []string{"Models"}, "Order", "Models.Commerce.Order"},
		{"unknown schema", []string{}, "Unknown", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tree.GetRelativeRef(tt.fromPath, tt.toSchema)
			if result != tt.expected {
				t.Errorf("GetRelativeRef(%v, %q) = %q, want %q",
					tt.fromPath, tt.toSchema, result, tt.expected)
			}
		})
	}
}

func TestResolvedTree_CollectAllSchemas(t *testing.T) {
	registry := setupTestRegistry()
	resolver := NewResolver(registry)

	contents := []config.ContentItem{
		{Pattern: "users:User"},
		{
			Namespace: &config.NamespaceDef{
				Name: "Products",
				Contents: []config.ContentItem{
					{Pattern: "products:*"},
				},
			},
		},
	}

	tree, err := resolver.Resolve(contents)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	schemas := tree.CollectAllSchemas()

	// Should have User + Product + ProductCategory = 3
	if len(schemas) != 3 {
		t.Errorf("expected 3 schemas, got %d", len(schemas))
	}

	// Check names
	names := make(map[string]bool)
	for _, s := range schemas {
		names[s.Name] = true
	}

	expected := []string{"User", "Product", "ProductCategory"}
	for _, name := range expected {
		if !names[name] {
			t.Errorf("expected to find schema %q", name)
		}
	}
}

func TestNode_IsSchema(t *testing.T) {
	schemaNode := Node{Schema: &spec.Schema{Name: "Test"}}
	if !schemaNode.IsSchema() {
		t.Error("expected IsSchema() to be true")
	}
	if schemaNode.IsNamespace() {
		t.Error("expected IsNamespace() to be false")
	}
}

func TestNode_IsNamespace(t *testing.T) {
	nsNode := Node{Namespace: &NamespaceNode{Name: "Test"}}
	if nsNode.IsSchema() {
		t.Error("expected IsSchema() to be false")
	}
	if !nsNode.IsNamespace() {
		t.Error("expected IsNamespace() to be true")
	}
}

func TestPathEqual(t *testing.T) {
	tests := []struct {
		a, b     []string
		expected bool
	}{
		{[]string{}, []string{}, true},
		{[]string{"a"}, []string{"a"}, true},
		{[]string{"a", "b"}, []string{"a", "b"}, true},
		{[]string{"a"}, []string{"b"}, false},
		{[]string{"a"}, []string{"a", "b"}, false},
		{[]string{"a", "b"}, []string{"a"}, false},
		{nil, nil, true},
		{nil, []string{}, true},
	}

	for _, tt := range tests {
		result := pathEqual(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("pathEqual(%v, %v) = %v, want %v", tt.a, tt.b, result, tt.expected)
		}
	}
}

