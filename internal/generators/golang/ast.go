package golang

import (
	"fmt"
	"strings"
)

// Node is the base interface for all Go AST nodes
type Node interface {
	Serialize() string
}

// Comment serializes to Go line comment format: // ...
type Comment string

func (c Comment) Serialize() string {
	if c == "" {
		return ""
	}
	lines := strings.Split(string(c), "\n")
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = "// " + line
	}
	return strings.Join(result, "\n")
}

// File is the root AST node
type File struct {
	Header  string
	Package string
	Imports []string
	Types   []Node
}

func (f *File) Serialize() string {
	var parts []string

	if f.Header != "" {
		parts = append(parts, f.Header)
	}

	parts = append(parts, fmt.Sprintf("package %s", f.Package))

	if len(f.Imports) > 0 {
		if len(f.Imports) == 1 {
			parts = append(parts, fmt.Sprintf("import %q", f.Imports[0]))
		} else {
			imports := make([]string, len(f.Imports))
			for i, imp := range f.Imports {
				imports[i] = fmt.Sprintf("\t%q", imp)
			}
			parts = append(parts, "import (\n"+strings.Join(imports, "\n")+"\n)")
		}
	}

	for _, t := range f.Types {
		parts = append(parts, t.Serialize())
	}

	return strings.Join(parts, "\n\n") + "\n"
}

// Struct represents a Go struct type definition
type Struct struct {
	Name    string
	Comment Comment
	Fields  []StructField
}

func (s *Struct) Serialize() string {
	var lines []string

	if comment := s.Comment.Serialize(); comment != "" {
		lines = append(lines, comment)
	}

	lines = append(lines, fmt.Sprintf("type %s struct {", s.Name))

	for _, field := range s.Fields {
		fieldOutput := field.Serialize()
		for _, line := range strings.Split(fieldOutput, "\n") {
			lines = append(lines, "\t"+line)
		}
	}

	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}

// StructField represents a field in a struct
type StructField struct {
	Name    string
	Type    GoTypeExpr
	JSONTag string
	Comment Comment
}

func (f StructField) Serialize() string {
	var lines []string

	if comment := f.Comment.Serialize(); comment != "" {
		lines = append(lines, comment)
	}

	line := fmt.Sprintf("%s %s", f.Name, f.Type.Serialize())
	if f.JSONTag != "" {
		line += fmt.Sprintf(" `json:%q`", f.JSONTag)
	}
	lines = append(lines, line)

	return strings.Join(lines, "\n")
}

// TypeAlias represents a Go type alias: type Name = Type
type TypeAlias struct {
	Name    string
	Comment Comment
	Type    GoTypeExpr
}

func (t *TypeAlias) Serialize() string {
	var lines []string

	if comment := t.Comment.Serialize(); comment != "" {
		lines = append(lines, comment)
	}

	lines = append(lines, fmt.Sprintf("type %s = %s", t.Name, t.Type.Serialize()))

	return strings.Join(lines, "\n")
}

// TypeDef represents a Go type definition: type Name Type
type TypeDef struct {
	Name    string
	Comment Comment
	Type    GoTypeExpr
}

func (t *TypeDef) Serialize() string {
	var lines []string

	if comment := t.Comment.Serialize(); comment != "" {
		lines = append(lines, comment)
	}

	lines = append(lines, fmt.Sprintf("type %s %s", t.Name, t.Type.Serialize()))

	return strings.Join(lines, "\n")
}

// ConstBlock represents a const declaration block
type ConstBlock struct {
	TypeName string
	Consts   []ConstDecl
}

func (c *ConstBlock) Serialize() string {
	var lines []string
	lines = append(lines, "const (")
	for _, decl := range c.Consts {
		lines = append(lines, "\t"+decl.Serialize())
	}
	lines = append(lines, ")")
	return strings.Join(lines, "\n")
}

// ConstDecl represents a single constant declaration
type ConstDecl struct {
	Name     string
	TypeName string
	Value    string
}

func (c ConstDecl) Serialize() string {
	return fmt.Sprintf("%s %s = %q", c.Name, c.TypeName, c.Value)
}

// EnumDef represents a Go enum (type + const block)
type EnumDef struct {
	Name    string
	Comment Comment
	Values  []EnumValue
}

func (e *EnumDef) Serialize() string {
	var lines []string

	if comment := e.Comment.Serialize(); comment != "" {
		lines = append(lines, comment)
	}

	lines = append(lines, fmt.Sprintf("type %s string", e.Name))
	lines = append(lines, "")
	lines = append(lines, "const (")

	for _, v := range e.Values {
		lines = append(lines, fmt.Sprintf("\t%s%s %s = %q", e.Name, v.Name, e.Name, v.Value))
	}

	lines = append(lines, ")")

	return strings.Join(lines, "\n")
}

