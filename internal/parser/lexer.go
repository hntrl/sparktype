package parser

import (
	"bufio"
	"io"
	"strings"
	"unicode"
)

// KeywordLookup is a function that checks if an identifier is a keyword
type KeywordLookup func(lit string) Token

// Lexer tokenizes source code
type Lexer struct {
	pos           Position
	reader        *bufio.Reader
	keywordLookup KeywordLookup
}

// NewLexer creates a new lexer for the given input
func NewLexer(input string) *Lexer {
	return &Lexer{
		pos:    Position{Line: 1, Column: 0},
		reader: bufio.NewReader(strings.NewReader(input)),
	}
}

// SetKeywordLookup sets a custom keyword lookup function
func (l *Lexer) SetKeywordLookup(fn KeywordLookup) {
	l.keywordLookup = fn
}

func (l *Lexer) read() rune {
	r, _, err := l.reader.ReadRune()
	if err != nil {
		return 0
	}
	l.pos.Column++
	return r
}

func (l *Lexer) backup() {
	l.pos.Column--
	l.reader.UnreadRune()
}

func (l *Lexer) peek() rune {
	r := l.read()
	if r != 0 {
		l.backup()
	}
	return r
}

func (l *Lexer) resetLine() {
	l.pos.Line++
	l.pos.Column = 0
}

