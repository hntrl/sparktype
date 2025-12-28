package spec

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/hntrl/sparktype/internal/config"
)

func TestLoadSpec_Local_YAML(t *testing.T) {
	cwd, _ := os.Getwd()
	specConfig := config.Spec{Path: "./testdata/simple.yaml"}

	schemas, err := LoadSpec(specConfig, cwd, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schemas) != 2 {
		t.Errorf("expected 2 schemas, got %d", len(schemas))
	}

	// Check for User schema
	var userSchema *Schema
	for i := range schemas {
		if schemas[i].Name == "User" {
			userSchema = &schemas[i]
			break
		}
	}

	if userSchema == nil {
		t.Fatal("expected to find User schema")
	}

	if userSchema.Type != "object" {
		t.Errorf("expected User type 'object', got %q", userSchema.Type)
	}

	if userSchema.Description != "A user account" {
		t.Errorf("expected User description 'A user account', got %q", userSchema.Description)
	}

	if len(userSchema.Properties) != 3 {
		t.Errorf("expected 3 properties, got %d", len(userSchema.Properties))
	}

	// Check required fields
	if len(userSchema.Required) != 2 {
		t.Errorf("expected 2 required fields, got %d", len(userSchema.Required))
	}
}

func TestLoadSpec_Local_JSON(t *testing.T) {
	cwd, _ := os.Getwd()
	specConfig := config.Spec{Path: "./testdata/simple.json"}

	schemas, err := LoadSpec(specConfig, cwd, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schemas) != 1 {
		t.Errorf("expected 1 schema, got %d", len(schemas))
	}

	if schemas[0].Name != "Product" {
		t.Errorf("expected schema name 'Product', got %q", schemas[0].Name)
	}
}

func TestLoadSpec_Local_RelativePath(t *testing.T) {
	// Test that relative paths are resolved from basePath
	testPath := filepath.Join("testdata", "simple.yaml")
	cwd, _ := os.Getwd()
	specConfig := config.Spec{Path: testPath}

	schemas, err := LoadSpec(specConfig, cwd, "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schemas) == 0 {
		t.Error("expected at least one schema")
	}
}

func TestLoadSpec_Local_NotFound(t *testing.T) {
	cwd, _ := os.Getwd()
	specConfig := config.Spec{Path: "./testdata/nonexistent.yaml"}

	_, err := LoadSpec(specConfig, cwd, "test")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoadSpec_Inline(t *testing.T) {
	specConfig := config.Spec{
		Schemas: map[string]any{
			"InlineUser": map[string]any{
				"type":        "object",
				"description": "An inline user",
				"properties": map[string]any{
					"id": map[string]any{
						"type": "integer",
					},
					"name": map[string]any{
						"type": "string",
					},
				},
				"required": []any{"id"},
			},
			"InlineStatus": map[string]any{
				"type": "string",
				"enum": []any{"on", "off"},
			},
		},
	}

	schemas, err := LoadSpec(specConfig, "", "inline")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schemas) != 2 {
		t.Errorf("expected 2 schemas, got %d", len(schemas))
	}

	// Find InlineUser
	var userSchema *Schema
	for i := range schemas {
		if schemas[i].Name == "InlineUser" {
			userSchema = &schemas[i]
			break
		}
	}

	if userSchema == nil {
		t.Fatal("expected to find InlineUser schema")
	}

	if userSchema.Type != "object" {
		t.Errorf("expected type 'object', got %q", userSchema.Type)
	}

	if userSchema.SourceSpec != "inline" {
		t.Errorf("expected SourceSpec 'inline', got %q", userSchema.SourceSpec)
	}
}

