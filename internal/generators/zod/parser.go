package zod

import (
	"strconv"
	"strings"

	"github.com/hntrl/sparktype/internal/parser"
)

// Zod-specific tokens
const (
	EXPORT parser.Token = parser.KEYWORD_START + iota
	CONST
	TYPE
	NAMESPACE
	AS
	TYPEOF
	READONLY
	INTERFACE
)

var zodKeywords = map[string]parser.Token{
	"export":    EXPORT,
	"const":     CONST,
	"type":      TYPE,
	"namespace": NAMESPACE,
	"as":        AS,
	"typeof":    TYPEOF,
	"readonly":  READONLY,
	"interface": INTERFACE,
}

func zodKeywordLookup(lit string) parser.Token {
	if tok, ok := zodKeywords[lit]; ok {
		return tok
	}
	return parser.ILLEGAL
}

// Parser parses Zod source into AST nodes
type Parser struct {
	p *parser.Parser
}

// NewParser creates a new Zod parser
func NewParser(input string) *Parser {
	p := parser.NewParser(input)
	p.SetKeywordLookup(zodKeywordLookup)
	return &Parser{p: p}
}

// ParseFile parses a Zod file from a Parser
func (p *Parser) ParseFile() (*File, error) {
	file := &File{}

	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.EOF {
			break
		}

		if item.Tok == parser.COMMENT {
			continue
		}

		if item.Tok == EXPORT {
			node, section, err := p.ParseExport()
			if err != nil {
				return nil, err
			}
			if node != nil {
				switch section {
				case "schema":
					file.Schemas = append(file.Schemas, node)
				case "type":
					file.TypeSection = append(file.TypeSection, node)
				case "interface":
					file.Interfaces = append(file.Interfaces, node)
				}
			}
		} else if item.Tok == TYPE {
			// Non-exported type
			node, err := p.ParseTypeAliasDecl(false)
			if err != nil {
				return nil, err
			}
			if node != nil {
				file.TypeSection = append(file.TypeSection, node)
			}
		}
	}

	return file, nil
}

func (p *Parser) ParseExport() (Node, string, error) {
	item := p.p.ScanIgnoreWhitespace()

	switch item.Tok {
	case CONST:
		return p.ParseConstDecl()
	case TYPE:
		node, err := p.ParseTypeAliasDecl(true)
		return node, "type", err
	case NAMESPACE:
		node, err := p.ParseTSNamespaceDecl()
		return node, "interface", err
	case INTERFACE:
		node, err := p.ParseInterfaceDecl()
		return node, "interface", err
	default:
		p.p.Unscan()
		p.SkipToNextExport()
		return nil, "", nil
	}
}

func (p *Parser) SkipToNextExport() {
	depth := 0
	for {
		item := p.p.Scan()
		if item.Tok == parser.EOF {
			return
		}
		if item.Tok == parser.LBRACE {
			depth++
		} else if item.Tok == parser.RBRACE {
			depth--
			if depth < 0 {
				p.p.Unscan()
				return
			}
		} else if depth == 0 && item.Tok == EXPORT {
			p.p.Unscan()
			return
		}
	}
}

func (p *Parser) SkipUntil(tokens ...parser.Token) {
	for {
		item := p.p.Scan()
		if item.Tok == parser.EOF {
			return
		}
		for _, t := range tokens {
			if item.Tok == t {
				p.p.Unscan()
				return
			}
		}
	}
}

func (p *Parser) ParseConstDecl() (Node, string, error) {
	// Parse name
	nameItem, err := p.p.Expect(parser.IDENT)
	if err != nil {
		return nil, "", err
	}
	name := nameItem.Lit

	// Expect =
	if _, err := p.p.Expect(parser.EQUALS); err != nil {
		return nil, "", err
	}

	// Check what follows
	item := p.p.ScanIgnoreWhitespace()

	if item.Tok == parser.LBRACE {
		// It's a namespace: export const X = { ... } as const
		ns, err := p.ParseNamespaceBody(name)
		return ns, "schema", err
	}

	if item.Tok == parser.IDENT && item.Lit == "z" {
		// It's a schema: export const X = z.something()
		p.p.Unscan()
		schema, err := p.ParseZodExpr()
		if err != nil {
			return nil, "", err
		}

		// Skip trailing semicolon
		if p.p.Match(parser.SEMICOLON) {
			p.p.Scan()
		}

		return &SchemaDecl{Name: name, Schema: schema}, "schema", nil
	}

	// Unknown, skip
	p.p.Unscan()
	p.SkipToNextExport()
	return nil, "", nil
}

