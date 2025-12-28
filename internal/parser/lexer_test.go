package parser

import (
	"testing"
)

func TestLexer_Lex_Identifiers(t *testing.T) {
	tests := []struct {
		input    string
		expected []TokenItem
	}{
		{
			input: "foo bar",
			expected: []TokenItem{
				{Tok: IDENT, Lit: "foo"},
				{Tok: IDENT, Lit: "bar"},
				{Tok: EOF},
			},
		},
		{
			input: "camelCase PascalCase snake_case",
			expected: []TokenItem{
				{Tok: IDENT, Lit: "camelCase"},
				{Tok: IDENT, Lit: "PascalCase"},
				{Tok: IDENT, Lit: "snake_case"},
				{Tok: EOF},
			},
		},
		{
			input: "_private $jquery",
			expected: []TokenItem{
				{Tok: IDENT, Lit: "_private"},
				{Tok: IDENT, Lit: "$jquery"},
				{Tok: EOF},
			},
		},
		{
			input: "ident123 _123",
			expected: []TokenItem{
				{Tok: IDENT, Lit: "ident123"},
				{Tok: IDENT, Lit: "_123"},
				{Tok: EOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			for i, exp := range tt.expected {
				item := lexer.Lex()
				if item.Tok != exp.Tok {
					t.Errorf("token %d: expected tok %s, got %s", i, exp.Tok, item.Tok)
				}
				if exp.Lit != "" && item.Lit != exp.Lit {
					t.Errorf("token %d: expected lit %q, got %q", i, exp.Lit, item.Lit)
				}
			}
		})
	}
}

func TestLexer_Lex_Keywords(t *testing.T) {
	keywordLookup := func(lit string) Token {
		switch lit {
		case "export", "const", "interface", "type":
			return Token(100) // Custom keyword token
		}
		return ILLEGAL
	}

	lexer := NewLexer("export const interface type regular")
	lexer.SetKeywordLookup(keywordLookup)

	expected := []struct {
		tok Token
		lit string
	}{
		{Token(100), "export"},
		{Token(100), "const"},
		{Token(100), "interface"},
		{Token(100), "type"},
		{IDENT, "regular"},
		{EOF, ""},
	}

	for i, exp := range expected {
		item := lexer.Lex()
		if item.Tok != exp.tok {
			t.Errorf("token %d: expected tok %s, got %s", i, exp.tok, item.Tok)
		}
		if exp.lit != "" && item.Lit != exp.lit {
			t.Errorf("token %d: expected lit %q, got %q", i, exp.lit, item.Lit)
		}
	}
}

func TestLexer_Lex_Numbers(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"123", "123"},
		{"0", "0"},
		{"3.14", "3.14"},
		{"1_000_000", "1000000"},
		{"1.5", "1.5"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			item := lexer.Lex()
			if item.Tok != NUMBER {
				t.Errorf("expected NUMBER, got %s", item.Tok)
			}
			if item.Lit != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, item.Lit)
			}
		})
	}
}

func TestLexer_Lex_Strings(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`'hello'`, "hello"},
		{`"hello world"`, "hello world"},
		{`"with\nnewline"`, "with\nnewline"},
		{`"with\ttab"`, "with\ttab"},
		{`"escaped\"quote"`, "escaped\"quote"},
		{`"escaped\\backslash"`, "escaped\\backslash"},
		{`'single "quoted"'`, `single "quoted"`},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			item := lexer.Lex()
			if item.Tok != STRING {
				t.Errorf("expected STRING, got %s", item.Tok)
			}
			if item.Lit != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, item.Lit)
			}
		})
	}
}

func TestLexer_Lex_Punctuation(t *testing.T) {
	input := "{}[](),:;.|&?=<>"
	expected := []Token{
		LBRACE, RBRACE,
		LBRACKET, RBRACKET,
		LPAREN, RPAREN,
		COMMA, COLON, SEMICOLON,
		DOT, PIPE, AMPERSAND,
		QUESTION, EQUALS,
		LESS, GREATER,
		EOF,
	}

	lexer := NewLexer(input)
	for i, exp := range expected {
		item := lexer.Lex()
		if item.Tok != exp {
			t.Errorf("token %d: expected %s, got %s", i, exp, item.Tok)
		}
	}
}

