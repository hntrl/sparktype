package zod

import (
	"sort"

	"github.com/hntrl/sparktype/internal/contents"
	"github.com/hntrl/sparktype/internal/generators"
)

// Compare compares existing Zod file with expected output
func Compare(existing []byte, tree *contents.ResolvedTree, opts generators.Options) (*generators.CompareResult, error) {
	// Parse existing file
	parser := NewParser(string(existing))
	existingAST, err := parser.ParseFile()
	if err != nil {
		return nil, err
	}

	// Build expected AST
	expectedAST := BuildFile(tree, opts)

	// Compare ASTs
	return CompareFiles(expectedAST, existingAST), nil
}

// CompareFiles compares two File ASTs and returns a CompareResult
func CompareFiles(expected, existing *File) *generators.CompareResult {
	result := &generators.CompareResult{Match: true}

	// Extract declarations from both
	expectedDecls := ExtractSchemaNodeMap(expected.Schemas, "")
	existingDecls := ExtractSchemaNodeMap(existing.Schemas, "")

	// Get all unique names
	allNames := make(map[string]bool)
	for name := range expectedDecls {
		allNames[name] = true
	}
	for name := range existingDecls {
		allNames[name] = true
	}

	// Sort for consistent output
	names := make([]string, 0, len(allNames))
	for name := range allNames {
		names = append(names, name)
	}
	sort.Strings(names)

	// Compare each schema
	for _, name := range names {
		expectedNode, hasExpected := expectedDecls[name]
		existingNode, hasExisting := existingDecls[name]

		if hasExpected && !hasExisting {
			result.Match = false
			result.Types = append(result.Types, generators.TypeDiff{
				Name:   name,
				Status: generators.Added,
			})
		} else if !hasExpected && hasExisting {
			result.Match = false
			result.Types = append(result.Types, generators.TypeDiff{
				Name:   name,
				Status: generators.Removed,
			})
		} else {
			// Both exist - compare them
			propDiffs := CompareSchemaNodes(expectedNode, existingNode)
			if len(propDiffs) > 0 {
				result.Match = false
				result.Types = append(result.Types, generators.TypeDiff{
					Name:       name,
					Status:     generators.Changed,
					Properties: propDiffs,
				})
			}
		}
	}

	return result
}

func ExtractSchemaNodeMap(nodes []Node, prefix string) map[string]Node {
	result := make(map[string]Node)

	for _, node := range nodes {
		switch n := node.(type) {
		case *SchemaDecl:
			result[prefix+n.Name] = n
		case *Namespace:
			// Recurse into namespace
			for _, child := range n.Children {
				switch c := child.(type) {
				case *SchemaProperty:
					result[prefix+n.Name+"."+c.Name] = c
				case *NestedNamespace:
					ExtractNestedNodeMap(c, prefix+n.Name+".", result)
				}
			}
		}
	}

	return result
}

func ExtractNestedNodeMap(ns *NestedNamespace, prefix string, result map[string]Node) {
	for _, child := range ns.Children {
		switch c := child.(type) {
		case *SchemaProperty:
			result[prefix+ns.Name+"."+c.Name] = c
		case *NestedNamespace:
			ExtractNestedNodeMap(c, prefix+ns.Name+".", result)
		}
	}
}

func CompareSchemaNodes(expected, existing Node) []generators.PropertyDiff {
	// Extract schema expressions from both
	expectedExpr := GetSchemaExpr(expected)
	existingExpr := GetSchemaExpr(existing)

	if expectedExpr == nil || existingExpr == nil {
		// Type mismatch
		return []generators.PropertyDiff{{
			Name:     "(schema)",
			Status:   generators.Changed,
			OldValue: SchemaNodeTypeName(existing),
			NewValue: SchemaNodeTypeName(expected),
		}}
	}

	// Compare the Zod expressions
	return CompareZodExprs("", expectedExpr, existingExpr)
}

func GetSchemaExpr(n Node) ZodExpr {
	switch node := n.(type) {
	case *SchemaDecl:
		return node.Schema
	case *SchemaProperty:
		return node.Schema
	default:
		return nil
	}
}

func CompareZodExprs(propName string, expected, existing ZodExpr) []generators.PropertyDiff {
	expectedSer := expected.Serialize()
	existingSer := existing.Serialize()

	if expectedSer != existingSer {
		// For objects, we can do property-level comparison
		if expectedObj, ok := expected.(ZObject); ok {
			if existingObj, ok := existing.(ZObject); ok {
				return CompareZObjects(expectedObj, existingObj)
			}
		}

		// Otherwise report the whole expression as changed
		name := "(schema)"
		if propName != "" {
			name = propName
		}
		return []generators.PropertyDiff{{
			Name:     name,
			Status:   generators.Changed,
			OldValue: existingSer,
			NewValue: expectedSer,
		}}
	}

	return nil
}

func SchemaNodeTypeName(n Node) string {
	switch n.(type) {
	case *SchemaDecl:
		return "schema"
	case *SchemaProperty:
		return "property"
	case *Namespace:
		return "namespace"
	case *NestedNamespace:
		return "nested namespace"
	default:
		return "unknown"
	}
}

// CompareZObjects compares two ZObject expressions
func CompareZObjects(expected, existing ZObject) []generators.PropertyDiff {
	var diffs []generators.PropertyDiff

	// Build property maps
	expectedProps := make(map[string]ZodExpr)
	for _, p := range expected.Properties {
		expectedProps[p.Name] = p.Schema
	}

	existingProps := make(map[string]ZodExpr)
	for _, p := range existing.Properties {
		existingProps[p.Name] = p.Schema
	}

	// Get all property names
	allNames := make(map[string]bool)
	for name := range expectedProps {
		allNames[name] = true
	}
	for name := range existingProps {
		allNames[name] = true
	}

	names := make([]string, 0, len(allNames))
	for name := range allNames {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		expectedExpr, hasExpected := expectedProps[name]
		existingExpr, hasExisting := existingProps[name]

		if hasExpected && !hasExisting {
			diffs = append(diffs, generators.PropertyDiff{
				Name:     name,
				Status:   generators.Added,
				NewValue: expectedExpr.Serialize(),
			})
		} else if !hasExpected && hasExisting {
			diffs = append(diffs, generators.PropertyDiff{
				Name:     name,
				Status:   generators.Removed,
				OldValue: existingExpr.Serialize(),
			})
		} else {
			// Compare the expressions
			propDiffs := CompareZodExprs(name, expectedExpr, existingExpr)
			diffs = append(diffs, propDiffs...)
		}
	}

	return diffs
}
