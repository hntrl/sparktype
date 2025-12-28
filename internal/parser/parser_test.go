package parser

import (
	"testing"
)

func TestNewParser(t *testing.T) {
	p := NewParser("test input")
	if p == nil {
		t.Fatal("expected non-nil parser")
	}
	if p.pos != -1 {
		t.Errorf("expected initial pos -1, got %d", p.pos)
	}
}

func TestParser_Scan(t *testing.T) {
	p := NewParser("foo bar baz")

	item := p.Scan()
	if item.Lit != "foo" {
		t.Errorf("expected 'foo', got %q", item.Lit)
	}

	item = p.Scan()
	if item.Lit != "bar" {
		t.Errorf("expected 'bar', got %q", item.Lit)
	}

	item = p.Scan()
	if item.Lit != "baz" {
		t.Errorf("expected 'baz', got %q", item.Lit)
	}

	item = p.Scan()
	if item.Tok != EOF {
		t.Errorf("expected EOF, got %s", item.Tok)
	}
}

func TestParser_ScanIgnore(t *testing.T) {
	p := NewParser("foo\n\nbar")

	// Scan ignoring newlines
	item := p.ScanIgnore(NEWLINE)
	if item.Lit != "foo" {
		t.Errorf("expected 'foo', got %q", item.Lit)
	}

	item = p.ScanIgnore(NEWLINE)
	if item.Lit != "bar" {
		t.Errorf("expected 'bar', got %q", item.Lit)
	}
}

func TestParser_ScanIgnoreWhitespace(t *testing.T) {
	p := NewParser("foo\n// comment\nbar")

	item := p.ScanIgnoreWhitespace()
	if item.Lit != "foo" {
		t.Errorf("expected 'foo', got %q", item.Lit)
	}

	item = p.ScanIgnoreWhitespace()
	if item.Lit != "bar" {
		t.Errorf("expected 'bar', got %q", item.Lit)
	}
}

func TestParser_Peek(t *testing.T) {
	p := NewParser("foo bar")

	// Peek should not consume
	item := p.Peek()
	if item.Lit != "foo" {
		t.Errorf("expected 'foo', got %q", item.Lit)
	}

	// Peek again should return same
	item = p.Peek()
	if item.Lit != "foo" {
		t.Errorf("expected 'foo', got %q", item.Lit)
	}

	// Now scan
	item = p.Scan()
	if item.Lit != "foo" {
		t.Errorf("expected 'foo', got %q", item.Lit)
	}

	// Peek should return bar now
	item = p.Peek()
	if item.Lit != "bar" {
		t.Errorf("expected 'bar', got %q", item.Lit)
	}
}

func TestParser_PeekIgnoreWhitespace(t *testing.T) {
	p := NewParser("foo\n\nbar")

	p.Scan() // consume foo

	// PeekIgnoreWhitespace should skip newlines
	item := p.PeekIgnoreWhitespace()
	if item.Lit != "bar" {
		t.Errorf("expected 'bar', got %q", item.Lit)
	}
}

func TestParser_Unscan(t *testing.T) {
	p := NewParser("foo bar baz")

	p.Scan() // foo
	p.Scan() // bar
	p.Unscan()

	item := p.Scan()
	if item.Lit != "bar" {
		t.Errorf("expected 'bar', got %q", item.Lit)
	}
}

func TestParser_Rollback(t *testing.T) {
	p := NewParser("foo bar baz")

	p.Scan() // foo
	mark := p.Index()

	p.Scan() // bar
	p.Scan() // baz

	p.Rollback(mark)

	item := p.Scan()
	if item.Lit != "bar" {
		t.Errorf("expected 'bar', got %q", item.Lit)
	}
}

func TestParser_Index(t *testing.T) {
	p := NewParser("foo bar")

	if p.Index() != -1 {
		t.Errorf("expected initial index -1, got %d", p.Index())
	}

	p.Scan()
	if p.Index() != 0 {
		t.Errorf("expected index 0 after first scan, got %d", p.Index())
	}

	p.Scan()
	if p.Index() != 1 {
		t.Errorf("expected index 1 after second scan, got %d", p.Index())
	}
}

