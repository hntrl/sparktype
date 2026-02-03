package typescript

import (
	"fmt"
	"strings"
)

// Node is the base interface for all TypeScript AST nodes
type Node interface {
	Serialize() string
}

// Description serializes to JSDoc format: /** ... */
type Description string

func (d Description) Serialize() string {
	if d == "" {
		return ""
	}
	s := string(d)
	if !strings.Contains(s, "\n") && len(s) <= 60 {
		return fmt.Sprintf("/** %s */", s)
	}
	// Multi-line JSDoc
	var lines []string
	lines = append(lines, "/**")
	for _, line := range strings.Split(s, "\n") {
		lines = append(lines, " * "+line)
	}
	lines = append(lines, " */")
	return strings.Join(lines, "\n")
}

// File is the root AST node
type File struct {
	Header string
	Nodes  []Node
}

func (f *File) Serialize() string {
	var parts []string
	if f.Header != "" {
		parts = append(parts, strings.TrimSuffix(f.Header, "\n"))
	}
	for _, node := range f.Nodes {
		parts = append(parts, node.Serialize())
	}
	return strings.Join(parts, "\n\n") + "\n"
}

// Namespace represents a TypeScript namespace block
type Namespace struct {
	Name     string
	Children []Node
}

func (n *Namespace) Serialize() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("export namespace %s {", n.Name))

	for _, child := range n.Children {
		// Add blank line before each child (provides spacing after opening brace and between children)
		lines = append(lines, "")
		// Indent each line of child output
		childOutput := child.Serialize()
		for _, line := range strings.Split(childOutput, "\n") {
			if line != "" {
				lines = append(lines, "  "+line)
			} else {
				lines = append(lines, "")
			}
		}
	}

	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

// Interface represents an interface declaration
type Interface struct {
	Name        string
	Description Description
	Properties  []Property
	ExportStyle string // "interface" or "type"
}

func (i *Interface) Serialize() string {
	var lines []string

	// JSDoc
	if desc := i.Description.Serialize(); desc != "" {
		lines = append(lines, desc)
	}

	// Opening
	style := i.ExportStyle
	if style == "" {
		style = "interface"
	}
	if style == "type" {
		lines = append(lines, fmt.Sprintf("export type %s = {", i.Name))
	} else {
		lines = append(lines, fmt.Sprintf("export interface %s {", i.Name))
	}

	// Properties - add blank line between properties that have descriptions
	for idx, prop := range i.Properties {
		// Add blank line before property if it has a description (except first)
		if idx > 0 && prop.Description != "" {
			lines = append(lines, "")
		}
		propOutput := prop.Serialize()
		for _, line := range strings.Split(propOutput, "\n") {
			lines = append(lines, "  "+line)
		}
	}

	// Closing
	if style == "type" {
		lines = append(lines, "};")
	} else {
		lines = append(lines, "}")
	}

	return strings.Join(lines, "\n")
}

// Enum represents an enum declaration
type Enum struct {
	Name        string
	Description Description
	Members     []EnumMember
}

func (e *Enum) Serialize() string {
	var lines []string

	// JSDoc
	if desc := e.Description.Serialize(); desc != "" {
		lines = append(lines, desc)
	}

	lines = append(lines, fmt.Sprintf("export enum %s {", e.Name))
	for _, member := range e.Members {
		lines = append(lines, "  "+member.Serialize())
	}
	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}

// EnumMember represents a single enum member
type EnumMember struct {
	Key   string
	Value string
}

func (m EnumMember) Serialize() string {
	return fmt.Sprintf("%s = %q,", m.Key, m.Value)
}

// TypeAlias represents a type alias: export type X = Y
type TypeAlias struct {
	Name        string
	Description Description
	Type        TypeExpr
}

func (t *TypeAlias) Serialize() string {
	var lines []string

	// JSDoc
	if desc := t.Description.Serialize(); desc != "" {
		lines = append(lines, desc)
	}

	lines = append(lines, fmt.Sprintf("export type %s = %s;", t.Name, t.Type.Serialize()))

	return strings.Join(lines, "\n")
}

// Property represents an interface property
type Property struct {
	Name        string
	Type        TypeExpr
	Optional    bool
	ReadOnly    bool
	Nullable    bool
	Description Description
}