func (p *Parser) ParseNamespaceBody(name string) (*Namespace, error) {
	ns := &Namespace{Name: name}

	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.RBRACE {
			break
		}

		if item.Tok == parser.EOF {
			return nil, parser.NewParseError(item.Pos, "unexpected EOF in namespace")
		}

		if item.Tok == parser.COMMENT {
			continue
		}

		if item.Tok == parser.IDENT {
			propName := item.Lit

			// Expect :
			if _, err := p.p.Expect(parser.COLON); err != nil {
				return nil, err
			}

			// Check if nested namespace or schema
			nextItem := p.p.ScanIgnoreWhitespace()

			if nextItem.Tok == parser.LBRACE {
				// Nested namespace
				nested, err := p.ParseNestedNamespaceBody(propName)
				if err != nil {
					return nil, err
				}
				ns.Children = append(ns.Children, nested)
			} else if nextItem.Tok == parser.IDENT && nextItem.Lit == "z" {
				// Schema property
				p.p.Unscan()
				schema, err := p.ParseZodExpr()
				if err != nil {
					return nil, err
				}
				ns.Children = append(ns.Children, &SchemaProperty{Name: propName, Schema: schema})

				// Skip comma
				if p.p.Match(parser.COMMA) {
					p.p.Scan()
				}
			} else {
				p.p.Unscan()
				p.SkipUntil(parser.COMMA, parser.RBRACE)
			}
		}
	}

	// Skip "as const"
	item := p.p.ScanIgnoreWhitespace()
	if item.Tok == AS {
		p.p.ScanIgnoreWhitespace() // consume "const"
	} else {
		p.p.Unscan()
	}

	// Skip semicolon
	if p.p.Match(parser.SEMICOLON) {
		p.p.Scan()
	}

	return ns, nil
}

func (p *Parser) ParseNestedNamespaceBody(name string) (*NestedNamespace, error) {
	nested := &NestedNamespace{Name: name}

	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.RBRACE {
			break
		}

		if item.Tok == parser.EOF {
			return nil, parser.NewParseError(item.Pos, "unexpected EOF in nested namespace")
		}

		if item.Tok == parser.COMMENT {
			continue
		}

		if item.Tok == parser.IDENT {
			propName := item.Lit

			if _, err := p.p.Expect(parser.COLON); err != nil {
				return nil, err
			}

			nextItem := p.p.ScanIgnoreWhitespace()

			if nextItem.Tok == parser.LBRACE {
				innerNested, err := p.ParseNestedNamespaceBody(propName)
				if err != nil {
					return nil, err
				}
				nested.Children = append(nested.Children, innerNested)
			} else if nextItem.Tok == parser.IDENT && nextItem.Lit == "z" {
				p.p.Unscan()
				schema, err := p.ParseZodExpr()
				if err != nil {
					return nil, err
				}
				nested.Children = append(nested.Children, &SchemaProperty{Name: propName, Schema: schema})

				if p.p.Match(parser.COMMA) {
					p.p.Scan()
				}
			} else {
				p.p.Unscan()
				p.SkipUntil(parser.COMMA, parser.RBRACE)
			}
		}
	}

	// Skip comma after closing brace
	if p.p.Match(parser.COMMA) {
		p.p.Scan()
	}

	return nested, nil
}