func TestParser_Current(t *testing.T) {
	p := NewParser("foo bar")

	// Before any scan, current should be ILLEGAL
	item := p.Current()
	if item.Tok != ILLEGAL {
		t.Errorf("expected ILLEGAL before scan, got %s", item.Tok)
	}

	p.Scan()
	item = p.Current()
	if item.Lit != "foo" {
		t.Errorf("expected 'foo', got %q", item.Lit)
	}

	p.Scan()
	item = p.Current()
	if item.Lit != "bar" {
		t.Errorf("expected 'bar', got %q", item.Lit)
	}
}

func TestParser_Expect(t *testing.T) {
	p := NewParser("foo 123")

	// Expect IDENT
	item, err := p.Expect(IDENT)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if item.Lit != "foo" {
		t.Errorf("expected 'foo', got %q", item.Lit)
	}

	// Expect NUMBER
	item, err = p.Expect(NUMBER)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if item.Lit != "123" {
		t.Errorf("expected '123', got %q", item.Lit)
	}
}

func TestParser_Expect_Error(t *testing.T) {
	p := NewParser("foo")

	_, err := p.Expect(NUMBER)
	if err == nil {
		t.Error("expected error when expecting NUMBER but got IDENT")
	}
}

func TestParser_ExpectIdent(t *testing.T) {
	p := NewParser("foo")

	lit, err := p.ExpectIdent()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if lit != "foo" {
		t.Errorf("expected 'foo', got %q", lit)
	}
}

func TestParser_ExpectIdent_Error(t *testing.T) {
	p := NewParser("123")

	_, err := p.ExpectIdent()
	if err == nil {
		t.Error("expected error when expecting IDENT but got NUMBER")
	}
}

func TestParser_ExpectString(t *testing.T) {
	p := NewParser(`"hello"`)

	lit, err := p.ExpectString()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if lit != "hello" {
		t.Errorf("expected 'hello', got %q", lit)
	}
}

func TestParser_ExpectString_Error(t *testing.T) {
	p := NewParser("123")

	_, err := p.ExpectString()
	if err == nil {
		t.Error("expected error when expecting STRING but got NUMBER")
	}
}

func TestParser_Match(t *testing.T) {
	p := NewParser("foo bar")

	if !p.Match(IDENT) {
		t.Error("expected Match(IDENT) to be true")
	}

	if p.Match(NUMBER) {
		t.Error("expected Match(NUMBER) to be false")
	}

	// Match should not consume the token
	item := p.Scan()
	if item.Lit != "foo" {
		t.Errorf("expected 'foo', got %q", item.Lit)
	}
}

func TestParser_MatchAny(t *testing.T) {
	p := NewParser("foo")

	if !p.MatchAny(IDENT, NUMBER) {
		t.Error("expected MatchAny(IDENT, NUMBER) to be true")
	}

	if p.MatchAny(NUMBER, STRING) {
		t.Error("expected MatchAny(NUMBER, STRING) to be false")
	}
}

func TestParser_SkipUntil(t *testing.T) {
	p := NewParser("foo bar baz { content }")

	item := p.SkipUntil(LBRACE)
	if item.Tok != LBRACE {
		t.Errorf("expected LBRACE, got %s", item.Tok)
	}

	// Next should be content
	item = p.ScanIgnoreWhitespace()
	if item.Lit != "content" {
		t.Errorf("expected 'content', got %q", item.Lit)
	}
}

func TestParser_SkipUntil_EOF(t *testing.T) {
	p := NewParser("foo bar baz")

	item := p.SkipUntil(LBRACE) // Not found
	if item.Tok != EOF {
		t.Errorf("expected EOF, got %s", item.Tok)
	}
}

