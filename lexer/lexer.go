package lexer

import (
	"bytes"
	"fmt"
	"strconv"
	"unicode/utf8"
)

//go:generate stringer -type=TokenType

type TokenType uint8

const (
	_ = TokenType(iota)
	// single-character tokens
	LEFT_PAREN
	RIGHT_PAREN
	LEFT_BRACE
	RIGHT_BRACE
	COMMA
	DOT
	MINUS
	PLUS
	SEMICOLON
	SLASH
	STAR
	// one or two-character tokens
	BANG
	BANG_EQUAL
	EQUAL
	EQUAL_EQUAL
	GREATER
	GREATER_EQUAL
	LESS
	LESS_EQUAL
	COLON_EQUAL
	// literals
	IDENTIFIER
	STRING
	NUMBER
	// keywords
	AND
	OR
	ELSE
	FALSE
	FN
	FOR
	IF
	NIL
	RETURN
	SUPER
	THIS
	TRUE
	WHILE
	// meta
	EOF
)

var keywords = map[string]TokenType{
	"else":   ELSE,
	"false":  FALSE,
	"fn":     FN,
	"for":    FOR,
	"if":     IF,
	"nil":    NIL,
	"return": RETURN,
	"super":  SUPER,
	"this":   THIS,
	"true":   TRUE,
	"while":  WHILE,
}

type Token struct {
	// Note: we could store the filename information here, but that's
	// not really necessary, since we could likely stuff it in the AST's
	// root node.
	Type    TokenType
	Lexeme  string      // use utf8.RuneCountInString to get the length.
	Literal interface{} // simple literals (number, str, bool, nil)
	Line    int
	Column  int
}

type Error struct {
	Filename string
	Line     int
	Column   int
	Message  string
}

func (e *Error) Error() string { return e.String() }
func (e *Error) String() string {
	return fmt.Sprintf("%s:%d:%d: %s", e.Filename, e.Line, e.Column, e.Message)
}

type Lexer struct {
	Filename string  // filename
	source   string  // the complete source code
	Tokens   []Token // list of tokens produced
	Errors   []Error // list of lexer errors
	current  int     // where are we in the input?
	line     int     // line and column positions
	column   int     // NB: column position is in terms of runes
	start    int     // the first char of the lexeme being scanned
	startLn  int     // starting line number
	startCol int     // starting col number
	stop     bool    // whether we have met a fatal error and cannot advance any more
}

func New(filename string, source string) *Lexer {
	return &Lexer{
		Filename: filename,
		source:   source,
		Tokens:   []Token{},
		line:     1,
		column:   1,
		startLn:  1,
		startCol: 1,
	}
}

// utils

// isAtEnd lets us know if we've reached the end of the input.
func (l *Lexer) isAtEnd() bool { return l.current >= len(l.source) }

// advance consumes one rune and returns the consumed rune.
// current is incremented by the width of the returned rune.
func (l *Lexer) advance() rune {
	r, w := utf8.DecodeRuneInString(l.source[l.current:])
	if r == utf8.RuneError {
		l.error(fmt.Sprintf("invalid utf8 input at byte %d", l.current))
		l.stop = true
	}
	l.current += w
	if r == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return r
}

// peek is the same as advance, but does not advance .current.
func (l *Lexer) peek() rune {
	if l.stop || l.isAtEnd() {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.source[l.current:])
	return r
}

// peekNext peeks two runes in advance.
func (l *Lexer) peekNext() rune {
	if l.stop || l.isAtEnd() {
		return 0
	}
	_, w := utf8.DecodeRuneInString(l.source[l.current:])
	if l.current+w >= len(l.source) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.source[l.current+w:])
	return r
}

func (l *Lexer) match(ch rune) bool {
	if l.peek() != ch {
		return false
	}
	l.advance()
	return true
}

// public api, actual lexing

func (l *Lexer) ScanTokens() {
	for !l.stop && !l.isAtEnd() && len(l.Errors) <= 10 {
		l.start = l.current
		l.scanToken()
	}
	l.Tokens = append(l.Tokens, Token{EOF, "", nil, l.line, l.column})
}

