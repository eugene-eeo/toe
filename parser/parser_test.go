package parser_test

import (
	"testing"
	"toe/lexer"
	"toe/parser"
)

func TestParser(t *testing.T) {
	l := lexer.New("", "abcdef = 2")
	l.ScanTokens()
	if len(l.Errors) != 0 {
		t.Error("got errors")
		for _, err := range l.Errors {
			t.Error(err.String())
		}
		return
	}
	p := parser.New("", l.Tokens)
	t.Logf("%+v\n", p.ParseExpression())
	if len(p.Errors) != 0 {
		t.Error("got errors")
		for _, err := range p.Errors {
			t.Error(err.String())
		}
		return
	}
}
