package eval

// This package contains the runtime representations of toe values.
// All values implement the Value interface, which has a prototype
// field (for values controlling the runtime behaviour, such as Break
// and Return we simply ignore their prototypes).

type ValueType uint8

//go:generate stringer -type=ValueType

const (
	_ = ValueType(iota)
	NIL
	STRING
	BOOLEAN
	NUMBER
	OBJECT
	BUILTIN
	FUNCTION
	// Meta values
	ERROR
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

func (v *Nil) Type() ValueType     { return NIL }
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

// ----------------------
// Runtime Control Values
// ----------------------

type Error struct{ reason Value }

func (e *Error) Type() ValueType { return ERROR }
