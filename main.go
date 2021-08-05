package main

// implements a toe repl

import (
	"fmt"
	"os"
	"bufio"
	"toe/eval"
	"toe/lexer"
	"toe/parser"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	ctx := eval.NewContext()
	for {
		fmt.Printf("> ")
		if !scanner.Scan() {
			fmt.Println()
			return
		}
		line := scanner.Text()
		lexer := lexer.New("<stdin>", line)
		lexer.ScanTokens()
		if len(lexer.Errors) != 0 {
			for _, err := range lexer.Errors {
				fmt.Printf("%s\n", err.String())
			}
			continue
		}
		parser := parser.New("<stdin>", lexer.Tokens)
		module := parser.Parse()
		if len(parser.Errors) != 0 {
			for _, err := range parser.Errors {
				fmt.Printf("%s\n", err)
			}
			continue
		}
		var rv eval.Value = nil
		for _, stmt := range module.Stmts {
			rv = ctx.Eval(stmt)
		}
		if rv != nil {
			// if rv.(*parser.Error) {
			// 	fmt.Printf("%#v\n", rv)
			// }
			fmt.Printf("%#v\n", rv)
		}
	}
}
