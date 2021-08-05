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
	META
)

type Value interface {
	Prototype() Value
	Type() ValueType
}

// basic is an embeddable struct to give each `real' object
// the common properties.
type basic struct {
	proto Value
	typ   ValueType
}

func (b basic) Prototype() Value { return b.proto }
func (b basic) Type() ValueType  { return b.typ }

func makeBasic(proto Value, typ ValueType) basic {
	return basic{proto, typ}
}

// -----------------
// Boxed value types
// -----------------

type Nil struct{ basic }

func newNil(object Value) *Nil {
	return &Nil{makeBasic(object, NIL)}
}

type Boolean struct {
	basic
	value bool
}

func newBoolean(object Value, value bool) *Boolean {
	return &Boolean{makeBasic(object, BOOLEAN), value}
}

type Number struct {
	basic
	value float64
}

func newNumber(object Value, value float64) *Number {
	return &Number{makeBasic(object, NUMBER), value}
}

type String struct {
	basic
	value string
}

func newString(object Value, value string) *String {
	return &String{makeBasic(object, STRING), value}
}

type Object struct {
	basic
	props map[string]Value
}

func newObject(object Value) *Object {
	return &Object{
		basic: makeBasic(object, OBJECT),
		props: map[string]Value{},
	}
}

// ----------------------
// Runtime Control Values
// ----------------------

type Error struct {
	reason Value
}

func (e *Error) Prototype() Value { return nil }
func (e *Error) Type() ValueType { return META }