func (p *Parser) ParseTypeAliasDecl(exported bool) (Node, error) {
	// Parse name
	nameItem, err := p.p.Expect(parser.IDENT)
	if err != nil {
		return nil, err
	}
	name := nameItem.Lit

	// Expect =
	if _, err := p.p.Expect(parser.EQUALS); err != nil {
		return nil, err
	}

	// Check for z.infer<typeof X>
	item := p.p.ScanIgnoreWhitespace()

	if item.Tok == parser.IDENT && item.Lit == "z" {
		// Check for z.infer
		if p.p.Match(parser.DOT) {
			p.p.Scan()
			methodItem := p.p.ScanIgnoreWhitespace()
			if methodItem.Tok == parser.IDENT && methodItem.Lit == "infer" {
				// Parse <typeof schemaRef>
				if _, err := p.p.Expect(parser.LESS); err != nil {
					return nil, err
				}

				typeofItem := p.p.ScanIgnoreWhitespace()
				if typeofItem.Tok != TYPEOF {
					return nil, parser.NewParseError(typeofItem.Pos, "expected 'typeof'")
				}

				refItem, err := p.p.Expect(parser.IDENT)
				if err != nil {
					return nil, err
				}

				if _, err := p.p.Expect(parser.GREATER); err != nil {
					return nil, err
				}

				// Skip semicolon
				if p.p.Match(parser.SEMICOLON) {
					p.p.Scan()
				}

				return &InferredType{Name: name, SchemaRef: refItem.Lit, Export: exported}, nil
			}
		}
	}

	// Regular type alias - collect the type expression as string
	p.p.Unscan()
	typeStr := p.CollectTypeExpression()

	// Skip semicolon
	if p.p.Match(parser.SEMICOLON) {
		p.p.Scan()
	}

	return &TypeAlias{Name: name, Type: typeStr, Export: exported}, nil
}

func (p *Parser) CollectTypeExpression() string {
	var parts []string
	depth := 0

	for {
		item := p.p.Scan()

		if item.Tok == parser.EOF {
			break
		}

		if item.Tok == parser.SEMICOLON && depth == 0 {
			p.p.Unscan()
			break
		}

		if item.Tok == parser.LESS || item.Tok == parser.LBRACKET || item.Tok == parser.LPAREN {
			depth++
		}
		if item.Tok == parser.GREATER || item.Tok == parser.RBRACKET || item.Tok == parser.RPAREN {
			depth--
		}

		parts = append(parts, item.Lit)
	}

	return strings.Join(parts, "")
}

func (p *Parser) ParseTSNamespaceDecl() (*TSNamespace, error) {
	ns := &TSNamespace{}

	// Parse name
	nameItem, err := p.p.Expect(parser.IDENT)
	if err != nil {
		return nil, err
	}
	ns.Name = nameItem.Lit

	// Expect {
	if _, err := p.p.Expect(parser.LBRACE); err != nil {
		return nil, err
	}

	// Parse children
	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.RBRACE {
			break
		}

		if item.Tok == parser.EOF {
			return nil, parser.NewParseError(item.Pos, "unexpected EOF in namespace")
		}

		if item.Tok == parser.COMMENT {
			continue
		}

		if item.Tok == EXPORT {
			nextItem := p.p.ScanIgnoreWhitespace()
			if nextItem.Tok == INTERFACE {
				iface, err := p.ParseInterfaceDecl()
				if err != nil {
					return nil, err
				}
				ns.Children = append(ns.Children, iface)
			} else if nextItem.Tok == TYPE {
				alias, err := p.ParseTypeAliasDecl(true)
				if err != nil {
					return nil, err
				}
				ns.Children = append(ns.Children, alias)
			} else {
				p.p.Unscan()
				p.SkipToNextExport()
			}
		}
	}

	return ns, nil
}

func (p *Parser) ParseInterfaceDecl() (*Interface, error) {
	iface := &Interface{}

	// Parse name
	nameItem, err := p.p.Expect(parser.IDENT)
	if err != nil {
		return nil, err
	}
	iface.Name = nameItem.Lit

	// Expect {
	if _, err := p.p.Expect(parser.LBRACE); err != nil {
		return nil, err
	}

	// Parse properties
	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.RBRACE {
			break
		}

		if item.Tok == parser.EOF {
			return nil, parser.NewParseError(item.Pos, "unexpected EOF in interface")
		}

		if item.Tok == parser.COMMENT {
			continue
		}

		p.p.Unscan()
		prop, err := p.ParseInterfacePropertyDecl()
		if err != nil {
			return nil, err
		}
		iface.Properties = append(iface.Properties, *prop)
	}

	return iface, nil
}