func TestLexer_Lex_Comments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "line comment",
			input:    "// this is a comment",
			expected: " this is a comment",
		},
		{
			name:     "block comment",
			input:    "/* block comment */",
			expected: " block comment ",
		},
		{
			name:     "python comment",
			input:    "# python comment",
			expected: " python comment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			item := lexer.Lex()
			if item.Tok != COMMENT {
				t.Errorf("expected COMMENT, got %s", item.Tok)
			}
			if item.Lit != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, item.Lit)
			}
		})
	}
}

func TestLexer_Lex_BlockCommentMultiline(t *testing.T) {
	input := `/* multi
line
comment */`
	lexer := NewLexer(input)
	item := lexer.Lex()

	if item.Tok != COMMENT {
		t.Errorf("expected COMMENT, got %s", item.Tok)
	}

	expected := " multi\nline\ncomment "
	if item.Lit != expected {
		t.Errorf("expected %q, got %q", expected, item.Lit)
	}
}

func TestLexer_Lex_Newlines(t *testing.T) {
	input := "foo\nbar"
	lexer := NewLexer(input)

	expected := []Token{IDENT, NEWLINE, IDENT, EOF}

	for i, exp := range expected {
		item := lexer.Lex()
		if item.Tok != exp {
			t.Errorf("token %d: expected %s, got %s", i, exp, item.Tok)
		}
	}
}

func TestLexer_Lex_Position(t *testing.T) {
	input := "foo\nbar baz"
	lexer := NewLexer(input)

	// foo at line 1
	item := lexer.Lex()
	if item.Pos.Line != 1 {
		t.Errorf("expected line 1, got %d", item.Pos.Line)
	}

	// newline
	lexer.Lex()

	// bar at line 2
	item = lexer.Lex()
	if item.Pos.Line != 2 {
		t.Errorf("expected line 2, got %d", item.Pos.Line)
	}
}

func TestLexer_LexBacktickString(t *testing.T) {
	input := "`json:\"name\"`"
	lexer := NewLexer(input)

	// First token is BACKTICK
	item := lexer.Lex()
	if item.Tok != BACKTICK {
		t.Errorf("expected BACKTICK, got %s", item.Tok)
	}

	// Now read the backtick string content
	content := lexer.LexBacktickString()
	expected := `json:"name"`
	if content != expected {
		t.Errorf("expected %q, got %q", expected, content)
	}
}

func TestLexer_LexPythonDocstring(t *testing.T) {
	// We need to simulate the state after reading """
	input := `This is a
multiline docstring
"""`
	lexer := NewLexer(input)
	content := lexer.LexPythonDocstring()

	expected := `This is a
multiline docstring
`
	if content != expected {
		t.Errorf("expected %q, got %q", expected, content)
	}
}

func TestLexer_Lex_ComplexExpression(t *testing.T) {
	input := `interface User {
  id: number;
  name?: string;
}`
	lexer := NewLexer(input)

	// Just verify we can lex the whole thing
	tokens := []Token{}
	for {
		item := lexer.Lex()
		tokens = append(tokens, item.Tok)
		if item.Tok == EOF {
			break
		}
	}

	// Should have multiple tokens
	if len(tokens) < 10 {
		t.Errorf("expected at least 10 tokens, got %d", len(tokens))
	}
}

func TestLexer_Lex_WhitespaceHandling(t *testing.T) {
	input := "  foo   bar  \t baz  "
	lexer := NewLexer(input)

	expected := []string{"foo", "bar", "baz"}
	for _, exp := range expected {
		item := lexer.Lex()
		if item.Tok != IDENT {
			t.Errorf("expected IDENT, got %s", item.Tok)
		}
		if item.Lit != exp {
			t.Errorf("expected %q, got %q", exp, item.Lit)
		}
	}

	// Should be EOF
	item := lexer.Lex()
	if item.Tok != EOF {
		t.Errorf("expected EOF, got %s", item.Tok)
	}
}

func TestLexer_Lex_IllegalCharacter(t *testing.T) {
	input := "@"
	lexer := NewLexer(input)
	item := lexer.Lex()

	if item.Tok != ILLEGAL {
		t.Errorf("expected ILLEGAL, got %s", item.Tok)
	}
}
