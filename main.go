package main

// implements a toe repl

import (
	"fmt"
	"os"
	"toe/eval"
	"github.com/chzyer/readline"
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
	rl, err := readline.New("> ")
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	ctx := eval.NewInteractiveContext()
	for {
		line, err := rl.Readline()
		if err != nil {
			break
		}
		u, errs := ctx.Run(line)
		if errs != nil {
			reportErrors(errs)
		} else {
			if u == nil {
				continue
			} else if u.Type() == eval.ERROR {
				fmt.Fprintln(os.Stderr, u.(*eval.Error).String())
			} else {
				fmt.Println(ctx.Inspect(u))
			}
		}
	}
}