func TestParser_SetKeywordLookup(t *testing.T) {
	p := NewParser("export const foo")

	keywordLookup := func(lit string) Token {
		switch lit {
		case "export":
			return Token(100)
		case "const":
			return Token(101)
		}
		return ILLEGAL
	}

	p.SetKeywordLookup(keywordLookup)

	item := p.Scan()
	if item.Tok != Token(100) {
		t.Errorf("expected keyword token 100, got %s", item.Tok)
	}

	item = p.Scan()
	if item.Tok != Token(101) {
		t.Errorf("expected keyword token 101, got %s", item.Tok)
	}

	item = p.Scan()
	if item.Tok != IDENT {
		t.Errorf("expected IDENT, got %s", item.Tok)
	}
}

func TestParser_ScanBacktickString(t *testing.T) {
	p := NewParser("`json:\"name\"`")

	// First scan the backtick
	item := p.Scan()
	if item.Tok != BACKTICK {
		t.Errorf("expected BACKTICK, got %s", item.Tok)
	}

	// Now read the content
	content := p.ScanBacktickString()
	expected := `json:"name"`
	if content != expected {
		t.Errorf("expected %q, got %q", expected, content)
	}
}

func TestParseError(t *testing.T) {
	pos := Position{Line: 10, Column: 5}
	err := NewParseError(pos, "unexpected token %s", "ILLEGAL")

	expected := "parse error at line 10, column 5: unexpected token ILLEGAL"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestParser_BufferReuse(t *testing.T) {
	p := NewParser("a b c")

	// Scan all tokens
	p.Scan() // a
	p.Scan() // b
	p.Scan() // c

	// Rollback to beginning
	p.Rollback(-1)

	// Should be able to re-scan from buffer
	item := p.Scan()
	if item.Lit != "a" {
		t.Errorf("expected 'a', got %q", item.Lit)
	}

	item = p.Scan()
	if item.Lit != "b" {
		t.Errorf("expected 'b', got %q", item.Lit)
	}
}

func TestParser_ComplexParsing(t *testing.T) {
	// Simulate parsing a simple type definition
	p := NewParser("type User = { name: string }")

	keywordLookup := func(lit string) Token {
		switch lit {
		case "type":
			return Token(100)
		case "string":
			return Token(101)
		}
		return ILLEGAL
	}
	p.SetKeywordLookup(keywordLookup)

	// type
	item := p.ScanIgnoreWhitespace()
	if item.Tok != Token(100) {
		t.Errorf("expected 'type' keyword, got %s (%q)", item.Tok, item.Lit)
	}

	// User
	name, err := p.ExpectIdent()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if name != "User" {
		t.Errorf("expected 'User', got %q", name)
	}

	// =
	_, err = p.Expect(EQUALS)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// {
	_, err = p.Expect(LBRACE)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// name
	propName, err := p.ExpectIdent()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if propName != "name" {
		t.Errorf("expected 'name', got %q", propName)
	}

	// :
	_, err = p.Expect(COLON)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// string
	item = p.ScanIgnoreWhitespace()
	if item.Tok != Token(101) {
		t.Errorf("expected 'string' keyword, got %s (%q)", item.Tok, item.Lit)
	}

	// }
	_, err = p.Expect(RBRACE)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestParser_ScanRawUntil(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		stop     rune
		expected string
	}{
		{"simple", "hello)", ')', "hello"},
		{"with spaces", "hello world)", ')', "hello world"},
		{"empty", ")", ')', ""},
		{"stops at first occurrence", "a(b)c)", ')', "a(b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			result := p.ScanRawUntil(tt.stop)
			if result != tt.expected {
				t.Errorf("ScanRawUntil(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParser_ScanRegexUntilClose(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple regex", "/abc/)", "/abc/"},
		{"regex with special chars", "/^[a-z]+$/)", "/^[a-z]+$/"},
		{"regex with parens inside", "/(?:foo)+/)", "/(?:foo)+/"},
		{"complex regex with nested groups", "/^[a-z0-9]+(?:-[a-z0-9]+)*$/)", "/^[a-z0-9]+(?:-[a-z0-9]+)*$/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser(tt.input)
			result := p.ScanRegexUntilClose()
			if result != tt.expected {
				t.Errorf("ScanRegexUntilClose(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