func (p *Parser) ParseInterfacePropertyDecl() (*InterfaceProperty, error) {
	prop := &InterfaceProperty{}

	item := p.p.ScanIgnoreWhitespace()

	// Check for readonly
	if item.Tok == READONLY {
		prop.ReadOnly = true
		item = p.p.ScanIgnoreWhitespace()
	}

	// Property name
	if item.Tok != parser.IDENT {
		return nil, parser.NewParseError(item.Pos, "expected property name")
	}
	prop.Name = item.Lit

	// Check for optional ?
	item = p.p.ScanIgnoreWhitespace()
	if item.Tok == parser.QUESTION {
		prop.Optional = true
		item = p.p.ScanIgnoreWhitespace()
	}

	// Expect :
	if item.Tok != parser.COLON {
		return nil, parser.NewParseError(item.Pos, "expected ':'")
	}

	// Parse type: IndexType["IndexKey"]
	typeItem, err := p.p.Expect(parser.IDENT)
	if err != nil {
		return nil, err
	}
	prop.IndexType = typeItem.Lit

	// Check for indexed access
	if p.p.Match(parser.LBRACKET) {
		p.p.Scan()
		keyItem := p.p.ScanIgnoreWhitespace()
		if keyItem.Tok == parser.STRING {
			prop.IndexKey = keyItem.Lit
		}
		p.p.Expect(parser.RBRACKET)
	}

	// Skip semicolon
	if p.p.Match(parser.SEMICOLON) {
		p.p.Scan()
	}

	return prop, nil
}

// --- Zod Expression Parsing ---

func (p *Parser) ParseZodExpr() (ZodExpr, error) {
	// Expect "z"
	item := p.p.ScanIgnoreWhitespace()
	if item.Tok != parser.IDENT || item.Lit != "z" {
		return nil, parser.NewParseError(item.Pos, "expected 'z'")
	}

	// Expect .
	if _, err := p.p.Expect(parser.DOT); err != nil {
		return nil, err
	}

	// Get method name
	methodItem, err := p.p.Expect(parser.IDENT)
	if err != nil {
		return nil, err
	}

	var expr ZodExpr

	switch methodItem.Lit {
	case "object":
		expr, err = p.ParseZObjectExpr()
	case "array":
		expr, err = p.ParseZArrayExpr()
	case "string":
		expr, err = p.ParseZStringExpr()
	case "number":
		expr, err = p.ParseZNumberExpr()
	case "boolean":
		if _, err := p.p.Expect(parser.LPAREN); err != nil {
			return nil, err
		}
		if _, err := p.p.Expect(parser.RPAREN); err != nil {
			return nil, err
		}
		expr = ZBoolean{}
	case "enum":
		expr, err = p.ParseZEnumExpr()
	case "union":
		expr, err = p.ParseZUnionExpr()
	case "lazy":
		expr, err = p.ParseZLazyExpr()
	case "record":
		expr, err = p.ParseZRecordExpr()
	case "literal":
		expr, err = p.ParseZLiteralExpr()
	case "null":
		if _, err := p.p.Expect(parser.LPAREN); err != nil {
			return nil, err
		}
		if _, err := p.p.Expect(parser.RPAREN); err != nil {
			return nil, err
		}
		expr = ZNull{}
	case "unknown":
		if _, err := p.p.Expect(parser.LPAREN); err != nil {
			return nil, err
		}
		if _, err := p.p.Expect(parser.RPAREN); err != nil {
			return nil, err
		}
		expr = ZUnknown{}
	default:
		return nil, parser.NewParseError(methodItem.Pos, "unknown zod method: %s", methodItem.Lit)
	}

	if err != nil {
		return nil, err
	}

	// Parse chained modifiers and .and()
	return p.ParseModifiers(expr)
}

