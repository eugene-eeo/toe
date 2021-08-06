package resolver_test

import (
	"toe/lexer"
	"toe/parser"
	"toe/resolver"
	"testing"
)

func TestResolver(t *testing.T) {
	input := `
let a = 1;
{
	let b = 2;
	let c = 3;
	a = c;
}
c;
continue;
`
	fn := ""
	l := lexer.New(fn, input)
	l.ScanTokens()
	if len(l.Errors) != 0 {
		t.Error("got lexing errors:")
		for _, x := range l.Errors {
			t.Errorf("%s\n", x)
		}
		return
	}
	p := parser.New(fn, l.Tokens)
	module := p.Parse()
	if len(p.Errors) != 0 {
		t.Error("got parsing errors:")
		for _, x := range p.Errors {
			t.Errorf("%s\n", x)
		}
		return
	}
	r := resolver.New(module)
	r.Resolve()
	t.Log("--------")
	t.Log("resolution errors:")
	for _, x := range r.Errors {
		t.Logf("%s\n", x)
	}
	t.Log("--------")
	t.Log(r.Locs)
}
