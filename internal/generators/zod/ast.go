package zod

import (
	"fmt"
	"strings"
)

// Node is the base interface for all Zod AST nodes
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
	Header      string
	Import      string
	Schemas     []Node // SchemaDecl (all schemas flattened to top level)
	TypeSection []Node // InferredType, TSNamespace (namespaces with type exports)
	Interfaces  []Node // Interface, TSNamespace (for inferTypes mode)
}

func (f *File) Serialize() string {
	var parts []string
	if f.Header != "" {
		parts = append(parts, f.Header)
	}
	if f.Import != "" {
		parts = append(parts, f.Import)
	}

	// Schemas
	for _, schema := range f.Schemas {
		parts = append(parts, schema.Serialize())
	}

	// Type section
	if len(f.TypeSection) > 0 {
		parts = append(parts, "") // blank line
		for _, t := range f.TypeSection {
			parts = append(parts, t.Serialize())
		}
	}

	// Interfaces (for inferTypes)
	if len(f.Interfaces) > 0 {
		parts = append(parts, "") // blank line
		for _, iface := range f.Interfaces {
			parts = append(parts, iface.Serialize())
		}
	}

	return strings.Join(parts, "\n")
}

// SchemaDecl represents: export const fooSchema = <expr>
type SchemaDecl struct {
	Name   string
	Schema ZodExpr
}

func (s *SchemaDecl) Serialize() string {
	return fmt.Sprintf("export const %s = %s;", s.Name, s.Schema.Serialize())
}

// Namespace represents the old-style Zod namespace: export const X = { ... } as const
type Namespace struct {
	Name     string
	Children []Node // SchemaProperty, NestedNamespace
}

func (n *Namespace) Serialize() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("export const %s = {", n.Name))

	for _, child := range n.Children {
		childOutput := child.Serialize()
		for _, line := range strings.Split(childOutput, "\n") {
			if line != "" {
				lines = append(lines, "  "+line)
			}
		}
	}

	lines = append(lines, "} as const;")
	return strings.Join(lines, "\n")
}

type SchemaProperty struct {
	Name   string
	Schema ZodExpr
}

func (p *SchemaProperty) Serialize() string {
	return fmt.Sprintf("%s: %s,", p.Name, p.Schema.Serialize())
}

// NestedNamespace represents a nested object within a namespace
// Deprecated: Used only for parser backwards compatibility.
type NestedNamespace struct {
	Name     string
	Children []Node
}

func (n *NestedNamespace) Serialize() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("%s: {", n.Name))

	for _, child := range n.Children {
		childOutput := child.Serialize()
		for _, line := range strings.Split(childOutput, "\n") {
			if line != "" {
				lines = append(lines, "  "+line)
			}
		}
	}

	lines = append(lines, "},")
	return strings.Join(lines, "\n")
}

// InferredType represents: export type Foo = z.infer<typeof fooSchema>
type InferredType struct {
	Name      string
	SchemaRef string
	Export    bool
}

func (t *InferredType) Serialize() string {
	if t.Export {
		return fmt.Sprintf("export type %s = z.infer<typeof %s>;", t.Name, t.SchemaRef)
	}
	return fmt.Sprintf("type %s = z.infer<typeof %s>;", t.Name, t.SchemaRef)
}

// TypeSectionComment is a comment in the type section
type TypeSectionComment struct {
	Text string
}

func (c *TypeSectionComment) Serialize() string {
	return "// " + c.Text
}

// --- Zod Expressions ---

// ZodExpr is the interface for Zod schema expressions
type ZodExpr interface {
	Node
	isZodExpr()
}

// ZObject represents z.object({ ... })
type ZObject struct {
	Properties []ZProperty
}

func (o ZObject) Serialize() string {
	if len(o.Properties) == 0 {
		return "z.object({})"
	}

	var lines []string
	lines = append(lines, "z.object({")
	for _, prop := range o.Properties {
		lines = append(lines, "  "+prop.Serialize())
	}
	lines = append(lines, "})")
	return strings.Join(lines, "\n")
}
func (ZObject) isZodExpr() {}

// ZProperty represents a property in z.object
type ZProperty struct {
	Name   string
	Schema ZodExpr
}

func (p ZProperty) Serialize() string {
	return fmt.Sprintf("%s: %s,", p.Name, p.Schema.Serialize())
}

