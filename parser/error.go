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

func (pe ParserError) Error() string { return pe.String() }
func (pe ParserError) String() string {
	return fmt.Sprintf("%s:%d:%d: %s", pe.Filename, pe.Token.Line, pe.Token.Column, pe.Message)
}

func (p *Parser) error(s string, args ...interface{}) error {
	err := ParserError{
		Filename: p.filename,
		Token:    p.previous(),
		Message:  fmt.Sprintf(s, args...),
	}
	p.Errors = append(p.Errors, err)
	return err
	// panic(err)
}

func (p *Parser) expect(typ lexer.TokenType, s string, args ...interface{}) lexer.Token {
	if !p.match(typ) {
		panic(p.error(s, args...))
	}
	return p.previous()
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
		case lexer.LET:
		case lexer.IF:
		case lexer.RETURN:
		case lexer.FOR:
		case lexer.WHILE:
			return
		}
		p.consume()
	}
}
