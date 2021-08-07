package parser_test

import (
	"testing"
	"toe/lexer"
	"toe/parser"
)

func TestParserValid(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"abcdef = 2;", "(abcdef = 2);"},
		{"a + b + c;", "((a + b) + c);"},
		{"a + b + (c = 7);", "((a + b) + (c = 7));"},
		{"a + b * c;", "(a + (b * c));"},
		{"a + b >= c == true;", "(((a + b) >= c) == true);"},
		{"a + !b or x;", "((a + (!b)) or x);"},
		{"a and b or c;", "((a and b) or c);"},
		{"a + -b * c / d;", "(a + (((-b) * c) / d));"},
		{"a / (c - f) / d + e;", "(((a / (c - f)) / d) + e);"},
		{"a = b = c;", "(a = (b = c));"},
		{"for (x : a) true;", "for (x : a) true;"},
		{"for (x : a) { let x = 2; b >= 10; }", "for (x : a) {let x = 2;(b >= 10);}"},
		{"while (true) for (x : a) true;", "while (true) for (x : a) true;"},
		{"if (true) { x = 1; }", "if (true) {(x = 1);}"},
		{"if (true) { x = 1; } else nil;", "if (true) {(x = 1);} else nil;"},
		{"for (x : a) if (x == 1) break;", "for (x : a) if ((x == 1)) break;"},
		{"while (true) continue;", "while (true) continue;"},
		{"true.x;", "(true.x);"},
		{"true.x.y == 2.u;", "(((true.x).y) == (2.u));"},
		{"true.x.y != 2.u.z or 2;", "((((true.x).y) != ((2.u).z)) or 2);"},
		{"x.true.false.nil;", "(((x.true).false).nil);"},
		{"a.b = true;", "(a.b = true);"},
		{"x.true.u = 1;", "((x.true).u = 1);"},
		{"a->b = true;", "(a->b = true);"},
		{"fn(a) { true; };", "fn(a){true;};"},
		{"fn(a,) { true; };", "fn(a){true;};"},
		{"fn(a,b) { true; };", "fn(a, b){true;};"},
		{"fn(a) { return 1; };", "fn(a){return 1;};"},
		{"fn(a) { return; };", "fn(a){return;};"},
		{"isEven(n);", "(isEven(n));"},
		{"isEven(1,);", "(isEven(1));"},
		{"isEven(1,2,3,4,5,);", "(isEven(1, 2, 3, 4, 5));"},
		{"isEven.call(n)->this.that;", "((((isEven.call)(n))->this).that);"},
	}
	for i, test := range tests {
		var tokens []lexer.Token
		if !checkLexerErrors(t, test.input, &tokens) {
			t.Errorf("tests[%d] (%q) failed", i, test.input)
			continue
		}
		p := parser.New("", tokens)
		expr := p.Parse()
		if len(p.Errors) != 0 {
			t.Errorf("tests[%d] (%q)", i, test.input)
			t.Error("parser errors:")
			for _, err := range p.Errors {
				t.Errorf("%s\n", err)
			}
			continue
		}
		if expr.String() != test.expected {
			t.Errorf("tests[%d] (%q)", i, test.input)
			t.Errorf("expected=%q, got=%q", test.expected, expr.String())
			continue
		}
	}
}

func TestParserInvalid(t *testing.T) {
	tests := []struct{
		input string
		numErrs int
	}{
		{"abcdef = 2", 1},
		{"if (x) { if (x) wtf! omg }", 2}, // panic mode should help here
		{"u + v + (x + 1; (x) then(); if (x) then()", 3},
		{"!!;", 1},
		{"=;", 1}, // this was known to make the parser panic
		{";", 1},
		{"1 = 2; x", 2}, // should continue parsing
		{"x.1;", 1},
		{"fn(x,,){}", 1},
		{"fn(,){}", 1},
		{"z(1,2,3,,);", 1},
	}
	for i, test := range tests {
		var tokens []lexer.Token
		if !checkLexerErrors(t, test.input, &tokens) {
			t.Errorf("tests[%d] (%q) failed", i, test.input)
			continue
		}
		p := parser.New("", tokens)
		p.Parse()
		t.Logf("tests[%d]: %+v\n", i, p.Errors)
		if len(p.Errors) != test.numErrs {
			t.Errorf("tests[%d] (%q)", i, test.input)
			t.Errorf("expected=%d errors, got=%d", test.numErrs, len(p.Errors))
			t.Errorf("%+v\n", p.Errors)
		}
	}
}

func checkLexerErrors(t *testing.T, input string, out *[]lexer.Token) bool {
	l := lexer.New("", input)
	l.ScanTokens()
	if len(l.Errors) != 0 {
		t.Error("lexer errors:")
		for _, err := range l.Errors {
			t.Errorf("%s\n", err)
		}
		return false
	}
	*out = l.Tokens
	return true
}
