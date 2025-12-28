package golang

import (
	"strings"

	"github.com/hntrl/sparktype/internal/parser"
)

// Go-specific tokens
const (
	PACKAGE parser.Token = parser.KEYWORD_START + iota
	IMPORT
	TYPE
	STRUCT
	INTERFACE
	MAP
	CONST
	FUNC
)

var goKeywords = map[string]parser.Token{
	"package":   PACKAGE,
	"import":    IMPORT,
	"type":      TYPE,
	"struct":    STRUCT,
	"interface": INTERFACE,
	"map":       MAP,
	"const":     CONST,
	"func":      FUNC,
}

func goKeywordLookup(lit string) parser.Token {
	if tok, ok := goKeywords[lit]; ok {
		return tok
	}
	return parser.ILLEGAL
}

// Parser parses Go source into AST nodes
type Parser struct {
	p *parser.Parser
}

// NewParser creates a new Go parser
func NewParser(input string) *Parser {
	p := parser.NewParser(input)
	p.SetKeywordLookup(goKeywordLookup)
	return &Parser{p: p}
}

// ParseFile parses a Go file from a Parser
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

		switch item.Tok {
		case PACKAGE:
			pkgItem := p.p.ScanIgnoreWhitespace()
			if pkgItem.Tok == parser.IDENT {
				file.Package = pkgItem.Lit
			}

		case IMPORT:
			imports, err := p.ParseImports()
			if err != nil {
				return nil, err
			}
			file.Imports = append(file.Imports, imports...)

		case TYPE:
			node, err := p.ParseTypeDecl()
			if err != nil {
				return nil, err
			}
			if node != nil {
				file.Types = append(file.Types, node)
			}

		case CONST:
			// Parse const blocks - they might be enum values
			enumDef, err := p.ParseConstBlock()
			if err != nil {
				return nil, err
			}
			if enumDef != nil {
				// Look for matching TypeDef to convert to EnumDef
				for i, existingType := range file.Types {
					if td, ok := existingType.(*TypeDef); ok && td.Name == enumDef.Name {
						// Replace TypeDef with EnumDef
						file.Types[i] = enumDef
						enumDef = nil
						break
					}
				}
				// If no matching TypeDef was found, still add the EnumDef
				if enumDef != nil {
					file.Types = append(file.Types, enumDef)
				}
			}
		}
	}

	return file, nil
}

func (p *Parser) ParseImports() ([]string, error) {
	var imports []string

	item := p.p.ScanIgnoreWhitespace()

	if item.Tok == parser.STRING {
		// Single import
		imports = append(imports, item.Lit)
	} else if item.Tok == parser.LPAREN {
		// Import block
		for {
			item = p.p.ScanIgnoreWhitespace()
			if item.Tok == parser.RPAREN {
				break
			}
			if item.Tok == parser.STRING {
				imports = append(imports, item.Lit)
			}
		}
	}

	return imports, nil
}

func (p *Parser) ParseTypeDecl() (Node, error) {
	// Parse type name
	nameItem := p.p.ScanIgnoreWhitespace()
	if nameItem.Tok != parser.IDENT {
		return nil, parser.NewParseError(nameItem.Pos, "expected type name")
	}
	name := nameItem.Lit

	// Check what follows
	item := p.p.ScanIgnoreWhitespace()

	if item.Tok == STRUCT {
		return p.ParseStructDecl(name)
	}

	if item.Tok == INTERFACE {
		// Skip interface definitions for now
		p.SkipBraces()
		return nil, nil
	}

	if item.Tok == parser.EQUALS {
		// Type alias: type Name = Type
		typeExpr, err := p.ParseGoTypeExpr()
		if err != nil {
			return nil, err
		}
		return &TypeAlias{Name: name, Type: typeExpr}, nil
	}

	// Type definition: type Name Type
	p.p.Unscan()
	typeExpr, err := p.ParseGoTypeExpr()
	if err != nil {
		return nil, err
	}

	// Check if it looks like an enum (type Name string)
	if _, ok := typeExpr.(GoString); ok {
		return &TypeDef{Name: name, Type: typeExpr}, nil
	}

	return &TypeDef{Name: name, Type: typeExpr}, nil
}

