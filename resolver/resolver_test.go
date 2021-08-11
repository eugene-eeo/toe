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
let a = 1;
if (1) {
	let g = nil;
	g.func_name = fn() {
		return a;
	};
	let a = 2;
}
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
	f_func := module.Stmts[0].(*parser.Let).Value.(*parser.Function)
	g_set := module.Stmts[3].(*parser.If).Then.(*parser.Block).Stmts[1].(*parser.ExprStmt).Expr.(*parser.Set)
	g_ident := g_set.Object.(*parser.Identifier)
	g_func := g_set.Right.(*parser.Function)
	f_inside_f := f_func.Body.Stmts[0].(*parser.ExprStmt).Expr.(*parser.Call).Callee.(*parser.Identifier)
	f_call := module.Stmts[1].(*parser.ExprStmt).Expr.(*parser.Call).Callee.(*parser.Identifier)
	a_inside_g := g_func.Body.Stmts[0].(*parser.Return).Expr.(*parser.Identifier)
	if f_func.Name != "f" {
		t.Errorf("expected f_func.Name='f', got=%q", f_func.Name)
	}
	if f_inside_f.Loc != 2 {
		t.Errorf("expected f inside f to be 2, got=%d", f_inside_f.Loc)
	}
	if f_call.Loc != 0 {
		t.Errorf("expected f outside to be 0, got=%d", f_call.Loc)
	}
	if g_ident.Loc != 0 {
		t.Errorf("expected g to be 0, got=%d", g_ident.Loc)
	}
	if g_func.Name != "func_name" {
		t.Errorf("expected g.func_name.Name=\"func_name\" got=%q", g_func.Name)
	}
	if a_inside_g.Loc != 3 {
		t.Errorf("expected a inside g be 3, got=%d", a_inside_g.Loc)
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
