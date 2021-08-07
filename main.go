package main

// implements a toe repl

import (
	"bufio"
	"fmt"
	"os"
	"toe/eval"
	"toe/lexer"
	"toe/parser"
	"toe/resolver"
)

var LOGO = `
_|_ _  _
 |_(_)(/_
 `

func reportErrors(errors []error) bool {
	if len(errors) == 0 {
		return false
	}
	for _, err := range errors {
		fmt.Fprintf(os.Stderr, "%s\n", err)
	}
	return true
}

func main() {
	fmt.Println(LOGO)
	scanner := bufio.NewScanner(os.Stdin)
	res := resolver.New(&parser.Module{Filename: "<stdin>"})
	ctx := eval.NewContext(res.Locs)
	env, _ := ctx.NewModuleEnv("<stdin>")
	ctx.Env = env
	for {
		fmt.Printf("> ")
		if !scanner.Scan() {
			fmt.Println()
			return
		}
		line := scanner.Text()
		hasErrors := false

		lexer := lexer.New("<stdin>", line)
		lexer.ScanTokens()
		hasErrors = reportErrors(lexer.Errors)

		if hasErrors {
			continue
		}

		parser := parser.New("<stdin>", lexer.Tokens)
		module := parser.Parse()
		hasErrors = reportErrors(parser.Errors)
		if hasErrors {
			continue
		}

		for _, stmt := range module.Stmts {
			res.ResolveOne(stmt)
			if hasErrors = reportErrors(res.Errors); hasErrors {
				res.Errors = []error{}
				break
			}
		}
		if hasErrors {
			continue
		}
		// all is well -- can execute!
		for _, stmt := range module.Stmts {
			rv := ctx.Eval(stmt)
			if rv != nil {
				// if rv.(*parser.Error) {
				// 	fmt.Printf("%#v\n", rv)
				// }
				fmt.Printf("%#v\n", rv)
			}
		}
	}
}
