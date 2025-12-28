package python

import (
	"fmt"
	"strings"
)

// Node is the base interface for all Python AST nodes
type Node interface {
	Serialize() string
}

// Docstring serializes to Python docstring format: """..."""
// Used for class-level documentation that appears in __doc__
type Docstring string

func (d Docstring) Serialize() string {
	if d == "" {
		return ""
	}
	s := string(d)
	// Single line docstring
	if !strings.Contains(s, "\n") && len(s) <= 60 {
		return fmt.Sprintf(`"""%s"""`, s)
	}
	// Multi-line docstring
	var sb strings.Builder
	sb.WriteString(`"""`)
	sb.WriteString("\n")
	for _, line := range strings.Split(s, "\n") {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	sb.WriteString(`"""`)
	return sb.String()
}

// Comment serializes to Python comment format: # ...
// Used for inline documentation (fields, type aliases)
type Comment string

func (c Comment) Serialize() string {
	if c == "" {
		return ""
	}
	lines := strings.Split(string(c), "\n")
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = "# " + line
	}
	return strings.Join(result, "\n")
}

// File is the root AST node
type File struct {
	Header  string
	Imports []string
	Classes []Node
}

func (f *File) Serialize() string {
	var parts []string

	if f.Header != "" {
		parts = append(parts, f.Header)
	}

	if len(f.Imports) > 0 {
		parts = append(parts, strings.Join(f.Imports, "\n"))
	}

	for _, class := range f.Classes {
		parts = append(parts, class.Serialize())
	}

	return strings.Join(parts, "\n\n") + "\n"
}

// TypedDict represents a TypedDict class definition
type TypedDict struct {
	Name      string
	Docstring Docstring
	Fields    []Field
	Total     bool // If false, adds total=False
}

func (t *TypedDict) Serialize() string {
	var lines []string

	// Class definition
	totalStr := ""
	if !t.Total {
		totalStr = ", total=False"
	}
	lines = append(lines, fmt.Sprintf("class %s(TypedDict%s):", t.Name, totalStr))

	// Docstring goes inside the class body
	if docstring := t.Docstring.Serialize(); docstring != "" {
		for _, line := range strings.Split(docstring, "\n") {
			lines = append(lines, "    "+line)
		}
	}

	// Fields
	if len(t.Fields) == 0 && t.Docstring == "" {
		lines = append(lines, "    pass")
	} else if len(t.Fields) == 0 {
		// Docstring is present but no fields - that's valid, no pass needed
	} else {
		for i, field := range t.Fields {
			// Add blank line between fields (after docstring of previous field)
			if i > 0 && t.Fields[i-1].Docstring != "" {
				lines = append(lines, "")
			}
			fieldOutput := field.Serialize()
			for _, line := range strings.Split(fieldOutput, "\n") {
				lines = append(lines, "    "+line)
			}
		}
	}

	return strings.Join(lines, "\n")
}

// Field represents a TypedDict field
type Field struct {
	Name      string
	Type      PyTypeExpr
	Docstring Docstring
}

func (f Field) Serialize() string {
	var lines []string

	// Field annotation
	lines = append(lines, fmt.Sprintf("%s: %s", f.Name, f.Type.Serialize()))

	// Attribute docstring follows the field (PEP 257 style)
	if docstring := f.Docstring.Serialize(); docstring != "" {
		lines = append(lines, docstring)
	}

	return strings.Join(lines, "\n")
}

// EnumClass represents a Python Enum class
type EnumClass struct {
	Name      string
	Docstring Docstring
	Members   []EnumMember
}

func (e *EnumClass) Serialize() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("class %s(str, Enum):", e.Name))

	// Docstring goes inside the class body
	if docstring := e.Docstring.Serialize(); docstring != "" {
		for _, line := range strings.Split(docstring, "\n") {
			lines = append(lines, "    "+line)
		}
	}

	for _, member := range e.Members {
		lines = append(lines, "    "+member.Serialize())
	}

	return strings.Join(lines, "\n")
}

// EnumMember represents a single enum member
type EnumMember struct {
	Key   string
	Value string
}

func (m EnumMember) Serialize() string {
	return fmt.Sprintf("%s = %q", m.Key, m.Value)
}

// TypeAlias represents a Python type alias: Name = Type
type TypeAlias struct {
	Name    string
	Comment Comment
	Type    PyTypeExpr
}

