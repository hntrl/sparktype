package typescript

import (
	"github.com/hntrl/sparktype/internal/parser"
)

// TypeScript-specific tokens
const (
	EXPORT parser.Token = parser.KEYWORD_START + iota
	INTERFACE
	TYPE
	NAMESPACE
	READONLY
	ENUM
	CONST
)

var tsKeywords = map[string]parser.Token{
	"export":    EXPORT,
	"interface": INTERFACE,
	"type":      TYPE,
	"namespace": NAMESPACE,
	"readonly":  READONLY,
	"enum":      ENUM,
	"const":     CONST,
}

func tsKeywordLookup(lit string) parser.Token {
	if tok, ok := tsKeywords[lit]; ok {
		return tok
	}
	return parser.ILLEGAL
}

// Parser parses TypeScript source into AST nodes
type Parser struct {
	p *parser.Parser
}

// NewParser creates a new TypeScript parser
func NewParser(input string) *Parser {
	p := parser.NewParser(input)
	p.SetKeywordLookup(tsKeywordLookup)
	return &Parser{p: p}
}

// ParseFile parses a TypeScript file from a Parser
func (p *Parser) ParseFile() (*File, error) {
	file := &File{}

	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.EOF {
			break
		}

		// Skip comments but capture JSDoc
		if item.Tok == parser.COMMENT {
			continue
		}

		// Look for export declarations
		if item.Tok == EXPORT {
			node, err := p.ParseExport()
			if err != nil {
				return nil, err
			}
			if node != nil {
				file.Nodes = append(file.Nodes, node)
			}
		}
	}

	return file, nil
}

func (p *Parser) ParseExport() (Node, error) {
	// Check what follows export
	item := p.p.ScanIgnoreWhitespace()

	switch item.Tok {
	case INTERFACE:
		return p.ParseInterface()
	case TYPE:
		return p.ParseTypeAlias()
	case NAMESPACE:
		return p.ParseNamespace()
	case ENUM:
		return p.ParseEnum()
	case CONST:
		// Could be a const enum - skip for now
		p.SkipToNextExport()
		return nil, nil
	default:
		// Unknown export, skip to next
		p.p.Unscan()
		p.SkipToNextExport()
		return nil, nil
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

func (p *Parser) ParseNamespace() (*Namespace, error) {
	ns := &Namespace{}

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

	// Parse children until }
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
			node, err := p.ParseExport()
			if err != nil {
				return nil, err
			}
			if node != nil {
				ns.Children = append(ns.Children, node)
			}
		}
	}

	return ns, nil
}

func (p *Parser) ParseInterface() (*Interface, error) {
	iface := &Interface{ExportStyle: "interface"}

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

	// Parse properties until }
	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.RBRACE {
			break
		}

		if item.Tok == parser.EOF {
			return nil, parser.NewParseError(item.Pos, "unexpected EOF in interface")
		}

		// Skip comments but could capture JSDoc here
		if item.Tok == parser.COMMENT {
			continue
		}

		p.p.Unscan()
		prop, err := p.ParseProperty()
		if err != nil {
			return nil, err
		}
		iface.Properties = append(iface.Properties, *prop)
	}

	return iface, nil
}

func (p *Parser) ParseEnum() (*Enum, error) {
	enum := &Enum{}

	// Parse name
	nameItem, err := p.p.Expect(parser.IDENT)
	if err != nil {
		return nil, err
	}
	enum.Name = nameItem.Lit

	// Expect {
	if _, err := p.p.Expect(parser.LBRACE); err != nil {
		return nil, err
	}

	// Parse members until }
	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.RBRACE {
			break
		}

		if item.Tok == parser.EOF {
			return nil, parser.NewParseError(item.Pos, "unexpected EOF in enum")
		}

		if item.Tok == parser.COMMENT {
			continue
		}

		if item.Tok == parser.IDENT {
			member := EnumMember{Key: item.Lit}

			// Check for = value
			if p.p.Match(parser.EQUALS) {
				p.p.Scan()
				valueItem := p.p.ScanIgnoreWhitespace()
				if valueItem.Tok == parser.STRING {
					member.Value = valueItem.Lit
				} else if valueItem.Tok == parser.NUMBER {
					member.Value = valueItem.Lit
				}
			}

			// Skip comma
			if p.p.Match(parser.COMMA) {
				p.p.Scan()
			}

			enum.Members = append(enum.Members, member)
		}
	}

	return enum, nil
}

