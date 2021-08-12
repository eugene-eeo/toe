package eval

import (
	"bytes"
	"fmt"
	"toe/parser"
)

//go:generate stringer -type=ValueType

type ValueType uint8

const (
	_ = ValueType(iota)
	// Wildcard type for operator lookup
	VT_ANY
	// Real values
	VT_NIL
	VT_BOOLEAN
	VT_NUMBER
	VT_STRING
	VT_FUNCTION
	VT_OBJECT
	VT_ARRAY
	VT_HASH
	VT_BUILTIN
	// Runtime Control
	VT_BREAK
	VT_CONTINUE
	VT_RETURN
	VT_ERROR
	// Hashtable -- tombstones
	VT_TOMBSTONE
)

type Value interface {
	Type() ValueType
}

// =======
// Objects
// =======
//
// When we speak of objects, they refer to the `real' values in the runtime,
// not control values or tombstones. All `real' values need to additionally
// implement the Stringer protocol.

type Nil struct{}
type Boolean bool
type Number float64
type String string

type Object struct {
	proto Value
	slots map[string]Value
}

func newObject(proto Value) *Object {
	return &Object{
		proto: proto,
		slots: map[string]Value{},
	}
}

type Function struct {
	*Object
	node     *parser.Function
	closure  *environment
	filename string
	this     Value
}

func newFunction(filename string, node *parser.Function, env *environment) *Function {
	return &Function{
		Object:   newObject(nil),
		filename: filename,
		node:     node,
		closure:  env,
	}
}

type Array struct {
	*Object
	values []Value
}

func newArray(values []Value) *Array {
	return &Array{
		Object: newObject(nil),
		values: values,
	}
}

type Hash struct {
	*Object
	table *hashTable
}

func newHash(ctx *Context) *Hash {
	return &Hash{
		Object: newObject(nil),
		table:  newHashTable(ctx),
	}
}

// Builtin represents a built-in function
type Builtin struct {
	*Object
	name string
	this Value
	call func(ctx *Context, this Value, args []Value) Value
}

func (v Nil) Type() ValueType       { return VT_NIL }
func (v Boolean) Type() ValueType   { return VT_BOOLEAN }
func (v Number) Type() ValueType    { return VT_NUMBER }
func (v String) Type() ValueType    { return VT_STRING }
func (v *Object) Type() ValueType   { return VT_OBJECT }
func (v *Function) Type() ValueType { return VT_FUNCTION }
func (v *Array) Type() ValueType    { return VT_ARRAY }
func (v *Hash) Type() ValueType     { return VT_HASH }
func (v *Builtin) Type() ValueType  { return VT_BUILTIN }

// ==========
// Singletons
// ==========

var (
	NIL       = Nil{}
	TRUE      = Boolean(true)
	FALSE     = Boolean(false)
	BREAK     = Break{}
	CONTINUE  = Continue{}
	TOMBSTONE = Value(tombstone{})
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
	for i, ctx := range e.stack {
		buf.WriteString(fmt.Sprintf("  at %s:%d:%d: %s", ctx.fn, ctx.ln, ctx.col, ctx.ctx))
		if i != len(e.stack)-1 {
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

func (v Break) Type() ValueType    { return VT_BREAK }
func (v Continue) Type() ValueType { return VT_CONTINUE }
func (v Return) Type() ValueType   { return VT_RETURN }
func (v Error) Type() ValueType    { return VT_ERROR }

// ====================
// Hash table tombstone
// ====================

type tombstone struct{}

func (v tombstone) Type() ValueType { return VT_TOMBSTONE }
