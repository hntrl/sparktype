package typescript

import (
	"sort"

	"github.com/hntrl/sparktype/internal/contents"
	"github.com/hntrl/sparktype/internal/generators"
)

// Compare compares existing TypeScript file with expected output
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
	expectedDecls := extractNodeMap(expected.Nodes, "")
	existingDecls := extractNodeMap(existing.Nodes, "")

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
			propDiffs := CompareNodes(expectedNode, existingNode)
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

func extractNodeMap(nodes []Node, prefix string) map[string]Node {
	result := make(map[string]Node)

	for _, node := range nodes {
		switch n := node.(type) {
		case *Interface:
			result[prefix+n.Name] = n
		case *TypeAlias:
			result[prefix+n.Name] = n
		case *Enum:
			result[prefix+n.Name] = n
		case *Namespace:
			// Recurse into namespace
			for k, v := range extractNodeMap(n.Children, prefix+n.Name+".") {
				result[k] = v
			}
		}
	}

	return result
}

func CompareNodes(expected, existing Node) []generators.PropertyDiff {
	switch e := expected.(type) {
	case *Interface:
		if ex, ok := existing.(*Interface); ok {
			return CompareInterfaces(e, ex)
		}
	case *TypeAlias:
		if ex, ok := existing.(*TypeAlias); ok {
			return CompareTypeAliases(e, ex)
		}
	case *Enum:
		if ex, ok := existing.(*Enum); ok {
			return CompareEnums(e, ex)
		}
	}

	// Type mismatch - report as changed
	return []generators.PropertyDiff{{
		Name:     "(type)",
		Status:   generators.Changed,
		OldValue: nodeTypeName(existing),
		NewValue: nodeTypeName(expected),
	}}
}

func nodeTypeName(n Node) string {
	switch n.(type) {
	case *Interface:
		return "interface"
	case *TypeAlias:
		return "type alias"
	case *Enum:
		return "enum"
	case *Namespace:
		return "namespace"
	default:
		return "unknown"
	}
}

// CompareInterfaces compares two Interface nodes
func CompareInterfaces(expected, existing *Interface) []generators.PropertyDiff {
	var diffs []generators.PropertyDiff

	// Build property maps
	expectedProps := make(map[string]Property)
	for _, p := range expected.Properties {
		expectedProps[p.Name] = p
	}

	existingProps := make(map[string]Property)
	for _, p := range existing.Properties {
		existingProps[p.Name] = p
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
		expectedProp, hasExpected := expectedProps[name]
		existingProp, hasExisting := existingProps[name]

		if hasExpected && !hasExisting {
			diffs = append(diffs, generators.PropertyDiff{
				Name:     name,
				Status:   generators.Added,
				NewValue: expectedProp.Type.Serialize(),
			})
		} else if !hasExpected && hasExisting {
			diffs = append(diffs, generators.PropertyDiff{
				Name:     name,
				Status:   generators.Removed,
				OldValue: existingProp.Type.Serialize(),
			})
		} else {
			// Compare property details
			if diff := CompareProperties(expectedProp, existingProp); diff != nil {
				diffs = append(diffs, *diff)
			}
		}
	}

	return diffs
}

// CompareEnums compares two Enum nodes
func CompareEnums(expected, existing *Enum) []generators.PropertyDiff {
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

// CompareProperties compares two Property values
func CompareProperties(expected, existing Property) *generators.PropertyDiff {
	expectedType := expected.Type.Serialize()
	existingType := existing.Type.Serialize()

	// Check type
	if expectedType != existingType {
		return &generators.PropertyDiff{
			Name:     expected.Name,
			Status:   generators.Changed,
			OldValue: existingType,
			NewValue: expectedType,
		}
	}

	// Check nullable
	if expected.Nullable != existing.Nullable {
		oldNull := "not nullable"
		newNull := "not nullable"
		if existing.Nullable {
			oldNull = "nullable"
		}
		if expected.Nullable {
			newNull = "nullable"
		}
		return &generators.PropertyDiff{
			Name:     expected.Name,
			Status:   generators.Changed,
			OldValue: oldNull,
			NewValue: newNull,
		}
	}

	// Check optional
	if expected.Optional != existing.Optional {
		oldOpt := "required"
		newOpt := "required"
		if existing.Optional {
			oldOpt = "optional"
		}
		if expected.Optional {
			newOpt = "optional"
		}
		return &generators.PropertyDiff{
			Name:     expected.Name,
			Status:   generators.Changed,
			OldValue: oldOpt,
			NewValue: newOpt,
		}
	}

	// Check readonly
	if expected.ReadOnly != existing.ReadOnly {
		oldRO := "mutable"
		newRO := "mutable"
		if existing.ReadOnly {
			oldRO = "readonly"
		}
		if expected.ReadOnly {
			newRO = "readonly"
		}
		return &generators.PropertyDiff{
			Name:     expected.Name,
			Status:   generators.Changed,
			OldValue: oldRO,
			NewValue: newRO,
		}
	}

	return nil
}
