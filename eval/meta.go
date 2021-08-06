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
	case NIL:      return ctx._Nil
	case STRING:   return ctx._String
	case NUMBER:   return ctx._Number
	case BUILTIN:  fallthrough
	case FUNCTION: return ctx._Function
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
