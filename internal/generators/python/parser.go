package python

import (
	"github.com/hntrl/sparktype/internal/parser"
)

// Python-specific tokens
const (
	CLASS parser.Token = parser.KEYWORD_START + iota
	DEF
	PASS
	FROM
	IMPORT
)

var pyKeywords = map[string]parser.Token{
	"class":  CLASS,
	"def":    DEF,
	"pass":   PASS,
	"from":   FROM,
	"import": IMPORT,
}

func pyKeywordLookup(lit string) parser.Token {
	if tok, ok := pyKeywords[lit]; ok {
		return tok
	}
	return parser.ILLEGAL
}

// Parser parses Python source into AST nodes
type Parser struct {
	p *parser.Parser
}

// NewParser creates a new Python parser
func NewParser(input string) *Parser {
	p := parser.NewParser(input)
	p.SetKeywordLookup(pyKeywordLookup)
	return &Parser{p: p}
}

// ParseFile parses a Python file from a Parser
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

		if item.Tok == CLASS {
			node, err := p.ParseClass()
			if err != nil {
				return nil, err
			}
			if node != nil {
				file.Classes = append(file.Classes, node)
			}
		} else if item.Tok == parser.IDENT {
			// Could be a type alias: Name = Type
			name := item.Lit
			if p.p.Match(parser.EQUALS) {
				p.p.Scan()
				typeExpr, err := p.ParseTypeExpr()
				if err != nil {
					return nil, err
				}
				file.Classes = append(file.Classes, &TypeAlias{Name: name, Type: typeExpr})
			}
		}
	}

	return file, nil
}

func (p *Parser) ParseClass() (Node, error) {
	// Parse class name
	nameItem, err := p.p.Expect(parser.IDENT)
	if err != nil {
		return nil, err
	}
	name := nameItem.Lit

	// Parse base classes: (TypedDict) or (str, Enum)
	if _, err := p.p.Expect(parser.LPAREN); err != nil {
		return nil, err
	}

	// Get base class(es)
	var bases []string
	for {
		item := p.p.ScanIgnoreWhitespace()
		if item.Tok == parser.RPAREN {
			break
		}
		if item.Tok == parser.IDENT {
			bases = append(bases, item.Lit)
		}
		if item.Tok == parser.COMMA {
			continue
		}
		// Skip total=False and similar kwargs
		if item.Tok == parser.EQUALS {
			p.p.ScanIgnoreWhitespace() // consume value
		}
	}

	// Expect :
	if _, err := p.p.Expect(parser.COLON); err != nil {
		return nil, err
	}

	// Determine class type based on bases
	isTypedDict := false
	isEnum := false
	for _, base := range bases {
		if base == "TypedDict" {
			isTypedDict = true
		}
		if base == "Enum" {
			isEnum = true
		}
	}

	if isTypedDict {
		return p.ParseTypedDictBody(name)
	} else if isEnum {
		return p.ParseEnumBody(name)
	}

	// Unknown class type, skip body
	p.SkipClassBody()
	return nil, nil
}