func TestLoadSpec_Remote(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check custom header
		if r.Header.Get("X-Custom-Header") != "test-value" {
			t.Errorf("expected custom header 'test-value', got %q", r.Header.Get("X-Custom-Header"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"openapi": "3.0.0",
			"info": {"title": "Remote API", "version": "1.0.0"},
			"paths": {},
			"components": {
				"schemas": {
					"RemoteType": {
						"type": "object",
						"properties": {
							"value": {"type": "string"}
						}
					}
				}
			}
		}`))
	}))
	defer server.Close()

	specConfig := config.Spec{
		URL: server.URL,
		Headers: map[string]string{
			"X-Custom-Header": "test-value",
		},
	}

	schemas, err := LoadSpec(specConfig, "", "remote")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schemas) != 1 {
		t.Errorf("expected 1 schema, got %d", len(schemas))
	}

	if schemas[0].Name != "RemoteType" {
		t.Errorf("expected schema name 'RemoteType', got %q", schemas[0].Name)
	}
}

func TestLoadSpec_Remote_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	specConfig := config.Spec{URL: server.URL}

	_, err := LoadSpec(specConfig, "", "remote")
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestLoadSpec_NoSource(t *testing.T) {
	specConfig := config.Spec{}

	_, err := LoadSpec(specConfig, "", "empty")
	if err == nil {
		t.Error("expected error for spec with no source")
	}
}

func TestLoadSpec_EmptySpec(t *testing.T) {
	cwd, _ := os.Getwd()
	specConfig := config.Spec{Path: "./testdata/empty.yaml"}

	schemas, err := LoadSpec(specConfig, cwd, "empty")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(schemas) != 0 {
		t.Errorf("expected 0 schemas, got %d", len(schemas))
	}
}

func TestLoadSpec_Complex(t *testing.T) {
	cwd, _ := os.Getwd()
	specConfig := config.Spec{Path: "./testdata/complex.yaml"}

	schemas, err := LoadSpec(specConfig, cwd, "complex")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have multiple schemas
	schemaMap := make(map[string]*Schema)
	for i := range schemas {
		schemaMap[schemas[i].Name] = &schemas[i]
	}

	// Test Address schema
	address, ok := schemaMap["Address"]
	if !ok {
		t.Fatal("expected Address schema")
	}
	if len(address.Properties) != 3 {
		t.Errorf("expected 3 properties in Address, got %d", len(address.Properties))
	}

	// Test Person schema with $ref
	person, ok := schemaMap["Person"]
	if !ok {
		t.Fatal("expected Person schema")
	}
	// Check address property is a reference
	var addressProp *Property
	for i := range person.Properties {
		if person.Properties[i].Name == "address" {
			addressProp = &person.Properties[i]
			break
		}
	}
	if addressProp == nil {
		t.Fatal("expected address property in Person")
	}
	if addressProp.Schema.Ref == "" {
		t.Error("expected address property to be a reference")
	}

	// Test ExtendedUser with allOf
	extUser, ok := schemaMap["ExtendedUser"]
	if !ok {
		t.Fatal("expected ExtendedUser schema")
	}
	if len(extUser.AllOf) != 2 {
		t.Errorf("expected 2 allOf schemas, got %d", len(extUser.AllOf))
	}

	// Test Response with oneOf
	response, ok := schemaMap["Response"]
	if !ok {
		t.Fatal("expected Response schema")
	}
	if len(response.OneOf) != 2 {
		t.Errorf("expected 2 oneOf schemas, got %d", len(response.OneOf))
	}

	// Test NullableField
	nullable, ok := schemaMap["NullableField"]
	if !ok {
		t.Fatal("expected NullableField schema")
	}
	var valueProp *Property
	for i := range nullable.Properties {
		if nullable.Properties[i].Name == "value" {
			valueProp = &nullable.Properties[i]
			break
		}
	}
	if valueProp == nil {
		t.Fatal("expected value property in NullableField")
	}
	if !valueProp.Schema.Nullable {
		t.Error("expected value property to be nullable")
	}

	// Test ArrayWithConstraints
	arrConst, ok := schemaMap["ArrayWithConstraints"]
	if !ok {
		t.Fatal("expected ArrayWithConstraints schema")
	}
	if arrConst.MinItems == nil || *arrConst.MinItems != 1 {
		t.Errorf("expected minItems 1, got %v", arrConst.MinItems)
	}
	if arrConst.MaxItems == nil || *arrConst.MaxItems != 10 {
		t.Errorf("expected maxItems 10, got %v", arrConst.MaxItems)
	}
	if !arrConst.UniqueItems {
		t.Error("expected uniqueItems true")
	}

	// Test StringWithConstraints
	strConst, ok := schemaMap["StringWithConstraints"]
	if !ok {
		t.Fatal("expected StringWithConstraints schema")
	}
	if strConst.MinLength == nil || *strConst.MinLength != 3 {
		t.Errorf("expected minLength 3, got %v", strConst.MinLength)
	}
	if strConst.MaxLength == nil || *strConst.MaxLength != 50 {
		t.Errorf("expected maxLength 50, got %v", strConst.MaxLength)
	}
	if strConst.Pattern != "^[a-z]+$" {
		t.Errorf("expected pattern '^[a-z]+$', got %q", strConst.Pattern)
	}

	// Test NumberWithConstraints
	numConst, ok := schemaMap["NumberWithConstraints"]
	if !ok {
		t.Fatal("expected NumberWithConstraints schema")
	}
	if numConst.Minimum == nil || *numConst.Minimum != 0 {
		t.Errorf("expected minimum 0, got %v", numConst.Minimum)
	}
	if numConst.Maximum == nil || *numConst.Maximum != 100 {
		t.Errorf("expected maximum 100, got %v", numConst.Maximum)
	}
}
