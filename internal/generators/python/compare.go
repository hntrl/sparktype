package python

import (
	"sort"

	"github.com/hntrl/sparktype/internal/contents"
	"github.com/hntrl/sparktype/internal/generators"
)

// Compare compares existing Python file with expected output
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
	expectedDecls := ExtractClassNodeMap(expected.Classes)
	existingDecls := ExtractClassNodeMap(existing.Classes)

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
			propDiffs := CompareClassNodes(expectedNode, existingNode)
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

func ExtractClassNodeMap(nodes []Node) map[string]Node {
	result := make(map[string]Node)

	for _, node := range nodes {
		switch n := node.(type) {
		case *TypedDict:
			result[n.Name] = n
		case *EnumClass:
			result[n.Name] = n
		case *TypeAlias:
			result[n.Name] = n
		}
	}

	return result
}

func CompareClassNodes(expected, existing Node) []generators.PropertyDiff {
	switch e := expected.(type) {
	case *TypedDict:
		if ex, ok := existing.(*TypedDict); ok {
			return CompareTypedDicts(e, ex)
		}
	case *EnumClass:
		if ex, ok := existing.(*EnumClass); ok {
			return CompareEnumClasses(e, ex)
		}
	case *TypeAlias:
		if ex, ok := existing.(*TypeAlias); ok {
			return CompareTypeAliases(e, ex)
		}
	}

	// Type mismatch
	return []generators.PropertyDiff{{
		Name:     "(type)",
		Status:   generators.Changed,
		OldValue: ClassNodeTypeName(existing),
		NewValue: ClassNodeTypeName(expected),
	}}
}

func ClassNodeTypeName(n Node) string {
	switch n.(type) {
	case *TypedDict:
		return "TypedDict"
	case *EnumClass:
		return "Enum"
	case *TypeAlias:
		return "type alias"
	default:
		return "unknown"
	}
}

// CompareTypedDicts compares two TypedDict nodes
func CompareTypedDicts(expected, existing *TypedDict) []generators.PropertyDiff {
	var diffs []generators.PropertyDiff

	// Build field maps
	expectedFields := make(map[string]Field)
	for _, f := range expected.Fields {
		expectedFields[f.Name] = f
	}

	existingFields := make(map[string]Field)
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
			// Compare field types
			if diff := CompareFields(expectedField, existingField); diff != nil {
				diffs = append(diffs, *diff)
			}
		}
	}

	return diffs
}

// CompareFields compares two Field values
func CompareFields(expected, existing Field) *generators.PropertyDiff {
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

	return nil
}

// CompareEnumClasses compares two EnumClass nodes
func CompareEnumClasses(expected, existing *EnumClass) []generators.PropertyDiff {
	var diffs []generators.PropertyDiff

	// Build member maps
	expectedMembers := make(map[string]string)
	for _, m := range expected.Members {
		expectedMembers[m.Key] = m.Value
	}

	existingMembers := make(map[string]string)
	for _, m := range existing.Members {
		existingMembers[m.Key] = m.Value
	}

	// Get all member keys
	allKeys := make(map[string]bool)
	for key := range expectedMembers {
		allKeys[key] = true
	}
	for key := range existingMembers {
		allKeys[key] = true
	}

	keys := make([]string, 0, len(allKeys))
	for key := range allKeys {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		expectedVal, hasExpected := expectedMembers[key]
		existingVal, hasExisting := existingMembers[key]

		if hasExpected && !hasExisting {
			diffs = append(diffs, generators.PropertyDiff{
				Name:     key,
				Status:   generators.Added,
				NewValue: expectedVal,
			})
		} else if !hasExpected && hasExisting {
			diffs = append(diffs, generators.PropertyDiff{
				Name:     key,
				Status:   generators.Removed,
				OldValue: existingVal,
			})
		} else if expectedVal != existingVal {
			diffs = append(diffs, generators.PropertyDiff{
				Name:     key,
				Status:   generators.Changed,
				OldValue: existingVal,
				NewValue: expectedVal,
			})
		}
	}

	return diffs
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