func (p *Parser) ParseTypeAlias() (Node, error) {
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

	// Check if it's an object type (like an interface)
	item := p.p.ScanIgnoreWhitespace()
	if item.Tok == parser.LBRACE {
		// It's an inline object type - parse as interface with type style
		iface := &Interface{Name: name, ExportStyle: "type"}

		for {
			item := p.p.ScanIgnoreWhitespace()
			if item.Tok == parser.RBRACE {
				break
			}
			if item.Tok == parser.EOF {
				return nil, parser.NewParseError(item.Pos, "unexpected EOF in type")
			}
			if item.Tok == parser.COMMENT {
				continue
			}

			p.p.Unscan()
			prop, err := p.ParseProperty()
			if err != nil {
				return nil, err
			}
			iface.Properties = append(iface.Properties, *prop)
		}

		// Consume optional semicolon
		if p.p.Match(parser.SEMICOLON) {
			p.p.Scan()
		}

		return iface, nil
	}

	// It's a regular type alias
	p.p.Unscan()
	typeExpr, err := p.ParseTypeExpr()
	if err != nil {
		return nil, err
	}

	// Consume optional semicolon
	if p.p.Match(parser.SEMICOLON) {
		p.p.Scan()
	}

	return &TypeAlias{Name: name, Type: typeExpr}, nil
}

func (p *Parser) ParseProperty() (*Property, error) {
	prop := &Property{}

	item := p.p.ScanIgnoreWhitespace()

	// Check for readonly
	if item.Tok == READONLY {
		prop.ReadOnly = true
		item = p.p.ScanIgnoreWhitespace()
	}

	// Property name
	if item.Tok != parser.IDENT && item.Tok != parser.STRING {
		return nil, parser.NewParseError(item.Pos, "expected property name, got %s", item.Tok)
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
		return nil, parser.NewParseError(item.Pos, "expected ':', got %s", item.Tok)
	}

	// Parse type
	typeExpr, err := p.ParseTypeExpr()
	if err != nil {
		return nil, err
	}

	// Check if the type is a union ending with null (T | null) and extract Nullable flag
	if union, ok := typeExpr.(UnionType); ok {
		if len(union.Types) >= 2 {
			lastType := union.Types[len(union.Types)-1]
			if _, isNull := lastType.(NullType); isNull {
				prop.Nullable = true
				// Remove null from the union
				remaining := union.Types[:len(union.Types)-1]
				if len(remaining) == 1 {
					// Single type remaining, use it directly
					typeExpr = remaining[0]
				} else {
					// Multiple types remaining, keep as union
					typeExpr = UnionType{Types: remaining}
				}
			}
		}
	}

	prop.Type = typeExpr

	// Consume semicolon or comma
	item = p.p.ScanIgnoreWhitespace()
	if item.Tok != parser.SEMICOLON && item.Tok != parser.COMMA && item.Tok != parser.RBRACE {
		// Some valid token we should unscan
		p.p.Unscan()
	} else if item.Tok == parser.RBRACE {
		p.p.Unscan()
	}

	return prop, nil
}

func (p *Parser) ParseTypeExpr() (TypeExpr, error) {
	// Parse primary type
	primary, err := p.ParsePrimaryType()
	if err != nil {
		return nil, err
	}

	// Check for union or intersection
	item := p.p.ScanIgnoreWhitespace()

	if item.Tok == parser.PIPE {
		// Union type
		types := []TypeExpr{primary}
		for {
			nextType, err := p.ParsePrimaryType()
			if err != nil {
				return nil, err
			}
			types = append(types, nextType)

			item = p.p.ScanIgnoreWhitespace()
			if item.Tok != parser.PIPE {
				p.p.Unscan()
				break
			}
		}
		return UnionType{Types: types}, nil
	}

	if item.Tok == parser.AMPERSAND {
		// Intersection type
		types := []TypeExpr{primary}
		for {
			nextType, err := p.ParsePrimaryType()
			if err != nil {
				return nil, err
			}
			types = append(types, nextType)

			item = p.p.ScanIgnoreWhitespace()
			if item.Tok != parser.AMPERSAND {
				p.p.Unscan()
				break
			}
		}
		return IntersectionType{Types: types}, nil
	}

	p.p.Unscan()
	return primary, nil
}

