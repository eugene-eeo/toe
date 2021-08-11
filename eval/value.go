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
	this     Value
	closure  *environment
	filename string
}

func newFunction(filename string, node *parser.Function, this Value, env *environment) *Function {
	return &Function{
		Object:   newObject(nil),
		filename: filename,
		node:     node,
		this:     this,
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
	return fmt.Sprintf("[Object %p]", v)
}

func (v *Array) String() string {
	var buf bytes.Buffer
	last_idx := len(v.values) - 1
	buf.WriteString("[")
	for i, x := range v.values {
		buf.WriteString(inspect(x))
		if i != last_idx {
			buf.WriteString(", ")
		}
	}
	buf.WriteString("]")
	return buf.String()
}

func (v *Hash) String() string {
	var buf bytes.Buffer
	buf.WriteString("{")
	j := uint64(0)
	size := v.table.size()
	for i := 0; i < len(v.table.entries); i++ {
		ref := &v.table.entries[i]
		if ref.hasValue() {
			j++
			buf.WriteString(inspect(*ref.key))
			buf.WriteString(": ")
			buf.WriteString(inspect(*ref.value))
			if j < size {
				buf.WriteString(", ")
			}
		}
	}
	buf.WriteString("}")
	return buf.String()
}

func (v *Builtin) String() string {
	isBound := ""
	if v.this != nil {
		isBound = " bound"
	}
	return fmt.Sprintf("[Function%s %s]", isBound, v.name)
}

func (v String) Inspect() string { return fmt.Sprintf("%q", string(v)) }

type Inspect interface{ Inspect() string }
type Stringer interface{ String() string }

func inspect(v Value) string {
	switch v := v.(type) {
	case Inspect:
		return v.Inspect()
	case Stringer:
		return v.String()
	}
	panic(fmt.Sprintf("cannot inspect %#v", v))
}

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