func (p *Parser) SkipBraces() {
	depth := 0
	foundFirst := false
	for {
		item := p.p.Scan()
		if item.Tok == parser.EOF {
			return
		}
		if item.Tok == parser.LBRACE {
			depth++
			foundFirst = true
		} else if item.Tok == parser.RBRACE {
			depth--
			if foundFirst && depth == 0 {
				return
			}
		}
	}
}

func (p *Parser) SkipConstBlock() {
	item := p.p.ScanIgnoreWhitespace()
	if item.Tok == parser.LPAREN {
		depth := 1
		for depth > 0 {
			item = p.p.Scan()
			if item.Tok == parser.EOF {
				return
			}
			if item.Tok == parser.LPAREN {
				depth++
			} else if item.Tok == parser.RPAREN {
				depth--
			}
		}
	}
}

// ParseConstBlock parses a const block and returns an EnumDef if it looks like an enum
func (p *Parser) ParseConstBlock() (*EnumDef, error) {
	item := p.p.ScanIgnoreWhitespace()
	if item.Tok != parser.LPAREN {
		// Single const, not a block - skip it
		p.p.Unscan()
		p.SkipToNewline()
		return nil, nil
	}

	// Parse const block looking for pattern: ConstName TypeName = "value"
	var values []EnumValue
	var enumTypeName string

	for {
		item = p.p.ScanIgnoreWhitespace()
		if item.Tok == parser.RPAREN {
			break
		}
		if item.Tok == parser.EOF {
			return nil, parser.NewParseError(item.Pos, "unexpected EOF in const block")
		}
		if item.Tok == parser.COMMENT || item.Tok == parser.NEWLINE {
			continue
		}

		if item.Tok == parser.IDENT {
			constName := item.Lit

			// Check for type name
			typeItem := p.p.ScanIgnoreWhitespace()
			if typeItem.Tok == parser.IDENT {
				// We have a type name
				typeName := typeItem.Lit
				if enumTypeName == "" {
					enumTypeName = typeName
				}

				// Look for = "value"
				if p.p.Match(parser.EQUALS) {
					p.p.Scan()
					valueItem := p.p.ScanIgnoreWhitespace()
					if valueItem.Tok == parser.STRING {
						// The const name typically starts with the type name (e.g., StatusActive for Status enum)
						// Strip the type name prefix to get just the value name
						valueName := constName
						if strings.HasPrefix(constName, enumTypeName) {
							valueName = constName[len(enumTypeName):]
						}
						values = append(values, EnumValue{
							Name:  valueName,
							Value: valueItem.Lit,
						})
					}
				}
			} else {
				p.p.Unscan()
			}
		}
	}

	if len(values) > 0 && enumTypeName != "" {
		return &EnumDef{
			Name:   enumTypeName,
			Values: values,
		}, nil
	}

	return nil, nil
}