func (p Property) Serialize() string {
	var lines []string

	// JSDoc
	if desc := p.Description.Serialize(); desc != "" {
		lines = append(lines, desc)
	}

	// Build property line
	var propLine strings.Builder
	if p.ReadOnly {
		propLine.WriteString("readonly ")
	}
	propLine.WriteString(p.Name)
	if p.Optional {
		propLine.WriteString("?")
	}
	propLine.WriteString(": ")
	propLine.WriteString(p.Type.Serialize())
	if p.Nullable {
		propLine.WriteString(" | null")
	}
	propLine.WriteString(";")

	lines = append(lines, propLine.String())

	return strings.Join(lines, "\n")
}

// TypeExpr is the interface for type expressions
type TypeExpr interface {
	Node
	isTypeExpr()
}

// StringType represents the string type
type StringType struct{}

func (StringType) Serialize() string { return "string" }
func (StringType) isTypeExpr()       {}

// NumberType represents the number type
type NumberType struct{}

func (NumberType) Serialize() string { return "number" }
func (NumberType) isTypeExpr()       {}

// BooleanType represents the boolean type
type BooleanType struct{}

func (BooleanType) Serialize() string { return "boolean" }
func (BooleanType) isTypeExpr()       {}

// NullType represents the null type
type NullType struct{}

func (NullType) Serialize() string { return "null" }
func (NullType) isTypeExpr()       {}

// UnknownType represents the unknown type
type UnknownType struct{}

func (UnknownType) Serialize() string { return "unknown" }
func (UnknownType) isTypeExpr()       {}

// WidenedStringType represents (string & {}) for widening string literal unions
// This allows any string while still providing autocomplete for known literals
type WidenedStringType struct{}

func (WidenedStringType) Serialize() string { return "(string & {})" }
func (WidenedStringType) isTypeExpr()       {}

// ArrayType represents Array<T>
type ArrayType struct {
	Element TypeExpr
}

func (a ArrayType) Serialize() string {
	return fmt.Sprintf("Array<%s>", a.Element.Serialize())
}
func (ArrayType) isTypeExpr() {}

// UnionType represents T | U | V
type UnionType struct {
	Types []TypeExpr
}

func (u UnionType) Serialize() string {
	parts := make([]string, len(u.Types))
	for i, t := range u.Types {
		parts[i] = t.Serialize()
	}
	return strings.Join(parts, " | ")
}
func (UnionType) isTypeExpr() {}

// IntersectionType represents T & U & V
type IntersectionType struct {
	Types []TypeExpr
}

func (i IntersectionType) Serialize() string {
	parts := make([]string, len(i.Types))
	for idx, t := range i.Types {
		parts[idx] = t.Serialize()
	}
	return strings.Join(parts, " & ")
}
func (IntersectionType) isTypeExpr() {}

// RecordType represents Record<K, V>
type RecordType struct {
	Key   TypeExpr
	Value TypeExpr
}

func (r RecordType) Serialize() string {
	return fmt.Sprintf("Record<%s, %s>", r.Key.Serialize(), r.Value.Serialize())
}
func (RecordType) isTypeExpr() {}

// ReferenceType represents a reference to another type by name
type ReferenceType struct {
	Name string
}

func (r ReferenceType) Serialize() string { return r.Name }
func (ReferenceType) isTypeExpr()         {}

// LiteralType represents a literal type like "foo" or 42
type LiteralType struct {
	Value any
}

func (l LiteralType) Serialize() string {
	switch v := l.Value.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case float64:
		if v == float64(int(v)) {
			return fmt.Sprintf("%d", int(v))
		}
		return fmt.Sprintf("%g", v)
	case bool:
		return fmt.Sprintf("%t", v)
	case nil:
		return "null"
	default:
		return fmt.Sprintf("%v", v)
	}
}
func (LiteralType) isTypeExpr() {}

// IndexedType represents Base["Index"] for indexed access types
type IndexedType struct {
	Base  string
	Index string
}

func (i IndexedType) Serialize() string {
	return fmt.Sprintf("%s[\"%s\"]", i.Base, i.Index)
}
func (IndexedType) isTypeExpr() {}

// ObjectType represents an inline object type { prop: type; ... }
type ObjectType struct {
	Properties []Property
}

func (o ObjectType) Serialize() string {
	if len(o.Properties) == 0 {
		return "{}"
	}
	parts := make([]string, len(o.Properties))
	for i, prop := range o.Properties {
		var sb strings.Builder
		sb.WriteString(prop.Name)
		if prop.Optional {
			sb.WriteString("?")
		}
		sb.WriteString(": ")
		sb.WriteString(prop.Type.Serialize())
		parts[i] = sb.String()
	}
	return "{ " + strings.Join(parts, "; ") + " }"
}
func (ObjectType) isTypeExpr() {}