func (l *Lexer) scanToken() {
	ch := l.advance()
	if l.stop {
		// invalid utf8 char
		return
	}
	switch ch {
	// Ignore whitespace
	case ' ', '\t' , '\r', '\n':
		for isWhiteSpace(l.peek()) {
			l.advance()
		}
		l.ignore()
	case ';':
		l.emit(SEMICOLON)
	case '(':
		l.emit(LEFT_PAREN)
	case ')':
		l.emit(RIGHT_PAREN)
	case '{':
		l.emit(LEFT_BRACE)
	case '}':
		l.emit(RIGHT_PAREN)
	case ',':
		l.emit(COMMA)
	case '.':
		l.emit(DOT)
	case '-':
		l.emit(MINUS)
	case '+':
		l.emit(PLUS)
	case '*':
		l.emit(STAR)
	case '/':
		if l.match('/') {
			for l.peek() != '\n' && !l.stop && !l.isAtEnd() {
				l.advance()
			}
			l.ignore()
		} else {
			l.emit(SLASH)
		}
	case '!':
		if l.match('=') {
			l.emit(BANG_EQUAL)
		} else {
			l.emit(BANG)
		}
	case '=':
		if l.match('=') {
			l.emit(EQUAL_EQUAL)
		} else {
			l.emit(EQUAL)
		}
	case '<':
		if l.match('=') {
			l.emit(LESS_EQUAL)
		} else {
			l.emit(LESS)
		}
	case '>':
		if l.match('=') {
			l.emit(GREATER_EQUAL)
		} else {
			l.emit(GREATER)
		}
	case '|':
		if l.match('|') {
			l.emit(OR)
		} else {
			l.error("invalid operator")
		}
	case '&':
		if l.match('&') {
			l.emit(AND)
		} else {
			l.error("invalid operator")
		}
	case ':':
		if l.match('=') {
			l.emit(COLON_EQUAL)
		} else {
			l.error("invalid operator")
		}
	case '"':
		l.lexString()
	default:
		if isDigit(ch) {
			l.lexNumber()
		} else if isAlpha(ch) {
			l.lexIdentifier()
		} else {
			l.error("unexpected character %U %q", ch, ch)
		}
	}
}

func (l *Lexer) lexIdentifier() {
	for isIdentifier(l.peek()) {
		l.advance()
	}
	word := l.source[l.start:l.current]
	if typ, ok := keywords[word]; ok {
		l.emit(typ)
	} else {
		l.emitLiteral(IDENTIFIER, word)
	}
}

func (l *Lexer) lexNumber() {
	// match a run of digits
	for isDigit(l.peek()) {
		l.advance()
	}
	if l.peek() == '.' && isDigit(l.peekNext()) {
		l.advance() // consume '.'
		for isDigit(l.peek()) {
			l.advance()
		}
	}
	num, err := strconv.ParseFloat(l.source[l.start:l.current], 64)
	if err != nil {
		l.error(err.Error())
	}
	l.emitLiteral(NUMBER, num)
}

func (l *Lexer) lexString() {
	// we've already ate one '"' token.
	var buf bytes.Buffer
	esc := false
	for !l.isAtEnd() {
		ch := l.advance()
		if l.stop {
			// this will be put into .errors
			return
		}
		if !esc {
			switch ch {
			case '\\':
				esc = true
			case '"':
				l.emitLiteral(STRING, buf.String())
				return
			case 0:
				fallthrough
			case '\r':
				fallthrough
			case '\n':
				l.error("unexpected char in string literal: %U %q", ch, ch)
			default:
				buf.WriteRune(ch)
			}
		} else {
			switch ch {
			case '\\':
				buf.WriteRune('\\')
			case '"':
				buf.WriteRune('"')
			case '0':
				buf.WriteRune(0)
			case 'r':
				buf.WriteRune('\r')
			case 'n':
				buf.WriteRune('\n')
			default:
				l.error("invalid escape in string literal: %q", "\\"+string(ch))
			}
			esc = false
		}
	}
	// if we've reached here, then there was no terminating "
	l.error("unterminated string")
}

// ignore ignores the currently scanned lexeme
func (l *Lexer) ignore() {
	l.start = l.current
	l.startLn = l.line
	l.startCol = l.column
}

func (l *Lexer) emit(typ TokenType) { l.emitLiteral(typ, nil) }
func (l *Lexer) emitLiteral(typ TokenType, lit interface{}) {
	l.Tokens = append(l.Tokens, Token{
		Type:    typ,
		Lexeme:  l.source[l.start:l.current],
		Literal: lit,
		Line:    l.startLn,
		Column:  l.startCol,
	})
	l.start = l.current
	l.startLn = l.line
	l.startCol = l.column
}

func (l *Lexer) error(s string, args ...interface{}) {
	l.Errors = append(l.Errors, Error{
		Filename: l.Filename,
		Line:     l.line,
		Column:   l.column,
		Message:  fmt.Sprintf(s, args...),
	})
}

func isWhiteSpace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' }
func isIdentifier(ch rune) bool { return isAlpha(ch) || isDigit(ch) }
func isAlpha(ch rune) bool      { return ch == '_' || ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') }
func isDigit(ch rune) bool      { return '0' <= ch && ch <= '9' }