// EnumValue represents a single enum value
type EnumValue struct {
	Name  string
	Value string
}

// InterfaceComment represents a comment-only type for documenting mixed enums
type InterfaceComment struct {
	Name    string
	Comment Comment
	Values  []string
}

func (i *InterfaceComment) Serialize() string {
	var lines []string

	if comment := i.Comment.Serialize(); comment != "" {
		lines = append(lines, comment)
	}

	lines = append(lines, fmt.Sprintf("type %s interface{}", i.Name))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("// Valid values for %s:", i.Name))
	for _, v := range i.Values {
		lines = append(lines, fmt.Sprintf("//   %s", v))
	}

	return strings.Join(lines, "\n")
}

// --- Go Type Expressions ---

// GoTypeExpr is the interface for Go type expressions
type GoTypeExpr interface {
	Node
	isGoTypeExpr()
	NeedsTimeImport() bool
}

// GoString represents string
type GoString struct{}

func (GoString) Serialize() string     { return "string" }
func (GoString) isGoTypeExpr()         {}
func (GoString) NeedsTimeImport() bool { return false }

// GoInt represents int
type GoInt struct{}

func (GoInt) Serialize() string     { return "int" }
func (GoInt) isGoTypeExpr()         {}
func (GoInt) NeedsTimeImport() bool { return false }

// GoInt32 represents int32
type GoInt32 struct{}

func (GoInt32) Serialize() string     { return "int32" }
func (GoInt32) isGoTypeExpr()         {}
func (GoInt32) NeedsTimeImport() bool { return false }

// GoInt64 represents int64
type GoInt64 struct{}

func (GoInt64) Serialize() string     { return "int64" }
func (GoInt64) isGoTypeExpr()         {}
func (GoInt64) NeedsTimeImport() bool { return false }

// GoFloat64 represents float64
type GoFloat64 struct{}

func (GoFloat64) Serialize() string     { return "float64" }
func (GoFloat64) isGoTypeExpr()         {}
func (GoFloat64) NeedsTimeImport() bool { return false }

// GoBool represents bool
type GoBool struct{}

func (GoBool) Serialize() string     { return "bool" }
func (GoBool) isGoTypeExpr()         {}
func (GoBool) NeedsTimeImport() bool { return false }

// GoTime represents time.Time
type GoTime struct{}

func (GoTime) Serialize() string     { return "time.Time" }
func (GoTime) isGoTypeExpr()         {}
func (GoTime) NeedsTimeImport() bool { return true }

// GoBytes represents []byte
type GoBytes struct{}

func (GoBytes) Serialize() string     { return "[]byte" }
func (GoBytes) isGoTypeExpr()         {}
func (GoBytes) NeedsTimeImport() bool { return false }

// GoSlice represents []T
type GoSlice struct {
	Element GoTypeExpr
}

func (s GoSlice) Serialize() string {
	return "[]" + s.Element.Serialize()
}
func (GoSlice) isGoTypeExpr()           {}
func (s GoSlice) NeedsTimeImport() bool { return s.Element.NeedsTimeImport() }

// GoMap represents map[K]V
type GoMap struct {
	Key   GoTypeExpr
	Value GoTypeExpr
}

func (m GoMap) Serialize() string {
	return fmt.Sprintf("map[%s]%s", m.Key.Serialize(), m.Value.Serialize())
}
func (GoMap) isGoTypeExpr()           {}
func (m GoMap) NeedsTimeImport() bool { return m.Key.NeedsTimeImport() || m.Value.NeedsTimeImport() }

// GoPointer represents *T
type GoPointer struct {
	Inner GoTypeExpr
}

func (p GoPointer) Serialize() string {
	return "*" + p.Inner.Serialize()
}
func (GoPointer) isGoTypeExpr()           {}
func (p GoPointer) NeedsTimeImport() bool { return p.Inner.NeedsTimeImport() }

// GoInterface represents interface{}
type GoInterface struct{}

func (GoInterface) Serialize() string     { return "interface{}" }
func (GoInterface) isGoTypeExpr()         {}
func (GoInterface) NeedsTimeImport() bool { return false }

// GoReference represents a reference to another type by name
type GoReference struct {
	Name string
}

func (r GoReference) Serialize() string   { return r.Name }
func (GoReference) isGoTypeExpr()         {}
func (GoReference) NeedsTimeImport() bool { return false }

// Helper to check if a type needs a pointer
func IsPointerType(t GoTypeExpr) bool {
	switch t.(type) {
	case GoPointer, GoSlice, GoMap, GoInterface:
		return true
	default:
		return false
	}
}