func (p *Parser) ParseZObjectExpr() (ZodExpr, error) {
	obj := ZObject{}

	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return nil, err
	}

	if _, err := p.p.Expect(parser.LBRACE); err != nil {
		return nil, err
	}

	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.RBRACE {
			break
		}

		if item.Tok == parser.EOF {
			return nil, parser.NewParseError(item.Pos, "unexpected EOF in z.object")
		}

		if item.Tok == parser.COMMENT {
			continue
		}

		if item.Tok == parser.IDENT {
			propName := item.Lit

			if _, err := p.p.Expect(parser.COLON); err != nil {
				return nil, err
			}

			schema, err := p.ParseZodExpr()
			if err != nil {
				return nil, err
			}

			obj.Properties = append(obj.Properties, ZProperty{Name: propName, Schema: schema})

			// Skip comma
			if p.p.Match(parser.COMMA) {
				p.p.Scan()
			}
		}
	}

	if _, err := p.p.Expect(parser.RPAREN); err != nil {
		return nil, err
	}

	return obj, nil
}

func (p *Parser) ParseZArrayExpr() (ZodExpr, error) {
	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return nil, err
	}

	elem, err := p.ParseZodExpr()
	if err != nil {
		return nil, err
	}

	if _, err := p.p.Expect(parser.RPAREN); err != nil {
		return nil, err
	}

	return ZArray{Element: elem}, nil
}

func (p *Parser) ParseZStringExpr() (ZodExpr, error) {
	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return nil, err
	}
	if _, err := p.p.Expect(parser.RPAREN); err != nil {
		return nil, err
	}
	return ZString{}, nil
}

func (p *Parser) ParseZNumberExpr() (ZodExpr, error) {
	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return nil, err
	}
	if _, err := p.p.Expect(parser.RPAREN); err != nil {
		return nil, err
	}
	return ZNumber{}, nil
}

func (p *Parser) ParseZEnumExpr() (ZodExpr, error) {
	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return nil, err
	}
	if _, err := p.p.Expect(parser.LBRACKET); err != nil {
		return nil, err
	}

	var values []string
	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.RBRACKET {
			break
		}

		if item.Tok == parser.STRING {
			values = append(values, item.Lit)
		}

		if p.p.Match(parser.COMMA) {
			p.p.Scan()
		}
	}

	if _, err := p.p.Expect(parser.RPAREN); err != nil {
		return nil, err
	}

	return ZEnum{Values: values}, nil
}

func (p *Parser) ParseZUnionExpr() (ZodExpr, error) {
	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return nil, err
	}
	if _, err := p.p.Expect(parser.LBRACKET); err != nil {
		return nil, err
	}

	var schemas []ZodExpr
	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.RBRACKET {
			break
		}

		p.p.Unscan()
		schema, err := p.ParseZodExpr()
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)

		if p.p.Match(parser.COMMA) {
			p.p.Scan()
		}
	}

	if _, err := p.p.Expect(parser.RPAREN); err != nil {
		return nil, err
	}

	return ZUnion{Schemas: schemas}, nil
}

func (p *Parser) ParseZLazyExpr() (ZodExpr, error) {
	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return nil, err
	}

	// Skip arrow function: () =>
	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return nil, err
	}
	if _, err := p.p.Expect(parser.RPAREN); err != nil {
		return nil, err
	}

	// Skip =>
	item := p.p.ScanIgnoreWhitespace()
	if item.Tok != parser.EQUALS {
		return nil, parser.NewParseError(item.Pos, "expected '=>'")
	}
	item = p.p.Scan()
	if item.Tok != parser.GREATER {
		return nil, parser.NewParseError(item.Pos, "expected '=>'")
	}

	// Get ref
	refItem, err := p.p.Expect(parser.IDENT)
	if err != nil {
		return nil, err
	}

	if _, err := p.p.Expect(parser.RPAREN); err != nil {
		return nil, err
	}

	return ZLazy{Ref: refItem.Lit}, nil
}

func (p *Parser) ParseZRecordExpr() (ZodExpr, error) {
	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return nil, err
	}

	key, err := p.ParseZodExpr()
	if err != nil {
		return nil, err
	}

	if _, err := p.p.Expect(parser.COMMA); err != nil {
		return nil, err
	}

	value, err := p.ParseZodExpr()
	if err != nil {
		return nil, err
	}

	if _, err := p.p.Expect(parser.RPAREN); err != nil {
		return nil, err
	}

	return ZRecord{Key: key, Value: value}, nil
}

