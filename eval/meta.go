package eval

import (
	"fmt"
	"toe/lexer"
)

// This package implements the meta-functions, for example
// for calling a function, fetching an attribute, operators,
// etc.

// =======================
// Prototypes & Attributes
// =======================

func (ctx *Context) getPrototype(obj Value) Value {
	switch obj.Type() {
	case NIL:
		return ctx._Nil
	case STRING:
		return ctx._String
	case NUMBER:
		return ctx._Number
	case BUILTIN:
		fallthrough
	case FUNCTION:
		return ctx._Function
	case OBJECT:
		return obj.(*Object).proto
	}
	panic(fmt.Sprintf("cannot return prototype of %#v", obj))
}

func (ctx *Context) getAttr(src Value, attr string) (Value, bool) {
	if src.Type() == OBJECT {
		obj := src.(*Object)
		if v, ok := obj.props[attr]; ok {
			return v, true
		}
	}
	// we have to search the prototype chain.
	proto := ctx.getPrototype(src)
	for proto != nil {
		if proto.Type() == OBJECT {
			obj := src.(*Object)
			if v, ok := obj.props[attr]; ok {
				return v, true
			}
		}
		proto = ctx.getPrototype(proto)
	}
	return nil, false
}

// callFunction calls the given function fn, with the given
// this binding and arguments. If fn is not a _callable_, then
// isCallable will be false. For convenience, if this == nil,
// then this is set to the nil object.
func (ctx *Context) callFunction(fn Value, this Value, args []Value) (rv Value, isCallable bool) {
	if this == nil {
		this = ctx._nil
	}
	switch fn := fn.(type) {
	case *Builtin:
		fn = fn.Bind(this)
		return fn.Call(ctx, args), true
	}
	return nil, false
}

// ================
// Iterator Support
// ================
//
// Iterators:
//                     | let iterator = it.iter()
//    for (x : it) {   | while !iterator.done() {
//        ...          |    let x = iterator.next()
//    }                | }
//                     | iterator.end()

type Iterator interface {
	Next() Value
	Done() Value
	End() Value
}

type StringIterator struct {
	ctx *Context
	s   *String
	i   int
}

func (si *StringIterator) End() Value { return si.ctx._nil }
func (si *StringIterator) Done() Value { return si.ctx.newBool(si.i >= len(si.s.value)) }
func (si *StringIterator) Next() Value {
	v := si.s.value[si.i]
	si.i++
	return &String{string(v)}
}

type ArrayIterator struct {
	ctx *Context
	arr *Array
	i   int
}

func (ai *ArrayIterator) End() Value { return ai.ctx._nil }
func (ai *ArrayIterator) Done() Value { return ai.ctx.newBool(ai.i >= len(ai.arr.arr)) }
func (ai *ArrayIterator) Next() Value {
	v := ai.arr.arr[ai.i]
	ai.i++
	return v
}

func (ctx *Context) getIterator(obj Value) (Iterator, bool) {
	switch obj.Type() {
	case ARRAY:
		return &ArrayIterator{ctx: ctx, arr: obj.(*Array)}, true
	case STRING:
		return &StringIterator{ctx: ctx, s: obj.(*String)}, true
	// TODO: with user-given iterators, we need to freeze
	// the current environment.
	}
	return nil, false
}

// ================
// Operator Support
// ================

func (ctx *Context) evalBinary(op lexer.TokenType, left, right Value) Value {
	// Fast case, == and != can fall-back to pointer equality.
	if op == lexer.EQUAL_EQUAL && left == right {
		return ctx._true
	}
	if op == lexer.BANG_EQUAL && left != right {
		return ctx._true
	}
	// See if we know how to handle the operator.
	switch {
	case left.Type() == NUMBER && right.Type() == NUMBER:
		return ctx.evalNumberBinary(op, left.(*Number), right.(*Number))
	}
	return &Error{&String{"unsupported operation"}}
}

func (ctx *Context) evalNumberBinary(op lexer.TokenType, left, right *Number) Value {
	lhs := left.value
	rhs := right.value
	switch op {
	case lexer.EQUAL_EQUAL:
		return ctx.newBool(lhs == rhs)
	case lexer.BANG_EQUAL:
		return ctx.newBool(lhs != rhs)
	case lexer.GREATER:
		return ctx.newBool(lhs > rhs)
	case lexer.GREATER_EQUAL:
		return ctx.newBool(lhs >= rhs)
	case lexer.LESS:
		return ctx.newBool(lhs < rhs)
	case lexer.LESS_EQUAL:
		return ctx.newBool(lhs <= rhs)
	case lexer.PLUS:
		return &Number{lhs + rhs}
	case lexer.MINUS:
		return &Number{lhs - rhs}
	case lexer.SLASH:
		return &Number{lhs / rhs}
	case lexer.STAR:
		return &Number{lhs * rhs}
	}
	return &Error{&String{fmt.Sprintf("unsupported op between numbers: %s", op)}}
}
