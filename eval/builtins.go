package eval

import (
	"strings"
	"toe/lexer"
)

// =========
// Operators
// =========

// binOpInfo represents a binary op entry -- when we execute binary operations
// we search the table using the (internal) types of both objects.
type binOpInfo struct {
	op          lexer.TokenType
	left, right ValueType
}

type binOpImpl func(ctx *Context, left, right Value) Value

var binOpTable = map[binOpInfo]binOpImpl{}

func initBinOpTable() {
	ops := []struct {
		op          lexer.TokenType
		left, right ValueType
		impl        binOpImpl
	}{
		// Numbers
		{lexer.LESS,          VT_NUMBER, VT_NUMBER, func(ctx *Context, a, b Value) Value { return Boolean(a.(Number) < b.(Number)) }},
		{lexer.LESS_EQUAL,    VT_NUMBER, VT_NUMBER, func(ctx *Context, a, b Value) Value { return Boolean(a.(Number) <= b.(Number)) }},
		{lexer.GREATER,       VT_NUMBER, VT_NUMBER, func(ctx *Context, a, b Value) Value { return Boolean(a.(Number) > b.(Number)) }},
		{lexer.GREATER_EQUAL, VT_NUMBER, VT_NUMBER, func(ctx *Context, a, b Value) Value { return Boolean(a.(Number) >= b.(Number)) }},
		{lexer.PLUS,          VT_NUMBER, VT_NUMBER, func(ctx *Context, a, b Value) Value { return a.(Number) + b.(Number) }},
		{lexer.MINUS,         VT_NUMBER, VT_NUMBER, func(ctx *Context, a, b Value) Value { return a.(Number) - b.(Number) }},
		{lexer.SLASH,         VT_NUMBER, VT_NUMBER, func(ctx *Context, a, b Value) Value { return a.(Number) / b.(Number) }},
		{lexer.STAR,          VT_NUMBER, VT_NUMBER, func(ctx *Context, a, b Value) Value { return a.(Number) * b.(Number) }},
		// String Operations
		{lexer.LESS,          VT_STRING, VT_STRING, func(ctx *Context, a, b Value) Value { return Boolean(a.(String) < b.(String)) }},
		{lexer.LESS_EQUAL,    VT_STRING, VT_STRING, func(ctx *Context, a, b Value) Value { return Boolean(a.(String) <= b.(String)) }},
		{lexer.GREATER,       VT_STRING, VT_STRING, func(ctx *Context, a, b Value) Value { return Boolean(a.(String) > b.(String)) }},
		{lexer.GREATER_EQUAL, VT_STRING, VT_STRING, func(ctx *Context, a, b Value) Value { return Boolean(a.(String) >= b.(String)) }},
		{lexer.PLUS,          VT_STRING, VT_STRING, func(ctx *Context, a, b Value) Value { return a.(String) + b.(String) }},
		{lexer.STAR,          VT_STRING, VT_NUMBER, func(ctx *Context, a, b Value) Value { return String(strings.Repeat(string(a.(String)), int(b.(Number)))) }},
		{lexer.LEFT_BRACKET,  VT_STRING, VT_NUMBER, func(ctx *Context, a, b Value) Value {
			str := a.(String)
			idx := int(b.(Number))
			if 0 < idx && idx < len(str) {
				return String(str[idx])
			}
			return newError(String("string index out of bounds"))
		}},
		// Array Operations
		{lexer.EQUAL_EQUAL,  VT_ARRAY, VT_ARRAY,  func(ctx *Context, a, b Value) Value { return bi_array_equal(ctx,      a.(*Array), b.(*Array)) }},
		{lexer.PLUS,         VT_ARRAY, VT_ARRAY,  func(ctx *Context, a, b Value) Value { return bi_array_concat_new(ctx, a.(*Array), b.(*Array)) }},
		{lexer.LEFT_BRACKET, VT_ARRAY, VT_NUMBER, func(ctx *Context, a, b Value) Value {
			arr := a.(*Array)
			idx := int(b.(Number))
			if 0 < idx && idx < len(arr.values) {
				return arr.values[idx]
			}
			return newError(String("array index out of bounds"))
		}},
		// Hash Operations
		{lexer.EQUAL_EQUAL,  VT_HASH, VT_HASH, func(ctx *Context, a, b Value) Value { return bi_hash_equal(ctx, a.(*Hash), b.(*Hash)) }},
		{lexer.LEFT_BRACKET, VT_HASH, VT_ANY,  func(ctx *Context, a, b Value) Value {
			hash := a.(*Hash)
			rv, found, err := hash.table.get(b)
			if err != nil {
				return err
			}
			if !found {
				return newError(String("key not in hash"))
			}
			return rv
		}},
	}
	for _, entry := range ops {
		binOpTable[binOpInfo{entry.op, entry.left, entry.right}] = entry.impl
	}
}

func bi_array_equal(ctx *Context, left, right *Array) Value {
	if len(left.values) != len(right.values) {
		return FALSE
	}
	for i, lhs := range left.values {
		rhs := right.values[i]
		rv := ctx.areObjectsEqual(lhs, rhs)
		if isError(rv) {
			return rv
		}
		if !isTruthy(rv) {
			return FALSE
		}
	}
	return TRUE
}

func bi_array_concat_new(ctx *Context, left, right *Array) Value {
	l_sz := len(left.values)
	r_sz := len(right.values)
	values := make([]Value, l_sz+r_sz)
	copy(values, left.values)
	copy(values[l_sz:], right.values)
	return newArray(values)
}

func bi_hash_equal(ctx *Context, left, right *Hash) Value {
	if left.table.size() != right.table.size() {
		return FALSE
	}
	for _, entry := range left.table.entries {
		if entry.hasValue() {
			key := *entry.key
			lhs := *entry.value
			rhs, found, err := right.table.get(key)
			if err != nil {
				return err
			}
			if !found {
				return FALSE
			}
			rv := ctx.areObjectsEqual(lhs, rhs)
			if isError(rv) {
				return rv
			}
			if !isTruthy(rv) {
				return FALSE
			}
		}
	}
	return TRUE
}
