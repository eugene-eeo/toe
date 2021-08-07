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
	case NIL_TYPE:
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
			obj := proto.(*Object)
			if v, ok := obj.props[attr]; ok {
				return v, true
			}
		}
		proto = ctx.getPrototype(proto)
	}
	return nil, false
}

func (ctx *Context) bind(fn Value, this Value) Value {
	switch fn.Type() {
	case BUILTIN:
		return fn.(*Builtin).Bind(this)
	case FUNCTION:
		return fn.(*Function).Bind(this)
	}
	return fn
}

// callFunction calls the given function fn, with the given
// this binding and arguments. If fn is not a _callable_, then
// isCallable will be false.
func (ctx *Context) callFunction(fn Value, args []Value) (rv Value, isCallable bool) {
	switch fn := fn.(type) {
	case *Builtin:
		return fn.Call(ctx, args), true
	case *Function:
		return fn.Call(ctx, args), true
	}
	return nil, false
}

// setAttr sets an attribute on an object.
func (ctx *Context) setAttr(obj Value, attr string, value Value) (Value, bool) {
	switch obj := obj.(type) {
	case *Object:
		obj.props[attr] = value
		return value, true
	}
	return nil, false
}

// ================
// Iterator Support
// ================
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

func (si *StringIterator) End() Value { return NIL }
func (si *StringIterator) Done() Value { return newBool(si.i >= len(si.s.value)) }
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

func (ai *ArrayIterator) End() Value { return NIL }
func (ai *ArrayIterator) Done() Value { return newBool(ai.i >= len(ai.arr.arr)) }
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
	}
	return nil, false
}

// ================
// Operator Support
// ================

func (ctx *Context) evalUnaryValues(op lexer.TokenType, right Value) Value {
	switch op {
	case lexer.BANG:
		return newBool(!isTruthy(right))
	case lexer.MINUS:
		if right.Type() == NUMBER {
			return &Number{-right.(*Number).value}
		}
	}
	return &Error{&String{"unsupported operation"}}
}

func (ctx *Context) evalBinaryValues(
	op lexer.TokenType,
	left, right Value,
) Value {
	// Fast case, == and != can fall-back to pointer equality.
	if op == lexer.EQUAL_EQUAL && left == right {
		return TRUE
	}
	if op == lexer.BANG_EQUAL && left != right {
		return TRUE
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
		return newBool(lhs == rhs)
	case lexer.BANG_EQUAL:
		return newBool(lhs != rhs)
	case lexer.GREATER:
		return newBool(lhs > rhs)
	case lexer.GREATER_EQUAL:
		return newBool(lhs >= rhs)
	case lexer.LESS:
		return newBool(lhs < rhs)
	case lexer.LESS_EQUAL:
		return newBool(lhs <= rhs)
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
