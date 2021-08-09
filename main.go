package main

// implements a toe repl

import (
	"fmt"
	"os"
	"strings"
	"toe/eval"
	"github.com/chzyer/readline"
)

var VERSION string
var LOGO = `
  __                 |
 |  |_.-----.-----.  | toe repl
 |   _|  _  |  -__|  | version: $VERSION
 |____|_____|_____|  |
`

func sliceVersion(v string) string {
	m := 10
	if len(v) < 10 {
		m = len(v)
	}
	return v[0:m]
}

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
	fmt.Println(strings.Replace(LOGO, "$VERSION", sliceVersion(VERSION), 1))
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
			} else if u.Type() == eval.VT_ERROR {
				fmt.Fprintln(os.Stderr, u.(*eval.Error).String())
			} else {
				fmt.Println(ctx.Inspect(u))
			}
		}
	}
}
