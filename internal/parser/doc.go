// Package parser provides shared lexer and parser infrastructure for parsing
// generated type definition files.
//
// This package is used by generators to parse existing output files back into
// AST representations for comparison. It provides a token-based lexer and a
// parser with lookahead and rollback capabilities.
//
// # Lexer
//
// The Lexer tokenizes source code into a stream of tokens. It handles:
//
//   - Identifiers and keywords (with pluggable keyword lookup)
//   - Numbers (integers and floats, with underscore separators)
//   - Strings (single and double quoted, with escape sequences)
//   - Comments (line // and block /* */ style, plus Python # style)
//   - Operators and punctuation
//   - Backtick strings (for Go struct tags)
//   - Python docstrings (triple-quoted strings)
//
// # Parser
//
// The Parser wraps a Lexer and provides:
//
//   - Token buffering for lookahead (Peek, PeekIgnoreWhitespace)
//   - Rollback support for backtracking (Rollback, Index)
//   - Whitespace-ignoring scan methods (ScanIgnoreWhitespace)
//   - Expect methods for required tokens (Expect, ExpectIdent, ExpectString)
//   - Skip utilities (SkipUntil)
//
// # Keyword Lookup
//
// Each language has different keywords. The lexer accepts a KeywordLookup function
// that maps identifier strings to keyword tokens. This allows the same lexer to
// be used for TypeScript, Python, and Go parsing.
//
// # Token Types
//
// Common tokens defined in this package:
//
//   - ILLEGAL, EOF: Error and end-of-file markers
//   - IDENT, NUMBER, STRING: Literals
//   - NEWLINE, COMMENT: Whitespace and comments
//   - Punctuation: LBRACE, RBRACE, LPAREN, RPAREN, etc.
//   - Operators: EQUALS, PIPE, AMPERSAND, LESS, GREATER, etc.
//
// Language-specific keyword tokens (EXPORT, INTERFACE, CLASS, TYPE, etc.) are
// defined in each generator's parser.go file and mapped via KeywordLookup.
package parser