// ZArray represents z.array(<element>)
type ZArray struct {
	Element   ZodExpr
	Modifiers []Modifier
}

func (a ZArray) Serialize() string {
	result := fmt.Sprintf("z.array(%s)", a.Element.Serialize())
	for _, mod := range a.Modifiers {
		result += mod.Serialize()
	}
	return result
}
func (ZArray) isZodExpr() {}

// ZString represents z.string() with optional modifiers
type ZString struct {
	Modifiers []Modifier
}

func (s ZString) Serialize() string {
	result := "z.string()"
	for _, mod := range s.Modifiers {
		result += mod.Serialize()
	}
	return result
}
func (ZString) isZodExpr() {}

// ZNumber represents z.number() with optional modifiers
type ZNumber struct {
	Modifiers []Modifier
}

func (n ZNumber) Serialize() string {
	result := "z.number()"
	for _, mod := range n.Modifiers {
		result += mod.Serialize()
	}
	return result
}
func (ZNumber) isZodExpr() {}

// ZBoolean represents z.boolean()
type ZBoolean struct{}

func (ZBoolean) Serialize() string { return "z.boolean()" }
func (ZBoolean) isZodExpr()        {}

// ZEnum represents z.enum([...])
type ZEnum struct {
	Values []string
}

func (e ZEnum) Serialize() string {
	parts := make([]string, len(e.Values))
	for i, v := range e.Values {
		parts[i] = fmt.Sprintf("%q", v)
	}
	return fmt.Sprintf("z.enum([%s])", strings.Join(parts, ", "))
}
func (ZEnum) isZodExpr() {}

// ZUnion represents z.union([...])
type ZUnion struct {
	Schemas []ZodExpr
}

func (u ZUnion) Serialize() string {
	parts := make([]string, len(u.Schemas))
	for i, s := range u.Schemas {
		parts[i] = s.Serialize()
	}
	return fmt.Sprintf("z.union([%s])", strings.Join(parts, ", "))
}
func (ZUnion) isZodExpr() {}

// ZIntersection represents schema1.and(schema2)
type ZIntersection struct {
	Schemas []ZodExpr
}

func (i ZIntersection) Serialize() string {
	if len(i.Schemas) == 0 {
		return "z.unknown()"
	}
	if len(i.Schemas) == 1 {
		return i.Schemas[0].Serialize()
	}
	result := i.Schemas[0].Serialize()
	for _, s := range i.Schemas[1:] {
		result = fmt.Sprintf("%s.and(%s)", result, s.Serialize())
	}
	return result
}
func (ZIntersection) isZodExpr() {}

// ZLazy represents z.lazy(() => ref)
type ZLazy struct {
	Ref string
}

func (l ZLazy) Serialize() string {
	return fmt.Sprintf("z.lazy(() => %s)", l.Ref)
}
func (ZLazy) isZodExpr() {}

// ZRecord represents z.record(key, value)
type ZRecord struct {
	Key   ZodExpr
	Value ZodExpr
}

func (r ZRecord) Serialize() string {
	return fmt.Sprintf("z.record(%s, %s)", r.Key.Serialize(), r.Value.Serialize())
}
func (ZRecord) isZodExpr() {}

// ZLiteral represents z.literal(value)
type ZLiteral struct {
	Value any
}

func (l ZLiteral) Serialize() string {
	return fmt.Sprintf("z.literal(%s)", formatLiteralValue(l.Value))
}
func (ZLiteral) isZodExpr() {}

// ZNull represents z.null()
type ZNull struct{}

func (ZNull) Serialize() string { return "z.null()" }
func (ZNull) isZodExpr()        {}

// ZUnknown represents z.unknown()
type ZUnknown struct{}

func (ZUnknown) Serialize() string { return "z.unknown()" }
func (ZUnknown) isZodExpr()        {}

// ZWithModifiers wraps a ZodExpr with additional modifiers
type ZWithModifiers struct {
	Expr      ZodExpr
	Modifiers []Modifier
}

func (w ZWithModifiers) Serialize() string {
	result := w.Expr.Serialize()
	for _, mod := range w.Modifiers {
		result += mod.Serialize()
	}
	return result
}
func (ZWithModifiers) isZodExpr() {}

// --- Modifiers ---

// Modifier is the interface for chainable Zod methods
type Modifier interface {
	Serialize() string
}

