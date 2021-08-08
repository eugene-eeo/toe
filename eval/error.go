package eval

import "toe/lexer"

// This package implements error formatting and reporting mechanisms.
// The protocol around adding errors is:
//
//   1. Every time we call a function, we need to do ctx.pushFunc("...")
//
//   2. Returning from a function similarly does a ctx.popFunc()
//
//   3. Whenever a user-generated error is produced, we add the
//      error location to the trace -- this means that, we take
//      the current function we're in (via ctx.currFunc()) and put
//      information about _where_ in the function the error happened.

func (ctx *Context) popFunc()          { ctx.funcs = ctx.funcs[:len(ctx.funcs)-1] }
func (ctx *Context) pushFunc(s string) { ctx.funcs = append(ctx.funcs, s) }
func (ctx *Context) currFunc() string  { return ctx.funcs[len(ctx.funcs)-1] }

func (ctx *Context) err(reason Value) *Error {
	return &Error{
		ctx: ctx,
		Reason: reason,
		Trace:  []TraceEntry{},
	}
}

func (e *Error) addContext(tok lexer.Token) {
	e.Trace = append(e.Trace, TraceEntry{
		Filename: e.ctx.Env.filename, // the file name in the currently executing env.
		Line:     tok.Line,
		Column:   tok.Column,
		Context:  e.ctx.currFunc(),
	})
}

// // setupErrors adds the error types.
// func setupErrors(ctx *Context) {
// 	ctx.globals.Error = newObject(ctx.globals.Object)
// 	ctx.globals.Error.props["ReferenceError"] = newObject(ctx.globals.Error)
// 	ctx.globals.Error.props["TypeError"] = newObject(ctx.globals.Error)
// }