func (p *Parser) ParseZLiteralExpr() (ZodExpr, error) {
	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return nil, err
	}

	item := p.p.ScanIgnoreWhitespace()
	var value any

	switch item.Tok {
	case parser.STRING:
		value = item.Lit
	case parser.NUMBER:
		if strings.Contains(item.Lit, ".") {
			value, _ = strconv.ParseFloat(item.Lit, 64)
		} else {
			v, _ := strconv.Atoi(item.Lit)
			value = float64(v)
		}
	case parser.IDENT:
		if item.Lit == "true" {
			value = true
		} else if item.Lit == "false" {
			value = false
		} else if item.Lit == "null" {
			value = nil
		}
	}

	if _, err := p.p.Expect(parser.RPAREN); err != nil {
		return nil, err
	}

	return ZLiteral{Value: value}, nil
}

func (p *Parser) ParseModifiers(expr ZodExpr) (ZodExpr, error) {
	var modifiers []Modifier

	for p.p.Match(parser.DOT) {
		p.p.Scan()

		methodItem := p.p.ScanIgnoreWhitespace()
		if methodItem.Tok != parser.IDENT {
			p.p.Unscan()
			break
		}

		switch methodItem.Lit {
		case "optional":
			p.ConsumeParens()
			modifiers = append(modifiers, Optional{})
		case "nullable":
			p.ConsumeParens()
			modifiers = append(modifiers, Nullable{})
		case "email":
			p.ConsumeParens()
			modifiers = append(modifiers, Email{})
		case "url":
			p.ConsumeParens()
			modifiers = append(modifiers, Url{})
		case "uuid":
			p.ConsumeParens()
			modifiers = append(modifiers, Uuid{})
		case "datetime":
			p.ConsumeParens()
			modifiers = append(modifiers, Datetime{})
		case "date":
			p.ConsumeParens()
			modifiers = append(modifiers, Date{})
		case "int":
			p.ConsumeParens()
			modifiers = append(modifiers, Int{})
		case "min":
			v := p.ParseNumericArg()
			modifiers = append(modifiers, Min{V: int(v)})
		case "max":
			v := p.ParseNumericArg()
			modifiers = append(modifiers, Max{V: int(v)})
		case "regex":
			pattern := p.ParseRegexArg()
			modifiers = append(modifiers, Regex{Pattern: pattern})
		case "and":
			// Intersection
			p.p.Expect(parser.LPAREN)
			other, err := p.ParseZodExpr()
			if err != nil {
				return nil, err
			}
			p.p.Expect(parser.RPAREN)

			// Wrap current expression in intersection
			if inter, ok := expr.(ZIntersection); ok {
				inter.Schemas = append(inter.Schemas, other)
				expr = inter
			} else {
				expr = ZIntersection{Schemas: []ZodExpr{expr, other}}
			}
		default:
			// Unknown modifier, skip it
			p.ConsumeParens()
		}
	}

	if len(modifiers) > 0 {
		return ZWithModifiers{Expr: expr, Modifiers: modifiers}, nil
	}
	return expr, nil
}

func (p *Parser) ConsumeParens() {
	if p.p.Match(parser.LPAREN) {
		p.p.Scan()
		depth := 1
		for depth > 0 {
			item := p.p.Scan()
			if item.Tok == parser.LPAREN {
				depth++
			} else if item.Tok == parser.RPAREN {
				depth--
			} else if item.Tok == parser.EOF {
				break
			}
		}
	}
}

func (p *Parser) ParseNumericArg() float64 {
	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return 0
	}
	item := p.p.ScanIgnoreWhitespace()
	var v float64
	if item.Tok == parser.NUMBER {
		v, _ = strconv.ParseFloat(item.Lit, 64)
	}
	p.p.Expect(parser.RPAREN)
	return v
}

func (p *Parser) ParseRegexArg() string {
	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return ""
	}
	// Regex is like /pattern/ - need to handle that regex can contain )
	// Scan until we see / followed by ), as the pattern is /.../)
	pattern := p.p.ScanRegexUntilClose()
	// Extract pattern from /.../ format
	pattern = strings.TrimSpace(pattern)
	if len(pattern) >= 2 && pattern[0] == '/' && pattern[len(pattern)-1] == '/' {
		pattern = pattern[1 : len(pattern)-1]
	}
	return pattern
}