func (t *TypeAlias) Serialize() string {
	var lines []string

	if comment := t.Comment.Serialize(); comment != "" {
		lines = append(lines, comment)
	}

	lines = append(lines, fmt.Sprintf("%s = %s", t.Name, t.Type.Serialize()))

	return strings.Join(lines, "\n")
}

// --- Python Type Expressions ---

// PyTypeExpr is the interface for Python type expressions
type PyTypeExpr interface {
	Node
	isPyTypeExpr()
}

// StrType represents str
type StrType struct{}

func (StrType) Serialize() string { return "str" }
func (StrType) isPyTypeExpr()     {}

// IntType represents int
type IntType struct{}

func (IntType) Serialize() string { return "int" }
func (IntType) isPyTypeExpr()     {}

// FloatType represents float
type FloatType struct{}

func (FloatType) Serialize() string { return "float" }
func (FloatType) isPyTypeExpr()     {}

// BoolType represents bool
type BoolType struct{}

func (BoolType) Serialize() string { return "bool" }
func (BoolType) isPyTypeExpr()     {}

// NoneType represents None
type NoneType struct{}

func (NoneType) Serialize() string { return "None" }
func (NoneType) isPyTypeExpr()     {}

// AnyType represents Any
type AnyType struct{}

func (AnyType) Serialize() string { return "Any" }
func (AnyType) isPyTypeExpr()     {}

// ListType represents List[T]
type ListType struct {
	Element PyTypeExpr
}

func (l ListType) Serialize() string {
	return fmt.Sprintf("List[%s]", l.Element.Serialize())
}
func (ListType) isPyTypeExpr() {}

// DictType represents Dict[K, V]
type DictType struct {
	Key   PyTypeExpr
	Value PyTypeExpr
}

func (d DictType) Serialize() string {
	return fmt.Sprintf("Dict[%s, %s]", d.Key.Serialize(), d.Value.Serialize())
}
func (DictType) isPyTypeExpr() {}

// UnionType represents Union[T1, T2, ...]
type UnionType struct {
	Types []PyTypeExpr
}

func (u UnionType) Serialize() string {
	if len(u.Types) == 0 {
		return "Any"
	}
	if len(u.Types) == 1 {
		return u.Types[0].Serialize()
	}
	parts := make([]string, len(u.Types))
	for i, t := range u.Types {
		parts[i] = t.Serialize()
	}
	return fmt.Sprintf("Union[%s]", strings.Join(parts, ", "))
}
func (UnionType) isPyTypeExpr() {}

// OptionalType represents T | None
type OptionalType struct {
	Inner PyTypeExpr
}

func (o OptionalType) Serialize() string {
	return fmt.Sprintf("%s | None", o.Inner.Serialize())
}
func (OptionalType) isPyTypeExpr() {}

// NotRequiredType represents NotRequired[T]
type NotRequiredType struct {
	Inner PyTypeExpr
}

func (n NotRequiredType) Serialize() string {
	return fmt.Sprintf("NotRequired[%s]", n.Inner.Serialize())
}
func (NotRequiredType) isPyTypeExpr() {}

// LiteralType represents Literal[v1, v2, ...]
type LiteralType struct {
	Values []any
}

func (l LiteralType) Serialize() string {
	parts := make([]string, len(l.Values))
	for i, v := range l.Values {
		parts[i] = formatPythonValue(v)
	}
	if len(parts) == 1 {
		return fmt.Sprintf("Literal[%s]", parts[0])
	}
	return fmt.Sprintf("Union[%s]", strings.Join(formatLiteralParts(l.Values), ", "))
}
func (LiteralType) isPyTypeExpr() {}

// ReferenceType represents a reference to another type
type ReferenceType struct {
	Name string
}

func (r ReferenceType) Serialize() string {
	// Quote for forward reference
	return fmt.Sprintf("%q", r.Name)
}
func (ReferenceType) isPyTypeExpr() {}

// --- Helpers ---

func formatPythonValue(val any) string {
	switch v := val.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	case float64:
		if v == float64(int(v)) {
			return fmt.Sprintf("%d", int(v))
		}
		return fmt.Sprintf("%g", v)
	case bool:
		if v {
			return "True"
		}
		return "False"
	case nil:
		return "None"
	default:
		return fmt.Sprintf("%v", v)
	}
}

func formatLiteralParts(values []any) []string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = fmt.Sprintf("Literal[%s]", formatPythonValue(v))
	}
	return parts
}