func (p *Parser) ParsePrimaryType() (TypeExpr, error) {
	item := p.p.ScanIgnoreWhitespace()

	switch item.Tok {
	case parser.IDENT:
		return p.ParseIdentType(item.Lit)

	case parser.STRING:
		return LiteralType{Value: item.Lit}, nil

	case parser.NUMBER:
		return LiteralType{Value: item.Lit}, nil

	case parser.LBRACE:
		// Inline object type
		return p.ParseInlineObjectType()

	case parser.LPAREN:
		// Parenthesized type
		inner, err := p.ParseTypeExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.p.Expect(parser.RPAREN); err != nil {
			return nil, err
		}
		return inner, nil

	default:
		return nil, parser.NewParseError(item.Pos, "unexpected token in type expression: %s (%q)", item.Tok, item.Lit)
	}
}

func (p *Parser) ParseIdentType(name string) (TypeExpr, error) {
	// Check for built-in types
	switch name {
	case "string":
		return StringType{}, nil
	case "number":
		return NumberType{}, nil
	case "boolean":
		return BooleanType{}, nil
	case "null":
		return NullType{}, nil
	case "unknown":
		return UnknownType{}, nil
	}

	// Check for namespace-qualified types like Namespace.Type
	fullName := name
	for p.p.Match(parser.DOT) {
		p.p.Scan() // consume the dot
		nextItem := p.p.ScanIgnoreWhitespace()
		if nextItem.Tok != parser.IDENT {
			return nil, parser.NewParseError(nextItem.Pos, "expected identifier after '.', got %s", nextItem.Tok)
		}
		fullName = fullName + "." + nextItem.Lit
	}
	name = fullName

	// Check for generic types like Array<T> or Record<K, V>
	if p.p.Match(parser.LESS) {
		p.p.Scan()

		switch name {
		case "Array":
			elem, err := p.ParseTypeExpr()
			if err != nil {
				return nil, err
			}
			if _, err := p.p.Expect(parser.GREATER); err != nil {
				return nil, err
			}
			return ArrayType{Element: elem}, nil

		case "Record":
			key, err := p.ParseTypeExpr()
			if err != nil {
				return nil, err
			}
			if _, err := p.p.Expect(parser.COMMA); err != nil {
				return nil, err
			}
			value, err := p.ParseTypeExpr()
			if err != nil {
				return nil, err
			}
			if _, err := p.p.Expect(parser.GREATER); err != nil {
				return nil, err
			}
			return RecordType{Key: key, Value: value}, nil

		default:
			// Generic type - for now just skip to >
			depth := 1
			for depth > 0 {
				item := p.p.Scan()
				if item.Tok == parser.LESS {
					depth++
				} else if item.Tok == parser.GREATER {
					depth--
				} else if item.Tok == parser.EOF {
					break
				}
			}
			return ReferenceType{Name: name}, nil
		}
	}

	// Check for indexed access type: Type["key"]
	if p.p.Match(parser.LBRACKET) {
		p.p.Scan()
		indexItem := p.p.ScanIgnoreWhitespace()
		if indexItem.Tok == parser.STRING {
			if _, err := p.p.Expect(parser.RBRACKET); err != nil {
				return nil, err
			}
			return IndexedType{Base: name, Index: indexItem.Lit}, nil
		}
		// Not a string index, skip it
		p.p.SkipUntil(parser.RBRACKET)
	}

	return ReferenceType{Name: name}, nil
}

func (p *Parser) ParseInlineObjectType() (TypeExpr, error) {
	obj := ObjectType{}

	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.RBRACE {
			break
		}

		if item.Tok == parser.EOF {
			return nil, parser.NewParseError(item.Pos, "unexpected EOF in object type")
		}

		if item.Tok == parser.COMMENT {
			continue
		}

		p.p.Unscan()
		prop, err := p.ParseProperty()
		if err != nil {
			return nil, err
		}
		obj.Properties = append(obj.Properties, *prop)
	}

	return obj, nil
}