// LexIdent reads an identifier
func (l *Lexer) LexIdent() string {
	var lit strings.Builder
	for {
		r := l.read()
		if r == 0 {
			break
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '$' {
			lit.WriteRune(r)
		} else {
			l.backup()
			break
		}
	}
	return lit.String()
}

// LexNumber reads a number (integer or float)
func (l *Lexer) LexNumber() string {
	var lit strings.Builder
	seenDot := false

	for {
		r := l.read()
		if r == 0 {
			break
		}
		if unicode.IsDigit(r) {
			lit.WriteRune(r)
		} else if r == '.' && !seenDot {
			// Check if next is digit (float) or something else (method call)
			next := l.peek()
			if unicode.IsDigit(next) {
				seenDot = true
				lit.WriteRune(r)
			} else {
				l.backup()
				break
			}
		} else if r == '_' {
			// Allow underscores in numbers (e.g., 1_000_000)
			continue
		} else {
			l.backup()
			break
		}
	}
	return lit.String()
}

// LexString reads a quoted string
func (l *Lexer) LexString(quote rune) string {
	var lit strings.Builder

	for {
		r := l.read()
		if r == 0 {
			break
		}
		if r == '\\' {
			// Handle escape sequences
			next := l.read()
			switch next {
			case 'n':
				lit.WriteRune('\n')
			case 't':
				lit.WriteRune('\t')
			case 'r':
				lit.WriteRune('\r')
			case '\\':
				lit.WriteRune('\\')
			case quote:
				lit.WriteRune(quote)
			default:
				lit.WriteRune('\\')
				lit.WriteRune(next)
			}
		} else if r == quote {
			break
		} else if r == '\n' {
			l.resetLine()
			lit.WriteRune(r)
		} else {
			lit.WriteRune(r)
		}
	}
	return lit.String()
}

// LexBacktickString reads a backtick-quoted string (for Go struct tags)
func (l *Lexer) LexBacktickString() string {
	var lit strings.Builder

	for {
		r := l.read()
		if r == 0 || r == '`' {
			break
		}
		if r == '\n' {
			l.resetLine()
		}
		lit.WriteRune(r)
	}
	return lit.String()
}

// LexRawUntil reads characters until the stop character is reached
// Used for content like regex patterns that don't follow normal tokenization
func (l *Lexer) LexRawUntil(stop rune) string {
	var lit strings.Builder

	for {
		r := l.read()
		if r == 0 || r == stop {
			break
		}
		if r == '\n' {
			l.resetLine()
		}
		lit.WriteRune(r)
	}
	return lit.String()
}

// LexRegexUntilClose reads a regex pattern /.../ until seeing /) which closes the regex call
// This handles regex patterns that may contain ) inside like /(?:pattern)+/
func (l *Lexer) LexRegexUntilClose() string {
	var lit strings.Builder

	inRegex := false
	for {
		r := l.read()
		if r == 0 {
			break
		}
		if r == '\n' {
			l.resetLine()
		}

		// Track when we enter/exit the regex (between / /)
		if r == '/' {
			if !inRegex {
				// Start of regex
				inRegex = true
				lit.WriteRune(r)
				continue
			} else {
				// End of regex - check if next is )
				lit.WriteRune(r)
				next := l.peek()
				if next == ')' {
					l.read() // consume the )
					break
				}
				continue
			}
		}

		lit.WriteRune(r)
	}
	return lit.String()
}

// LexLineComment reads a line comment (// ...)
func (l *Lexer) LexLineComment() string {
	var lit strings.Builder

	for {
		r := l.read()
		if r == 0 || r == '\n' {
			if r == '\n' {
				l.backup() // Let main loop handle newline
			}
			break
		}
		lit.WriteRune(r)
	}
	return lit.String()
}

// LexBlockComment reads a block comment (/* ... */)
func (l *Lexer) LexBlockComment() string {
	var lit strings.Builder

	for {
		r := l.read()
		if r == 0 {
			break
		}
		if r == '\n' {
			l.resetLine()
			lit.WriteRune(r)
		} else if r == '*' {
			if l.peek() == '/' {
				l.read() // consume '/'
				break
			}
			lit.WriteRune(r)
		} else {
			lit.WriteRune(r)
		}
	}
	return lit.String()
}

// LexPythonDocstring reads a Python docstring (""" ... """)
func (l *Lexer) LexPythonDocstring() string {
	var lit strings.Builder

	for {
		r := l.read()
		if r == 0 {
			break
		}
		if r == '\n' {
			l.resetLine()
			lit.WriteRune(r)
		} else if r == '"' {
			// Check for closing """
			if l.peek() == '"' {
				l.read()
				if l.peek() == '"' {
					l.read()
					break
				}
				lit.WriteRune('"')
				lit.WriteRune('"')
			} else {
				lit.WriteRune(r)
			}
		} else {
			lit.WriteRune(r)
		}
	}
	return lit.String()
}

// Lex returns the next token
func (l *Lexer) Lex() TokenItem {
	for {
		r, _, err := l.reader.ReadRune()
		if err != nil {
			if err == io.EOF {
				return TokenItem{Pos: l.pos, Tok: EOF, Lit: ""}
			}
			return TokenItem{Pos: l.pos, Tok: ILLEGAL, Lit: ""}
		}
		l.pos.Column++

		// Skip whitespace (except newlines)
		if unicode.IsSpace(r) && r != '\n' {
			continue
		}

		startPos := l.pos

		// Newline
		if r == '\n' {
			l.resetLine()
			return TokenItem{Pos: startPos, Tok: NEWLINE, Lit: "\\n"}
		}

		// Identifiers and keywords
		if unicode.IsLetter(r) || r == '_' || r == '$' {
			l.backup()
			lit := l.LexIdent()
			tok := IDENT
			if l.keywordLookup != nil {
				if kwTok := l.keywordLookup(lit); kwTok != ILLEGAL {
					tok = kwTok
				}
			}
			return TokenItem{Pos: startPos, Tok: tok, Lit: lit}
		}

		// Numbers
		if unicode.IsDigit(r) {
			l.backup()
			lit := l.LexNumber()
			return TokenItem{Pos: startPos, Tok: NUMBER, Lit: lit}
		}

		// Strings
		if r == '"' || r == '\'' {
			lit := l.LexString(r)
			return TokenItem{Pos: startPos, Tok: STRING, Lit: lit}
		}

		// Single character tokens
		switch r {
		case '{':
			return TokenItem{Pos: startPos, Tok: LBRACE, Lit: "{"}
		case '}':
			return TokenItem{Pos: startPos, Tok: RBRACE, Lit: "}"}
		case '[':
			return TokenItem{Pos: startPos, Tok: LBRACKET, Lit: "["}
		case ']':
			return TokenItem{Pos: startPos, Tok: RBRACKET, Lit: "]"}
		case '(':
			return TokenItem{Pos: startPos, Tok: LPAREN, Lit: "("}
		case ')':
			return TokenItem{Pos: startPos, Tok: RPAREN, Lit: ")"}
		case ',':
			return TokenItem{Pos: startPos, Tok: COMMA, Lit: ","}
		case ':':
			return TokenItem{Pos: startPos, Tok: COLON, Lit: ":"}
		case ';':
			return TokenItem{Pos: startPos, Tok: SEMICOLON, Lit: ";"}
		case '.':
			return TokenItem{Pos: startPos, Tok: DOT, Lit: "."}
		case '|':
			return TokenItem{Pos: startPos, Tok: PIPE, Lit: "|"}
		case '&':
			return TokenItem{Pos: startPos, Tok: AMPERSAND, Lit: "&"}
		case '?':
			return TokenItem{Pos: startPos, Tok: QUESTION, Lit: "?"}
		case '=':
			return TokenItem{Pos: startPos, Tok: EQUALS, Lit: "="}
		case '<':
			return TokenItem{Pos: startPos, Tok: LESS, Lit: "<"}
		case '>':
			return TokenItem{Pos: startPos, Tok: GREATER, Lit: ">"}
		case '`':
			return TokenItem{Pos: startPos, Tok: BACKTICK, Lit: "`"}
		case '/':
			// Check for comments
			if l.peek() == '/' {
				l.read() // consume second /
				lit := l.LexLineComment()
				return TokenItem{Pos: startPos, Tok: COMMENT, Lit: lit}
			} else if l.peek() == '*' {
				l.read() // consume *
				lit := l.LexBlockComment()
				return TokenItem{Pos: startPos, Tok: COMMENT, Lit: lit}
			}
			return TokenItem{Pos: startPos, Tok: ILLEGAL, Lit: string(r)}
		case '#':
			// Python-style comment
			lit := l.LexLineComment()
			return TokenItem{Pos: startPos, Tok: COMMENT, Lit: lit}
		default:
			return TokenItem{Pos: startPos, Tok: ILLEGAL, Lit: string(r)}
		}
	}
}
