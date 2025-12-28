package spec

import (
	"os"
	"sort"
	"sync"
	"testing"

	"github.com/hntrl/sparktype/internal/config"
)

func TestNewRegistry(t *testing.T) {
	specs := map[string]config.Spec{
		"api": {Path: "./testdata/simple.yaml"},
	}

	registry := NewRegistry(specs, ".")
	if registry == nil {
		t.Fatal("expected non-nil registry")
	}

	if len(registry.specs) != 1 {
		t.Errorf("expected 1 spec, got %d", len(registry.specs))
	}

	if len(registry.loaded) != 0 {
		t.Errorf("expected 0 loaded specs, got %d", len(registry.loaded))
	}
}

func TestRegistry_GetSchemas(t *testing.T) {
	cwd, _ := os.Getwd()
	specs := map[string]config.Spec{
		"simple": {Path: "./testdata/simple.yaml"},
	}

	registry := NewRegistry(specs, cwd)

	schemas, err := registry.GetSchemas("simple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schemas) != 2 {
		t.Errorf("expected 2 schemas, got %d", len(schemas))
	}

	// Verify User schema exists
	found := false
	for _, s := range schemas {
		if s.Name == "User" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find User schema")
	}
}

func TestRegistry_GetSchemas_Caching(t *testing.T) {
	cwd, _ := os.Getwd()
	specs := map[string]config.Spec{
		"simple": {Path: "./testdata/simple.yaml"},
	}

	registry := NewRegistry(specs, cwd)

	// First call loads the spec
	schemas1, err := registry.GetSchemas("simple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call should return cached version
	schemas2, err := registry.GetSchemas("simple")
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}

	// Should have same number of schemas
	if len(schemas1) != len(schemas2) {
		t.Errorf("expected same length, got %d and %d", len(schemas1), len(schemas2))
	}

	// Verify the spec is cached
	if len(registry.loaded) != 1 {
		t.Errorf("expected 1 loaded spec, got %d", len(registry.loaded))
	}
}

func TestRegistry_GetSchemas_DeepCopy(t *testing.T) {
	cwd, _ := os.Getwd()
	specs := map[string]config.Spec{
		"simple": {Path: "./testdata/simple.yaml"},
	}

	registry := NewRegistry(specs, cwd)

	// Get schemas
	schemas1, err := registry.GetSchemas("simple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Modify the returned schemas
	if len(schemas1) > 0 {
		schemas1[0].Name = "MODIFIED"
	}

	// Get schemas again
	schemas2, err := registry.GetSchemas("simple")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not be modified
	for _, s := range schemas2 {
		if s.Name == "MODIFIED" {
			t.Error("cached schema was modified - deep copy not working")
		}
	}
}

func TestRegistry_GetSchemas_NotFound(t *testing.T) {
	cwd, _ := os.Getwd()
	specs := map[string]config.Spec{
		"simple": {Path: "./testdata/simple.yaml"},
	}

	registry := NewRegistry(specs, cwd)

	_, err := registry.GetSchemas("nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent spec")
	}
}

func TestRegistry_GetSchemas_LoadError(t *testing.T) {
	cwd, _ := os.Getwd()
	specs := map[string]config.Spec{
		"broken": {Path: "./testdata/nonexistent.yaml"},
	}

	registry := NewRegistry(specs, cwd)

	_, err := registry.GetSchemas("broken")
	if err == nil {
		t.Error("expected error for broken spec")
	}
}

func TestRegistry_GetSchemas_Concurrent(t *testing.T) {
	cwd, _ := os.Getwd()
	specs := map[string]config.Spec{
		"simple": {Path: "./testdata/simple.yaml"},
	}

	registry := NewRegistry(specs, cwd)

	// Spawn multiple goroutines to access the same spec
	var wg sync.WaitGroup
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := registry.GetSchemas("simple")
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("concurrent access error: %v", err)
	}

	// Should only have loaded once
	if len(registry.loaded) != 1 {
		t.Errorf("expected 1 loaded spec, got %d", len(registry.loaded))
	}
}

func TestRegistry_GetAllSpecNames(t *testing.T) {
	specs := map[string]config.Spec{
		"api":      {Path: "./api.yaml"},
		"users":    {Path: "./users.yaml"},
		"products": {Path: "./products.yaml"},
	}

	registry := NewRegistry(specs, ".")

	names := registry.GetAllSpecNames()
	if len(names) != 3 {
		t.Errorf("expected 3 names, got %d", len(names))
	}

	// Sort for consistent comparison
	sort.Strings(names)
	expected := []string{"api", "products", "users"}

	for i, name := range names {
		if name != expected[i] {
			t.Errorf("expected name %q at index %d, got %q", expected[i], i, name)
		}
	}
}

func TestDeepCopySchema(t *testing.T) {
	original := Schema{
		Name:        "Test",
		Description: "A test schema",
		Type:        "object",
		Properties: []Property{
			{
				Name:     "prop1",
				Schema:   Schema{Type: "string"},
				Required: true,
			},
		},
		Required: []string{"prop1"},
		Items:    &Schema{Type: "integer"},
		Enum:     []any{"a", "b", "c"},
		AllOf: []Schema{
			{Type: "object"},
		},
		OneOf: []Schema{
			{Type: "string"},
		},
		AnyOf: []Schema{
			{Type: "number"},
		},
	}

	copied := deepCopySchema(original)

	// Modify original
	original.Name = "Modified"
	original.Properties[0].Name = "modified_prop"
	original.Required[0] = "modified"
	original.Items.Type = "modified"
	original.Enum[0] = "modified"
	original.AllOf[0].Type = "modified"
	original.OneOf[0].Type = "modified"
	original.AnyOf[0].Type = "modified"

	// Check copy is unaffected
	if copied.Name != "Test" {
		t.Errorf("Name was modified: got %q", copied.Name)
	}
	if copied.Properties[0].Name != "prop1" {
		t.Errorf("Property name was modified: got %q", copied.Properties[0].Name)
	}
	if copied.Required[0] != "prop1" {
		t.Errorf("Required was modified: got %q", copied.Required[0])
	}
	if copied.Items.Type != "integer" {
		t.Errorf("Items.Type was modified: got %q", copied.Items.Type)
	}
	if copied.Enum[0] != "a" {
		t.Errorf("Enum was modified: got %v", copied.Enum[0])
	}
	if copied.AllOf[0].Type != "object" {
		t.Errorf("AllOf was modified: got %q", copied.AllOf[0].Type)
	}
	if copied.OneOf[0].Type != "string" {
		t.Errorf("OneOf was modified: got %q", copied.OneOf[0].Type)
	}
	if copied.AnyOf[0].Type != "number" {
		t.Errorf("AnyOf was modified: got %q", copied.AnyOf[0].Type)
	}
}

func TestDeepCopySchemas(t *testing.T) {
	original := []Schema{
		{Name: "Schema1", Type: "object"},
		{Name: "Schema2", Type: "string"},
	}

	copied := deepCopySchemas(original)

	// Modify original
	original[0].Name = "Modified"

	// Check copy is unaffected
	if copied[0].Name != "Schema1" {
		t.Errorf("Schema was modified: got %q", copied[0].Name)
	}
}
