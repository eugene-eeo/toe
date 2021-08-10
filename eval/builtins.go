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

type binOpImpl func(left, right Value) Value

var binOpTable = map[binOpInfo]binOpImpl{}

func initBinOpTable() {
	ops := []struct {
		op          lexer.TokenType
		left, right ValueType
		impl        binOpImpl
	}{
		// Numbers
		{lexer.EQUAL_EQUAL, VT_NUMBER, VT_NUMBER, func(a, b Value) Value { return Boolean(a.(Number) == b.(Number)) }},
		{lexer.BANG_EQUAL, VT_NUMBER, VT_NUMBER, func(a, b Value) Value { return Boolean(a.(Number) != b.(Number)) }},
		{lexer.LESS, VT_NUMBER, VT_NUMBER, func(a, b Value) Value { return Boolean(a.(Number) < b.(Number)) }},
		{lexer.LESS_EQUAL, VT_NUMBER, VT_NUMBER, func(a, b Value) Value { return Boolean(a.(Number) <= b.(Number)) }},
		{lexer.GREATER, VT_NUMBER, VT_NUMBER, func(a, b Value) Value { return Boolean(a.(Number) > b.(Number)) }},
		{lexer.GREATER_EQUAL, VT_NUMBER, VT_NUMBER, func(a, b Value) Value { return Boolean(a.(Number) >= b.(Number)) }},
		{lexer.PLUS, VT_NUMBER, VT_NUMBER, func(a, b Value) Value { return a.(Number) + b.(Number) }},
		{lexer.MINUS, VT_NUMBER, VT_NUMBER, func(a, b Value) Value { return a.(Number) - b.(Number) }},
		{lexer.SLASH, VT_NUMBER, VT_NUMBER, func(a, b Value) Value { return a.(Number) / b.(Number) }},
		{lexer.STAR, VT_NUMBER, VT_NUMBER, func(a, b Value) Value { return a.(Number) * b.(Number) }},
		// String Operations
		{lexer.EQUAL_EQUAL, VT_STRING, VT_STRING, func(a, b Value) Value { return Boolean(a.(String) == b.(String)) }},
		{lexer.BANG_EQUAL, VT_STRING, VT_STRING, func(a, b Value) Value { return Boolean(a.(String) != b.(String)) }},
		{lexer.LESS, VT_STRING, VT_STRING, func(a, b Value) Value { return Boolean(a.(String) < b.(String)) }},
		{lexer.LESS_EQUAL, VT_STRING, VT_STRING, func(a, b Value) Value { return Boolean(a.(String) <= b.(String)) }},
		{lexer.GREATER, VT_STRING, VT_STRING, func(a, b Value) Value { return Boolean(a.(String) > b.(String)) }},
		{lexer.GREATER_EQUAL, VT_STRING, VT_STRING, func(a, b Value) Value { return Boolean(a.(String) >= b.(String)) }},
		{lexer.PLUS, VT_STRING, VT_STRING, func(a, b Value) Value { return a.(String) + b.(String) }},
		{lexer.STAR, VT_STRING, VT_NUMBER, func(a, b Value) Value {
			return String(strings.Repeat(string(a.(String)), int(b.(Number))))
		}},
	}
	for _, entry := range ops {
		binOpTable[binOpInfo{entry.op, entry.left, entry.right}] = entry.impl
	}
}
