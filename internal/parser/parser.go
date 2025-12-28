package parser

import (
	"fmt"
)

// Parser provides token buffering and lookahead for parsing
type Parser struct {
	lex    *Lexer
	buffer []TokenItem
	pos    int
}

// NewParser creates a new parser for the given input
func NewParser(input string) *Parser {
	return &Parser{
		lex:    NewLexer(input),
		buffer: []TokenItem{},
		pos:    -1,
	}
}

// SetKeywordLookup sets a custom keyword lookup function on the underlying lexer
func (p *Parser) SetKeywordLookup(fn KeywordLookup) {
	p.lex.SetKeywordLookup(fn)
}

// ScanBacktickString reads a backtick-quoted string (for Go struct tags)
// Should be called right after consuming a BACKTICK token
func (p *Parser) ScanBacktickString() string {
	return p.lex.LexBacktickString()
}

// Scan returns the next token, buffering it for potential rollback
func (p *Parser) Scan() TokenItem {
	p.pos++
	if p.pos < len(p.buffer) {
		return p.buffer[p.pos]
	}
	item := p.lex.Lex()
	p.buffer = append(p.buffer, item)
	return item
}

// ScanIgnore returns the next token that is not in the ignore list
func (p *Parser) ScanIgnore(ignore ...Token) TokenItem {
	for {
		item := p.Scan()
		skip := false
		for _, tok := range ignore {
			if item.Tok == tok {
				skip = true
				break
			}
		}
		if !skip {
			return item
		}
	}
}

// ScanIgnoreWhitespace returns the next non-whitespace token
func (p *Parser) ScanIgnoreWhitespace() TokenItem {
	return p.ScanIgnore(NEWLINE, COMMENT)
}

// Peek returns the next token without consuming it
func (p *Parser) Peek() TokenItem {
	item := p.Scan()
	p.Unscan()
	return item
}

// PeekIgnoreWhitespace returns the next non-whitespace token without consuming it
func (p *Parser) PeekIgnoreWhitespace() TokenItem {
	item := p.ScanIgnoreWhitespace()
	p.Unscan()
	return item
}

// Unscan moves back one token in the buffer
func (p *Parser) Unscan() {
	if p.pos >= 0 {
		p.pos--
	}
}

// Rollback returns to a previous position in the buffer
func (p *Parser) Rollback(n int) {
	if n >= -1 && n < len(p.buffer) {
		p.pos = n
	}
}

// Index returns the current position in the buffer
func (p *Parser) Index() int {
	return p.pos
}

// Current returns the current token (last scanned)
func (p *Parser) Current() TokenItem {
	if p.pos >= 0 && p.pos < len(p.buffer) {
		return p.buffer[p.pos]
	}
	return TokenItem{Tok: ILLEGAL}
}

// Expect scans the next token and returns an error if it doesn't match
func (p *Parser) Expect(tok Token) (TokenItem, error) {
	item := p.ScanIgnoreWhitespace()
	if item.Tok != tok {
		return item, fmt.Errorf("expected %s, got %s (%q) at line %d, column %d",
			tok, item.Tok, item.Lit, item.Pos.Line, item.Pos.Column)
	}
	return item, nil
}

// ExpectIdent scans the next token and returns an error if it's not an identifier
func (p *Parser) ExpectIdent() (string, error) {
	item, err := p.Expect(IDENT)
	if err != nil {
		return "", err
	}
	return item.Lit, nil
}

// ExpectString scans the next token and returns an error if it's not a string
func (p *Parser) ExpectString() (string, error) {
	item, err := p.Expect(STRING)
	if err != nil {
		return "", err
	}
	return item.Lit, nil
}

// Match checks if the next token matches without consuming it
func (p *Parser) Match(tok Token) bool {
	item := p.PeekIgnoreWhitespace()
	return item.Tok == tok
}

// MatchAny checks if the next token matches any of the given tokens
func (p *Parser) MatchAny(tokens ...Token) bool {
	item := p.PeekIgnoreWhitespace()
	for _, tok := range tokens {
		if item.Tok == tok {
			return true
		}
	}
	return false
}

// SkipUntil consumes tokens until one of the stop tokens is found
func (p *Parser) SkipUntil(stopTokens ...Token) TokenItem {
	for {
		item := p.Scan()
		if item.Tok == EOF {
			return item
		}
		for _, stop := range stopTokens {
			if item.Tok == stop {
				return item
			}
		}
	}
}

// ScanRawUntil reads raw characters from the lexer until the stop character
// This is useful for parsing content that doesn't follow normal tokenization rules (like regex patterns)
func (p *Parser) ScanRawUntil(stop rune) string {
	return p.lex.LexRawUntil(stop)
}

// ScanRegexUntilClose reads a regex pattern until seeing /) which closes the regex in a function call
// This handles regex patterns that contain ) like /(?:pattern)+/
func (p *Parser) ScanRegexUntilClose() string {
	return p.lex.LexRegexUntilClose()
}

// ParseError represents a parsing error with position information
type ParseError struct {
	Pos     Position
	Message string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d, column %d: %s", e.Pos.Line, e.Pos.Column, e.Message)
}

// NewParseError creates a new parse error
func NewParseError(pos Position, format string, args ...any) *ParseError {
	return &ParseError{
		Pos:     pos,
		Message: fmt.Sprintf(format, args...),
	}
}
