package golang

import (
	"sort"

	"github.com/hntrl/sparktype/internal/contents"
	"github.com/hntrl/sparktype/internal/generators"
)

// Compare compares existing Go file with expected output
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
	expectedDecls := ExtractTypeNodeMap(expected.Types)
	existingDecls := ExtractTypeNodeMap(existing.Types)

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

	// Compare each type
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
			propDiffs := CompareTypeNodes(expectedNode, existingNode)
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

func ExtractTypeNodeMap(nodes []Node) map[string]Node {
	result := make(map[string]Node)

	for _, node := range nodes {
		switch n := node.(type) {
		case *Struct:
			result[n.Name] = n
		case *TypeAlias:
			result[n.Name] = n
		case *TypeDef:
			result[n.Name] = n
		case *EnumDef:
			result[n.Name] = n
		}
	}

	return result
}

func CompareTypeNodes(expected, existing Node) []generators.PropertyDiff {
	switch e := expected.(type) {
	case *Struct:
		if ex, ok := existing.(*Struct); ok {
			return CompareStructs(e, ex)
		}
	case *TypeAlias:
		if ex, ok := existing.(*TypeAlias); ok {
			return CompareTypeAliases(e, ex)
		}
	case *TypeDef:
		if ex, ok := existing.(*TypeDef); ok {
			return CompareTypeDefs(e, ex)
		}
	case *EnumDef:
		if ex, ok := existing.(*EnumDef); ok {
			return CompareEnumDefs(e, ex)
		}
	}

	// Type mismatch
	return []generators.PropertyDiff{{
		Name:     "(type)",
		Status:   generators.Changed,
		OldValue: TypeNodeTypeName(existing),
		NewValue: TypeNodeTypeName(expected),
	}}
}

func TypeNodeTypeName(n Node) string {
	switch n.(type) {
	case *Struct:
		return "struct"
	case *TypeAlias:
		return "type alias"
	case *TypeDef:
		return "type"
	case *EnumDef:
		return "enum"
	default:
		return "unknown"
	}
}

// CompareStructs compares two Struct nodes
func CompareStructs(expected, existing *Struct) []generators.PropertyDiff {
	var diffs []generators.PropertyDiff

	// Build field maps
	expectedFields := make(map[string]StructField)
	for _, f := range expected.Fields {
		expectedFields[f.Name] = f
	}

	existingFields := make(map[string]StructField)
	for _, f := range existing.Fields {
		existingFields[f.Name] = f
	}

	// Get all field names
	allNames := make(map[string]bool)
	for name := range expectedFields {
		allNames[name] = true
	}
	for name := range existingFields {
		allNames[name] = true
	}

	names := make([]string, 0, len(allNames))
	for name := range allNames {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		expectedField, hasExpected := expectedFields[name]
		existingField, hasExisting := existingFields[name]

		if hasExpected && !hasExisting {
			diffs = append(diffs, generators.PropertyDiff{
				Name:     name,
				Status:   generators.Added,
				NewValue: expectedField.Type.Serialize(),
			})
		} else if !hasExpected && hasExisting {
			diffs = append(diffs, generators.PropertyDiff{
				Name:     name,
				Status:   generators.Removed,
				OldValue: existingField.Type.Serialize(),
			})
		} else {
			// Compare field types and tags
			if diff := CompareStructFields(expectedField, existingField); diff != nil {
				diffs = append(diffs, *diff)
			}
		}
	}

	return diffs
}

// CompareStructFields compares two StructField values
func CompareStructFields(expected, existing StructField) *generators.PropertyDiff {
	expectedType := expected.Type.Serialize()
	existingType := existing.Type.Serialize()

	if expectedType != existingType {
		return &generators.PropertyDiff{
			Name:     expected.Name,
			Status:   generators.Changed,
			OldValue: existingType,
			NewValue: expectedType,
		}
	}

	// Compare JSON tags
	if expected.JSONTag != existing.JSONTag {
		return &generators.PropertyDiff{
			Name:     expected.Name + " (json tag)",
			Status:   generators.Changed,
			OldValue: existing.JSONTag,
			NewValue: expected.JSONTag,
		}
	}

	return nil
}

// CompareTypeAliases compares two TypeAlias nodes
func CompareTypeAliases(expected, existing *TypeAlias) []generators.PropertyDiff {
	expectedType := expected.Type.Serialize()
	existingType := existing.Type.Serialize()

	if expectedType != existingType {
		return []generators.PropertyDiff{{
			Name:     "(type)",
			Status:   generators.Changed,
			OldValue: existingType,
			NewValue: expectedType,
		}}
	}

	return nil
}

// CompareTypeDefs compares two TypeDef nodes
func CompareTypeDefs(expected, existing *TypeDef) []generators.PropertyDiff {
	expectedType := expected.Type.Serialize()
	existingType := existing.Type.Serialize()

	if expectedType != existingType {
		return []generators.PropertyDiff{{
			Name:     "(type)",
			Status:   generators.Changed,
			OldValue: existingType,
			NewValue: expectedType,
		}}
	}

	return nil
}

// CompareEnumDefs compares two EnumDef nodes
func CompareEnumDefs(expected, existing *EnumDef) []generators.PropertyDiff {
	var diffs []generators.PropertyDiff

	// Build value maps
	expectedValues := make(map[string]string)
	for _, v := range expected.Values {
		expectedValues[v.Name] = v.Value
	}

	existingValues := make(map[string]string)
	for _, v := range existing.Values {
		existingValues[v.Name] = v.Value
	}

	// Get all value names
	allNames := make(map[string]bool)
	for name := range expectedValues {
		allNames[name] = true
	}
	for name := range existingValues {
		allNames[name] = true
	}

	names := make([]string, 0, len(allNames))
	for name := range allNames {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		expectedVal, hasExpected := expectedValues[name]
		existingVal, hasExisting := existingValues[name]

		if hasExpected && !hasExisting {
			diffs = append(diffs, generators.PropertyDiff{
				Name:     name,
				Status:   generators.Added,
				NewValue: expectedVal,
			})
		} else if !hasExpected && hasExisting {
			diffs = append(diffs, generators.PropertyDiff{
				Name:     name,
				Status:   generators.Removed,
				OldValue: existingVal,
			})
		} else if expectedVal != existingVal {
			diffs = append(diffs, generators.PropertyDiff{
				Name:     name,
				Status:   generators.Changed,
				OldValue: existingVal,
				NewValue: expectedVal,
			})
		}
	}

	return diffs
}
