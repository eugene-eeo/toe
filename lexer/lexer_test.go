package lexer_test

import (
	"testing"
	"toe/lexer"
)

func TestLexer(t *testing.T) {
	lex := lexer.New("", `
let Animal = Object.clone(nil);
let dog = PetDog.new("阿福");
21.50 == 2.10;
true == false == fn() { return 2 }`)
	lex.ScanTokens()
	if len(lex.Errors) != 0 {
		t.Errorf("failed: expected no errors, got:")
		for _, x := range lex.Errors {
			t.Log(x)
		}
	}
	t.Log(lex.Tokens)
}

func TestLexerBad(t *testing.T) {
	badInputs := []string{
		"\"ab\n\" def ghi",
		"def | holy shit",
		"abc & adhkfsai",
		"\"abraca\xc3\x28 dabra\"",
		"\xc3\x28",
		"abc def \xf0\x28\x8c\xbc uu \xc3\x28 omg",
		"abc def || omg &| abrac",
	}
	for i, input := range badInputs {
		lex := lexer.New("<test>", input)
		lex.ScanTokens()
		if len(lex.Errors) == 0 {
			t.Errorf("tests[%d] (%q) failed", i, input)
			t.Errorf("expected errors, got none")
		}
		for _, x := range lex.Errors {
			t.Logf("%s\n", x)
		}
	}
}
