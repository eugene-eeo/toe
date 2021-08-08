package eval

import (
	"toe/lexer"
	"toe/parser"
	"toe/resolver"
)

type InteractiveContext struct {
	ctx *Context
	res *resolver.Resolver
}

func NewInteractiveContext() *InteractiveContext {
	module := &parser.Module{Filename: "<stdin>"}
	res := resolver.New(module)
	ctx := NewContext(res.Locs)
	ctx.Env, _ = ctx.NewModuleEnv("<stdin>")
	ctx.pushFunc("<module>")
	return &InteractiveContext{
		ctx, res,
	}
}

func (ic *InteractiveContext) Inspect(v Value) string {
	return ic.ctx.Inspect(v)
}

func (ic *InteractiveContext) Run(input string) (Value, []error) {
	l := lexer.New("<stdin>", input)
	l.ScanTokens()
	if len(l.Errors) != 0 {
		return nil, l.Errors
	}
	p := parser.New("<stdin>", l.Tokens)
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
		rv = ic.ctx.Eval(stmt)
		if isError(rv) {
			return rv, nil
		}
	}
	return rv, nil
}
