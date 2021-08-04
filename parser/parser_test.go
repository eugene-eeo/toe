package parser_test

import (
	"testing"
	"toe/lexer"
	"toe/parser"
)

func TestParser(t *testing.T) {
	l := lexer.New("", "(a + 1.5) * true")
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
}
