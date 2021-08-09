package eval

import (
	"bytes"
	"fmt"
	"strconv"
	"toe/parser"
)

//go:generate stringer -type=ValueType

type ValueType uint8

const (
	_ = ValueType(iota)
	// Real values
	VT_NIL
	VT_BOOLEAN
	VT_NUMBER
	VT_STRING
	VT_FUNCTION
	VT_OBJECT
	// Runtime Control
	VT_BREAK
	VT_CONTINUE
	VT_RETURN
	VT_ERROR
)

type Value interface {
	Type() ValueType
}

// =======
// Objects
// =======

type Nil struct{}
type Boolean bool
type Number float64
type String string

type Object struct {
	frozen bool // whether this object is mutable
	proto  Value
	slots  map[string]Value
}

func newObject(proto Value) *Object {
	return &Object{
		frozen: false,
		proto:  proto,
		slots:  map[string]Value{},
	}
}

type Function struct {
	*Object
	node    *parser.Function
	this    Value
	closure *environment
	module  *parser.Module
	ctx     string // precomputed context
}

func newFunction(module *parser.Module, node *parser.Function, this Value, env *environment) *Function {
	f := &Function{
		Object:  newObject(nil),
		module:  module,
		node:    node,
		this:    this,
		closure: env,
	}
	f.ctx = f.String()
	return f
}

func (v Nil) Type() ValueType       { return VT_NIL }
func (v Boolean) Type() ValueType   { return VT_BOOLEAN }
func (v Number) Type() ValueType    { return VT_NUMBER }
func (v String) Type() ValueType    { return VT_STRING }
func (v *Object) Type() ValueType   { return VT_OBJECT }
func (v *Function) Type() ValueType { return VT_FUNCTION }

func (v Nil) String() string { return "nil" }
func (v Boolean) String() string {
	if v {
		return "true"
	}
	return "false"
}
func (v Number) String() string { return strconv.FormatFloat(float64(v), 'g', -1, 64) }
func (v String) String() string { return string(v) }
func (v *Function) String() string {
	name := v.node.Name
	if name != "" {
		name = " " + name
	}
	isBound := ""
	if v.this != nil {
		isBound = " bound"
	}
	return fmt.Sprintf("[Function%s%s]", isBound, name)
}
func (v *Object) String() string {
	isFrozen := ""
	if v.frozen {
		isFrozen = "frozen "
	}
	return fmt.Sprintf("[Object %s%p]", isFrozen, v)
}

func (v String) Inspect() string { return fmt.Sprintf("%q", string(v)) }

// ==========
// Singletons
// ==========

var (
	NIL      = Nil{}
	TRUE     = Boolean(true)
	FALSE    = Boolean(false)
	BREAK    = Break{}
	CONTINUE = Continue{}
)

// ===============
// Runtime Control
// ===============

type Break struct{}
type Continue struct{}
type Return struct{ value Value }

type context struct {
	fn  string // filename
	ln  int    // line no
	col int    // col
	ctx string // e.g. [Module] or [Function ...]
}

type Error struct {
	reason Value
	stack  []context
}

func newError(reason Value) *Error {
	return &Error{
		reason: reason,
		stack:  []context{},
	}
}

func (e *Error) String() string {
	var buf bytes.Buffer
	buf.WriteString("Error: ")
	buf.WriteString(e.reason.(Stringer).String())
	buf.WriteString("\n")
	for i := len(e.stack) - 1; i >= 0; i-- {
		ctx := e.stack[i]
		buf.WriteString(fmt.Sprintf("  at %s:%d:%d: %s", ctx.fn, ctx.ln, ctx.col, ctx.ctx))
		if i != 0 {
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

func (v Break) Type() ValueType    { return VT_BREAK }
func (v Continue) Type() ValueType { return VT_CONTINUE }
func (v Return) Type() ValueType   { return VT_RETURN }
func (v Error) Type() ValueType    { return VT_ERROR }