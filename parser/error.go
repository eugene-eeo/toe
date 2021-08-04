package parser

import (
	"fmt"
	"toe/lexer"
)

// Represents a parsing error. We use this internally to signal
// that we cannot continue parsing some expression/statement --
// as opposed to minor errors like assigning to a function.
type ParserError struct {
	Filename string
	Token    lexer.Token
	Message  string
}

func (p *Parser) error(s string, args ...interface{}) {
	err := ParserError{
		Filename: p.filename,
		Token:    p.previous(),
		Message:  fmt.Sprintf(s, args...),
	}
	p.Errors = append(p.Errors, err)
	panic(err)
}

func (p *Parser) expect(typ lexer.TokenType, s string, args ...interface{}) {
	if !p.match(typ) {
		p.error(s, args...)
	}
}

// synchronize synchronizes the parser by discarding tokens
// until we reach a token which starts a statement. This means
// that cascading errors are discarded, and we still report as
// many errors as possible.
func (p *Parser) synchronize() {
	p.consume()
	for (!p.isAtEnd()) {
		if p.previous().Type == lexer.SEMICOLON {
			return
		}
		switch p.peek().Type {
		case lexer.IF:
		case lexer.RETURN:
		case lexer.FOR:
		case lexer.WHILE:
			return
		}
		p.consume()
	}
}
