package resolver_test

import (
	"testing"
	"toe/lexer"
	"toe/parser"
	"toe/resolver"
)

func TestResolver(t *testing.T) {
	input := `
let f = fn(x) { f(x + 1); return this.x; };
f(1);
`
	module := lexAndParse(t, input)
	if module == nil {
		return
	}
	r := resolver.New(module)
	r.Resolve()
	if !noErrors(t, "resolver", r.Errors) {
		return
	}
	f_inside_func := module.Stmts[0].(*parser.Let).Value.(*parser.Function).Body.Stmts[0].(*parser.ExprStmt).Expr.(*parser.Call).Callee.(*parser.Identifier)
	f_outside_func := module.Stmts[1].(*parser.ExprStmt).Expr.(*parser.Call).Callee.(*parser.Identifier)
	if f_inside_func.Loc != 2 {
		t.Errorf("expected f inside to be 2, got=%d", f_inside_func.Loc)
	}
	if f_outside_func.Loc != 0 {
		t.Errorf("expected f outside to be 0, got=%d", f_outside_func.Loc)
	}
}

// utils

func lexAndParse(t *testing.T, input string) *parser.Module {
	fn := ""
	l := lexer.New(fn, input)
	l.ScanTokens()
	if !noErrors(t, "lexer", l.Errors) {
		return nil
	}
	p := parser.New(fn, l.Tokens)
	module := p.Parse()
	if !noErrors(t, "parser", p.Errors) {
		return nil
	}
	return module
}

func noErrors(t *testing.T, src string, errors []error) bool {
	if len(errors) != 0 {
		t.Errorf("got %s errors:\n", src)
		for _, x := range errors {
			t.Errorf("%s\n", x)
		}
		return false
	}
	return true
}
