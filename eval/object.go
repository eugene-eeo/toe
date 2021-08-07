package eval

// This package contains the runtime representations of toe values.
// All values implement the Value interface, which has a prototype
// field (for values controlling the runtime behaviour, such as Break
// and Return we simply ignore their prototypes).

import (
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
type Boolean struct{ value bool }
type Number struct{ value float64 }
type String struct{ value string }

type Object struct {
	props map[string]Value
	proto Value
}

func (v *Nil) Type() ValueType     { return NIL_TYPE }
func (v *Boolean) Type() ValueType { return BOOLEAN }
func (v *Number) Type() ValueType  { return NUMBER }
func (v *String) Type() ValueType  { return STRING }
func (v *Object) Type() ValueType  { return OBJECT }

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
	return v.fn(ctx, v.this, args)
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
	// create a new environment with bound parameters
	ctx.Env = newEnvironment(v.closure)
	this := v.this
	if this == nil {
		this = NIL
	}
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
	ctx.Env = old_env
	return rv
}

// ----------------------
// Runtime Control Values
// ----------------------

// Error is a currently propagating error (exception).
type Error struct{ Reason Value }

// Break signals that a loop is being broken from.
type Break struct{}
type Continue struct{}

// Return is a return-value -- needs to be unwrapped.
type Return struct{ value Value }

func (e *Error) Type() ValueType    { return ERROR }
func (b *Break) Type() ValueType    { return BREAK }
func (c *Continue) Type() ValueType { return CONTINUE }
func (r *Return) Type() ValueType   { return RETURN }
