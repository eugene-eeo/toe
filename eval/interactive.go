package eval

import (
	"toe/lexer"
	"toe/parser"
	"toe/resolver"
)

type InteractiveContext struct {
	Filename string
	ctx      *Context
	res      *resolver.Resolver
}

func NewInteractiveContext() *InteractiveContext {
	fn := "<stdin>"
	module := &parser.Module{Filename: fn}
	res := resolver.New(module)
	ctx := NewContext()
	ctx.globals.addToResolver(res)
	ctx.pushEnv()
	ctx.globals.addToEnv(ctx.env)
	ctx.pushFunc(&moduleCse{fn})
	return &InteractiveContext{fn, ctx, res}
}

func (ic *InteractiveContext) Inspect(v Value) (string, *Error) {
	rv := ic.ctx.call_method(v, "inspect", nil)
	if isError(rv) {
		return "", rv.(*Error)
	}
	str := ic.ctx.getSpecial(rv, VT_STRING)
	if str == nil {
		return "", newError(ic.ctx, String("inspect returned a non-string"))
	}
	return string(rv.(String)), nil
}

func (ic *InteractiveContext) Run(input string) (Value, []error) {
	l := lexer.New(ic.Filename, input)
	l.ScanTokens()
	if len(l.Errors) != 0 {
		return nil, l.Errors
	}
	p := parser.New(ic.Filename, l.Tokens)
	module := p.Parse()
	if len(p.Errors) != 0 {
		return nil, p.Errors
	}
	for _, stmt := range module.Stmts {
		ic.res.ResolveOne(stmt)
		if len(ic.res.Errors) != 0 {
			og := ic.res.Errors
			ic.res.Errors = []error{}
			return nil, og
		}
	}
	rv := Value(nil)
	// Still no errors? we can run it.
	for _, stmt := range module.Stmts {
		rv = ic.ctx.EvalStmt(stmt)
		if isError(rv) {
			return rv, nil
		}
	}
	return rv, nil
}