func (p *Parser) SkipToNewline() {
	for {
		item := p.p.Scan()
		if item.Tok == parser.EOF || item.Tok == parser.NEWLINE {
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

func (p *Parser) ParseStructDecl(name string) (*Struct, error) {
	s := &Struct{Name: name}

	// Expect {
	if _, err := p.p.Expect(parser.LBRACE); err != nil {
		return nil, err
	}

	// Parse fields
	for {
		item := p.p.ScanIgnoreWhitespace()

		if item.Tok == parser.RBRACE {
			break
		}

		if item.Tok == parser.EOF {
			return nil, parser.NewParseError(item.Pos, "unexpected EOF in struct")
		}

		if item.Tok == parser.COMMENT {
			continue
		}

		if item.Tok == parser.IDENT {
			fieldName := item.Lit

			// Parse type
			typeExpr, err := p.ParseGoTypeExpr()
			if err != nil {
				return nil, err
			}

			field := StructField{Name: fieldName, Type: typeExpr}

			// Check for struct tags
			item = p.p.ScanIgnoreWhitespace()
			if item.Tok == parser.BACKTICK {
				tag := p.p.ScanBacktickString()
				field.JSONTag = ExtractJSONTag(tag)
			} else {
				p.p.Unscan()
			}

			s.Fields = append(s.Fields, field)
		}
	}

	return s, nil
}

// ExtractJSONTag extracts the json tag value from a struct tag string
func ExtractJSONTag(tag string) string {
	// Parse: json:"fieldname,omitempty"
	parts := strings.Split(tag, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "json:") {
			jsonPart := strings.TrimPrefix(part, "json:")
			jsonPart = strings.Trim(jsonPart, `"`)
			return jsonPart
		}
	}
	return ""
}

// --- Type Expression Parsing ---

func (p *Parser) ParseGoTypeExpr() (GoTypeExpr, error) {
	item := p.p.ScanIgnoreWhitespace()

	switch item.Tok {
	case parser.IDENT:
		return p.ParseGoIdentType(item.Lit)

	case parser.LBRACKET:
		// Slice: []Type or [n]Type
		next := p.p.ScanIgnoreWhitespace()
		if next.Tok == parser.RBRACKET {
			// []Type
			elem, err := p.ParseGoTypeExpr()
			if err != nil {
				return nil, err
			}
			// Check for []byte
			if str, ok := elem.(GoString); ok && str.Serialize() == "byte" {
				return GoBytes{}, nil
			}
			return GoSlice{Element: elem}, nil
		}
		// [n]Type - skip the size and parse as slice for simplicity
		p.SkipUntil(parser.RBRACKET)
		p.p.Scan() // consume ]
		elem, err := p.ParseGoTypeExpr()
		if err != nil {
			return nil, err
		}
		return GoSlice{Element: elem}, nil

	case MAP:
		// map[K]V
		if _, err := p.p.Expect(parser.LBRACKET); err != nil {
			return nil, err
		}
		key, err := p.ParseGoTypeExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.p.Expect(parser.RBRACKET); err != nil {
			return nil, err
		}
		value, err := p.ParseGoTypeExpr()
		if err != nil {
			return nil, err
		}
		return GoMap{Key: key, Value: value}, nil

	case parser.AMPERSAND:
		fallthrough
	case parser.ILLEGAL:
		if item.Lit == "*" {
			// Pointer: *Type
			inner, err := p.ParseGoTypeExpr()
			if err != nil {
				return nil, err
			}
			return GoPointer{Inner: inner}, nil
		}
		return nil, parser.NewParseError(item.Pos, "unexpected token in type: %s", item.Lit)

	case INTERFACE:
		// interface{}
		p.p.Expect(parser.LBRACE)
		p.p.Expect(parser.RBRACE)
		return GoInterface{}, nil

	case STRUCT:
		// Anonymous struct - skip for now
		p.SkipBraces()
		return GoInterface{}, nil

	default:
		// Check if it's * (pointer)
		if item.Lit == "*" {
			inner, err := p.ParseGoTypeExpr()
			if err != nil {
				return nil, err
			}
			return GoPointer{Inner: inner}, nil
		}
		return nil, parser.NewParseError(item.Pos, "unexpected token in type: %s (%s)", item.Tok, item.Lit)
	}
}

func (p *Parser) ParseGoIdentType(name string) (GoTypeExpr, error) {
	// Check for qualified names (e.g., time.Time)
	if p.p.Match(parser.DOT) {
		p.p.Scan()
		qualItem := p.p.ScanIgnoreWhitespace()
		if qualItem.Tok == parser.IDENT {
			fullName := name + "." + qualItem.Lit
			if fullName == "time.Time" {
				return GoTime{}, nil
			}
			return GoReference{Name: fullName}, nil
		}
	}

	// Built-in types
	switch name {
	case "string":
		return GoString{}, nil
	case "int":
		return GoInt{}, nil
	case "int32":
		return GoInt32{}, nil
	case "int64":
		return GoInt64{}, nil
	case "float64":
		return GoFloat64{}, nil
	case "bool":
		return GoBool{}, nil
	case "byte":
		return GoString{}, nil // Will be handled specially for []byte
	case "interface":
		// interface{}
		p.p.Expect(parser.LBRACE)
		p.p.Expect(parser.RBRACE)
		return GoInterface{}, nil
	default:
		return GoReference{Name: name}, nil
	}
}