func (p *Parser) SkipClassBody() {
	depth := 1
	for depth > 0 {
		item := p.p.Scan()
		if item.Tok == parser.EOF {
			return
		}
		// Python indentation-based, look for next class or top-level statement
		if item.Tok == CLASS || item.Tok == FROM || item.Tok == IMPORT {
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

func (p *Parser) ParseTypedDictBody(name string) (*TypedDict, error) {
	td := &TypedDict{Name: name, Total: true}

	// Skip newline after :
	p.p.ScanIgnore(parser.COMMENT)

	// Check for docstring
	item := p.p.ScanIgnoreWhitespace()
	if item.Tok == parser.STRING {
		td.Docstring = Docstring(item.Lit)
		item = p.p.ScanIgnoreWhitespace()
	}

	// Parse fields
	for {
		if item.Tok == parser.EOF {
			break
		}

		// Check for dedent (new class or top-level statement)
		if item.Tok == CLASS || item.Tok == FROM || item.Tok == IMPORT {
			p.p.Unscan()
			break
		}

		if item.Tok == PASS {
			item = p.p.ScanIgnoreWhitespace()
			continue
		}

		if item.Tok == parser.COMMENT {
			item = p.p.ScanIgnoreWhitespace()
			continue
		}

		if item.Tok == parser.NEWLINE {
			item = p.p.ScanIgnoreWhitespace()
			continue
		}

		if item.Tok == parser.IDENT {
			fieldName := item.Lit

			// Check if this looks like a type annotation (name: Type)
			next := p.p.ScanIgnoreWhitespace()
			if next.Tok == parser.COLON {
				// It's a field
				typeExpr, err := p.ParseTypeExpr()
				if err != nil {
					return nil, err
				}

				field := Field{Name: fieldName, Type: typeExpr}

				// Check for docstring after field
				docItem := p.p.ScanIgnoreWhitespace()
				if docItem.Tok == parser.STRING {
					field.Docstring = Docstring(docItem.Lit)
				} else {
					p.p.Unscan()
				}

				td.Fields = append(td.Fields, field)
			} else if next.Tok == parser.EQUALS {
				// This is an assignment, likely an enum member or at class level
				// which means we've moved past the TypedDict fields
				p.p.Unscan()
				p.p.Unscan()
				break
			} else {
				p.p.Unscan()
			}
		}

		item = p.p.ScanIgnoreWhitespace()
	}

	return td, nil
}

func (p *Parser) ParseEnumBody(name string) (*EnumClass, error) {
	enum := &EnumClass{Name: name}

	// Skip newline after :
	p.p.ScanIgnore(parser.COMMENT)

	// Check for docstring
	item := p.p.ScanIgnoreWhitespace()
	if item.Tok == parser.STRING {
		enum.Docstring = Docstring(item.Lit)
		item = p.p.ScanIgnoreWhitespace()
	}

	// Parse members
	for {
		if item.Tok == parser.EOF {
			break
		}

		// Check for dedent
		if item.Tok == CLASS || item.Tok == FROM || item.Tok == IMPORT {
			p.p.Unscan()
			break
		}

		if item.Tok == parser.NEWLINE || item.Tok == parser.COMMENT {
			item = p.p.ScanIgnoreWhitespace()
			continue
		}

		if item.Tok == parser.IDENT {
			memberKey := item.Lit

			// Expect =
			if p.p.Match(parser.EQUALS) {
				p.p.Scan()
				valueItem := p.p.ScanIgnoreWhitespace()
				if valueItem.Tok == parser.STRING {
					enum.Members = append(enum.Members, EnumMember{Key: memberKey, Value: valueItem.Lit})
				}
			}
		}

		item = p.p.ScanIgnoreWhitespace()
	}

	return enum, nil
}

// --- Type Expression Parsing ---

func (p *Parser) ParseTypeExpr() (PyTypeExpr, error) {
	// Parse primary type
	primary, err := p.ParsePrimaryType()
	if err != nil {
		return nil, err
	}

	// Check for union with |
	item := p.p.ScanIgnoreWhitespace()
	if item.Tok == parser.PIPE {
		types := []PyTypeExpr{primary}
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
		// Check for T | None pattern - convert to OptionalType
		if len(types) == 2 {
			if _, isNone := types[1].(NoneType); isNone {
				return OptionalType{Inner: types[0]}, nil
			}
		}
		return UnionType{Types: types}, nil
	}

	p.p.Unscan()
	return primary, nil
}

func (p *Parser) ParsePrimaryType() (PyTypeExpr, error) {
	item := p.p.ScanIgnoreWhitespace()

	switch item.Tok {
	case parser.IDENT:
		return p.ParseIdentType(item.Lit)

	case parser.STRING:
		// Forward reference: "TypeName"
		return ReferenceType{Name: item.Lit}, nil

	default:
		p.p.Unscan()
		return AnyType{}, nil
	}
}

func (p *Parser) ParseIdentType(name string) (PyTypeExpr, error) {
	// Check for built-in types
	switch name {
	case "str":
		return StrType{}, nil
	case "int":
		return IntType{}, nil
	case "float":
		return FloatType{}, nil
	case "bool":
		return BoolType{}, nil
	case "None":
		return NoneType{}, nil
	case "Any":
		return AnyType{}, nil
	}

	// Check for generic types with []
	if p.p.Match(parser.LBRACKET) {
		p.p.Scan()

		switch name {
		case "List":
			elem, err := p.ParseTypeExpr()
			if err != nil {
				return nil, err
			}
			if _, err := p.p.Expect(parser.RBRACKET); err != nil {
				return nil, err
			}
			return ListType{Element: elem}, nil

		case "Dict":
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
			if _, err := p.p.Expect(parser.RBRACKET); err != nil {
				return nil, err
			}
			return DictType{Key: key, Value: value}, nil

		case "Union":
			var types []PyTypeExpr
			for {
				t, err := p.ParseTypeExpr()
				if err != nil {
					return nil, err
				}
				types = append(types, t)

				item := p.p.ScanIgnoreWhitespace()
				if item.Tok == parser.RBRACKET {
					break
				}
				if item.Tok != parser.COMMA {
					p.p.Unscan()
					break
				}
			}
			// Check for Union[T, None] pattern - convert to OptionalType
			if len(types) == 2 {
				if _, isNone := types[1].(NoneType); isNone {
					return OptionalType{Inner: types[0]}, nil
				}
			}
			return UnionType{Types: types}, nil

		case "Optional":
			inner, err := p.ParseTypeExpr()
			if err != nil {
				return nil, err
			}
			if _, err := p.p.Expect(parser.RBRACKET); err != nil {
				return nil, err
			}
			return OptionalType{Inner: inner}, nil

		case "NotRequired":
			inner, err := p.ParseTypeExpr()
			if err != nil {
				return nil, err
			}
			if _, err := p.p.Expect(parser.RBRACKET); err != nil {
				return nil, err
			}
			return NotRequiredType{Inner: inner}, nil

		case "Literal":
			// Parse literal values
			var values []any
			for {
				item := p.p.ScanIgnoreWhitespace()
				if item.Tok == parser.RBRACKET {
					break
				}
				if item.Tok == parser.STRING {
					values = append(values, item.Lit)
				} else if item.Tok == parser.NUMBER {
					values = append(values, item.Lit)
				} else if item.Tok == parser.IDENT {
					if item.Lit == "True" {
						values = append(values, true)
					} else if item.Lit == "False" {
						values = append(values, false)
					} else if item.Lit == "None" {
						values = append(values, nil)
					}
				}

				if p.p.Match(parser.COMMA) {
					p.p.Scan()
				}
			}
			return LiteralType{Values: values}, nil

		default:
			// Unknown generic, skip to ]
			p.SkipUntil(parser.RBRACKET)
			p.p.Scan() // consume ]
			return ReferenceType{Name: name}, nil
		}
	}

	return ReferenceType{Name: name}, nil
}