// Optional represents .optional()
type Optional struct{}

func (Optional) Serialize() string { return ".optional()" }

// Nullable represents .nullable()
type Nullable struct{}

func (Nullable) Serialize() string { return ".nullable()" }

// Min represents .min(v)
type Min struct{ V int }

func (m Min) Serialize() string { return fmt.Sprintf(".min(%d)", m.V) }

// Max represents .max(v)
type Max struct{ V int }

func (m Max) Serialize() string { return fmt.Sprintf(".max(%d)", m.V) }

// MinFloat represents .min(v) for float values
type MinFloat struct{ V float64 }

func (m MinFloat) Serialize() string { return fmt.Sprintf(".min(%v)", m.V) }

// MaxFloat represents .max(v) for float values
type MaxFloat struct{ V float64 }

func (m MaxFloat) Serialize() string { return fmt.Sprintf(".max(%v)", m.V) }

// Email represents .email()
type Email struct{}

func (Email) Serialize() string { return ".email()" }

// Url represents .url()
type Url struct{}

func (Url) Serialize() string { return ".url()" }

// Uuid represents .uuid()
type Uuid struct{}

func (Uuid) Serialize() string { return ".uuid()" }

// Datetime represents .datetime()
type Datetime struct{}

func (Datetime) Serialize() string { return ".datetime()" }

// Date represents .date()
type Date struct{}

func (Date) Serialize() string { return ".date()" }

// Int represents .int()
type Int struct{}

func (Int) Serialize() string { return ".int()" }

// Regex represents .regex(/pattern/)
type Regex struct{ Pattern string }

func (r Regex) Serialize() string { return fmt.Sprintf(".regex(/%s/)", r.Pattern) }

// --- TypeScript Interface Types (for inferTypes mode) ---

// TSNamespace represents a TypeScript namespace for interfaces
type TSNamespace struct {
	Name     string
	Children []Node
}

func (n *TSNamespace) Serialize() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("export namespace %s {", n.Name))

	for _, child := range n.Children {
		childOutput := child.Serialize()
		for _, line := range strings.Split(childOutput, "\n") {
			if line != "" {
				lines = append(lines, "  "+line)
			}
		}
	}

	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

// Interface represents a TypeScript interface
type Interface struct {
	Name        string
	Description Description
	Properties  []InterfaceProperty
}

func (i *Interface) Serialize() string {
	var lines []string

	if desc := i.Description.Serialize(); desc != "" {
		lines = append(lines, desc)
	}

	lines = append(lines, fmt.Sprintf("export interface %s {", i.Name))
	for _, prop := range i.Properties {
		propOutput := prop.Serialize()
		for _, line := range strings.Split(propOutput, "\n") {
			lines = append(lines, "  "+line)
		}
	}
	lines = append(lines, "}")

	return strings.Join(lines, "\n")
}

// InterfaceProperty represents a property in an interface
type InterfaceProperty struct {
	Name        string
	IndexType   string // e.g., "UserSchemaType"
	IndexKey    string // e.g., "email"
	Optional    bool
	ReadOnly    bool
	Description Description
}

func (p InterfaceProperty) Serialize() string {
	var lines []string

	if desc := p.Description.Serialize(); desc != "" {
		lines = append(lines, desc)
	}

	var propLine strings.Builder
	if p.ReadOnly {
		propLine.WriteString("readonly ")
	}
	propLine.WriteString(p.Name)
	if p.Optional {
		propLine.WriteString("?")
	}
	propLine.WriteString(fmt.Sprintf(": %s[\"%s\"];", p.IndexType, p.IndexKey))

	lines = append(lines, propLine.String())

	return strings.Join(lines, "\n")
}

// TypeAlias represents: export type Foo = Bar
type TypeAlias struct {
	Name        string
	Description Description
	Type        string
	Export      bool
}

func (t *TypeAlias) Serialize() string {
	var lines []string

	if desc := t.Description.Serialize(); desc != "" {
		lines = append(lines, desc)
	}

	if t.Export {
		lines = append(lines, fmt.Sprintf("export type %s = %s;", t.Name, t.Type))
	} else {
		lines = append(lines, fmt.Sprintf("type %s = %s;", t.Name, t.Type))
	}

	return strings.Join(lines, "\n")
}

// --- Helpers ---

func formatLiteralValue(val any) string {
	switch v := val.(type) {
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
