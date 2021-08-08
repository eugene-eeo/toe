package eval

// This package contains the runtime representations of toe values.
// All values implement the Value interface, which has a prototype
// field (for values controlling the runtime behaviour, such as Break
// and Return we simply ignore their prototypes).

import (
	"bytes"
	"fmt"
	"toe/parser"
)

type ValueType uint8

//go:generate stringer -type=ValueType

const (
	_ = ValueType(iota)
	NIL_TYPE
	STRING
	BOOLEAN
	NUMBER
	OBJECT
	ARRAY
	BUILTIN
	FUNCTION
	// Meta values
	ERROR
	BREAK
	CONTINUE
	RETURN
)

type Value interface {
	Type() ValueType
}

// -----------------
// Boxed value types
// -----------------

type Nil struct{}
type Boolean bool
type Number float64
type String string

type Object struct {
	props map[string]Value
	proto Value
}

func (v Nil) Type() ValueType     { return NIL_TYPE }
func (v Boolean) Type() ValueType { return BOOLEAN }
func (v Number) Type() ValueType  { return NUMBER }
func (v String) Type() ValueType  { return STRING }
func (v *Object) Type() ValueType { return OBJECT }

func newObject(object Value) *Object {
	return &Object{
		proto: object,
		props: map[string]Value{},
	}
}

type Array struct{ arr []Value }

func (v *Array) Type() ValueType { return ARRAY }

type Builtin struct {
	this Value // x.fn() --> this == x
	fn   func(ctx *Context, this Value, args []Value) Value
}

func (v *Builtin) Type() ValueType { return BUILTIN }

func (v *Builtin) Bind(this Value) *Builtin {
	if v.this == nil {
		return &Builtin{this, v.fn}
	}
	return v
}

func (v *Builtin) Call(ctx *Context, args []Value) Value {
	this := v.this
	if this == nil {
		this = NIL
	}
	ctx.pushFunc("builtin function")
	defer ctx.popFunc()
	return v.fn(ctx, this, args)
}

type Function struct {
	this    Value
	closure *Environment
	node    *parser.Function
}

func (v *Function) Type() ValueType { return FUNCTION }

func (v *Function) Bind(this Value) *Function {
	if v.this == nil {
		return &Function{this, v.closure, v.node}
	}
	return v
}

func (v *Function) Call(ctx *Context, args []Value) Value {
	// remember the current environment
	old_env := ctx.Env
	old_this := ctx.this
	// create a new environment with bound parameters
	ctx.Env = newEnvironment(v.closure.filename, v.closure)
	ctx.pushFunc("function")
	this := v.this
	if this == nil {
		this = NIL
	}
	ctx.this = this
	ctx.Env.Define("this", this)
	for i, param := range v.node.Params {
		value := Value(NIL)
		if len(args) > i {
			value = args[i]
		}
		ctx.Env.Define(param.Lexeme, value)
	}
	rv := ctx.Eval(v.node.Body)
	// restore the old environment
	ctx.this = old_this
	ctx.Env = old_env
	ctx.popFunc()
	return rv
}

// ----------------------
// Runtime Control Values
// ----------------------

// Error is a currently propagating error (exception).
type Error struct {
	ctx    *Context
	Reason Value
	Trace  []TraceEntry
}

func (e *Error) String() string {
	var buf bytes.Buffer
	buf.WriteString("Call stack:\n")
	for i := len(e.Trace) - 1; i >= 0; i-- {
		buf.WriteString("  ")
		buf.WriteString(e.Trace[i].String())
		buf.WriteString("\n")
	}
	buf.WriteString(e.ctx.Inspect(e.Reason))
	return buf.String()
}

// Not a value -- but entries on the stack trace.
type TraceEntry struct {
	Filename string
	Line     int
	Column   int
	Context  string
}

func (te *TraceEntry) String() string {
	return fmt.Sprintf("%s:%d:%d: in %s", te.Filename, te.Line, te.Column, te.Context)
}

// Break signals that a loop is being broken from.
type Break struct{}
type Continue struct{}

// Return is a return-value -- needs to be unwrapped.
type Return struct{ value Value }

func (e *Error) Type() ValueType    { return ERROR }
func (b *Break) Type() ValueType    { return BREAK }
func (c *Continue) Type() ValueType { return CONTINUE }
func (r *Return) Type() ValueType   { return RETURN }
