package eval

import (
	"fmt"
	"toe/lexer"
	"unicode/utf8"
)

// =========
// Iterators
// =========
//
// For loops are equivalent to:
//
//                  | let it = obj.iter()
// for (x : obj)    | while (!it.done()) {
//     ...          |   let x = it.next()
//                  |   ...
//                  | }
//                  | it.close()

type Iterator interface {
	Close() Value
	Done() Value
	Next() Value
}

type StringIterator struct {
	err bool
	i   int
	s   String
}

func (si *StringIterator) Close() Value { return NIL }
func (si *StringIterator) Done() Value  { return Boolean(si.err || si.i >= len(si.s)) }
func (si *StringIterator) Next() Value {
	r, w := utf8.DecodeRuneInString(string(si.s)[si.i:])
	si.i += w
	return String(r)
}

func getIterator(v Value) (Iterator, bool) {
	switch v := v.(type) {
	case String:
		return &StringIterator{s: v}, true
	}
	return nil, false
}

// ============
// Object Model
// ============
//
// Toe is prototype-based; thus there are no classes, just objects (bag-of-values).
// Below we implement the functions needed to modify these objects. Note that some
// of them may return error values; you have to check for them!
//
//  1. Whenever we see a nil prototype, we can stop searching.
//  2. All built-in `types' have prototype == Object, except for nil.
//  3. Object's prototype == nil.

// -------
// Binding
// -------

func (ctx *Context) bind(obj Value, this Value) Value {
	switch obj := obj.(type) {
	case *Function:
		return obj.Bind(this)
	}
	return obj
}

func (f *Function) Bind(this Value) *Function {
	// A bound function cannot be bound again. Note that this is different
	// from (explicitly) binding to NIL.
	if f.this != nil {
		return f
	}
	// This means that we transfer the properties, but also give the new
	// function it's own `namespace'.
	bound := newFunction(f.module, f.node, this, f.closure)
	bound.Object.proto = f
	return bound
}

// -------------
// Get Prototype
// -------------

func (ctx *Context) getPrototype(obj Value) Value {
	switch obj := obj.(type) {
	case *Object:
		return obj.proto
	case *Function:
		return obj.proto
	}
	return nil
}

// --------
// Get Slot
// --------

func (ctx *Context) getSlot(obj Value, name string) Value {
	for obj != nil {
		if obj_ho, ok := obj.(hasObject); ok {
			if v, ok := obj_ho.object().slots[name]; ok {
				return v
			}
		}
		obj = ctx.getPrototype(obj)
	}
	err := newError(String(fmt.Sprintf("object has no slot %q", name)))
	return err
}

type hasObject interface{ object() *Object }

func (o *Object) object() *Object { return o }

// --------
// Set Slot
// --------

func (ctx *Context) setSlot(obj Value, name string, val Value) Value {
	if obj_ho, ok := obj.(hasObject); ok {
		oo := obj_ho.object()
		if oo.frozen {
			err := newError(String(fmt.Sprintf("attempt to set slot %q on frozen object", name)))
			return err
		}
		oo.slots[name] = val
		return val
	}
	err := newError(String(fmt.Sprintf("cannot set slot %q on object", name)))
	return err
}

// ==============
// Function Calls
// ==============

func (ctx *Context) call(callee Value, args []Value) Value {
	switch callee := callee.(type) {
	case *Function:
		return callee.Call(ctx, args)
	}
	err := newError(String("not a function"))
	return err
}

func (f *Function) getThis() Value {
	if f.this == nil {
		return NIL
	} else {
		return f.this
	}
}

func (f *Function) Call(ctx *Context, args []Value) Value {
	old_mod := ctx.module
	old_env := ctx.env
	old_this := ctx.this

	ctx.env = f.closure
	ctx.module = f.module
	ctx.this = f.getThis()
	ctx.pushEnv()
	ctx.pushFunc(f.ctx)

	ctx.env.set("this", ctx.this)
	for i, id := range f.node.Params {
		name := id.Lexeme
		if len(args) <= i {
			ctx.env.set(name, NIL)
		} else {
			ctx.env.set(name, args[i])
		}
	}

	// Remember to unwrap return values.
	rv := ctx.evalBlock(f.node.Body)
	if isReturn(rv) {
		rv = rv.(Return).value
	}

	ctx.popFunc()
	ctx.popEnv()
	ctx.this = old_this
	ctx.env = old_env
	ctx.module = old_mod
	return rv
}

// =========
// Operators
// =========

func (ctx *Context) binary(op lexer.TokenType, left, right Value) Value {
	if op == lexer.EQUAL_EQUAL && left == right {
		return TRUE
	}
	if op == lexer.BANG_EQUAL && left != right {
		return TRUE
	}
	switch {
	case op == lexer.EQUAL_EQUAL:
		return Boolean(left == right)
	case op == lexer.BANG_EQUAL:
		return Boolean(left != right)
	case left.Type() == VT_NUMBER && right.Type() == VT_NUMBER:
		rv := numberBinaryOp(op, left.(Number), right.(Number))
		if rv != nil {
			return rv
		}
	case left.Type() == VT_STRING && right.Type() == VT_STRING:
		rv := stringBinaryOp(op, left.(String), right.(String))
		if rv != nil {
			return rv
		}
	}
	err := newError(String(fmt.Sprintf(
		"unsupported operands for %q: %s and %s",
		op.String(),
		left.Type().String(),
		right.Type().String(),
	)))
	return err
}

func (ctx *Context) unary(op lexer.TokenType, right Value) Value {
	switch {
	case op == lexer.BANG:
		return Boolean(!isTruthy(right))
	case op == lexer.MINUS && right.Type() == VT_NUMBER:
		return Number(-right.(Number))
	}
	err := newError(String(fmt.Sprintf(
		"unsupported operand for %q: %s",
		op.String(),
		right.Type().String(),
	)))
	return err
}
