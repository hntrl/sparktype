package parser

// Token represents a lexical token
type Token int

const (
	// Special tokens
	ILLEGAL Token = iota
	EOF
	NEWLINE
	COMMENT

	// Literals
	IDENT  // identifier
	STRING // "string" or 'string'
	NUMBER // 123 or 123.456

	// Delimiters
	LBRACE    // {
	RBRACE    // }
	LBRACKET  // [
	RBRACKET  // ]
	LPAREN    // (
	RPAREN    // )
	COMMA     // ,
	COLON     // :
	SEMICOLON // ;
	DOT       // .
	PIPE      // |
	AMPERSAND // &
	QUESTION  // ?
	EQUALS    // =
	LESS      // <
	GREATER   // >
	BACKTICK  // `

	// Keywords (language-specific parsers will define more)
	KEYWORD_START
)

var tokenNames = map[Token]string{
	ILLEGAL:   "ILLEGAL",
	EOF:       "EOF",
	NEWLINE:   "NEWLINE",
	COMMENT:   "COMMENT",
	IDENT:     "IDENT",
	STRING:    "STRING",
	NUMBER:    "NUMBER",
	LBRACE:    "LBRACE",
	RBRACE:    "RBRACE",
	LBRACKET:  "LBRACKET",
	RBRACKET:  "RBRACKET",
	LPAREN:    "LPAREN",
	RPAREN:    "RPAREN",
	COMMA:     "COMMA",
	COLON:     "COLON",
	SEMICOLON: "SEMICOLON",
	DOT:       "DOT",
	PIPE:      "PIPE",
	AMPERSAND: "AMPERSAND",
	QUESTION:  "QUESTION",
	EQUALS:    "EQUALS",
	LESS:      "LESS",
	GREATER:   "GREATER",
	BACKTICK:  "BACKTICK",
}

func (t Token) String() string {
	if name, ok := tokenNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// Position represents a position in source code
type Position struct {
	Line   int
	Column int
}

// TokenItem holds a token with its position and literal value
type TokenItem struct {
	Pos Position
	Tok Token
	Lit string
}
