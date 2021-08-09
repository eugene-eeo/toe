package eval

import "toe/lexer"

// =========
// Operators
// =========

func numberBinaryOp(op lexer.TokenType, left, right Number) Value {
	switch op {
	case lexer.EQUAL_EQUAL:
		return Boolean(left == right)
	case lexer.BANG_EQUAL:
		return Boolean(left != right)
	case lexer.LESS:
		return Boolean(left < right)
	case lexer.LESS_EQUAL:
		return Boolean(left <= right)
	case lexer.GREATER:
		return Boolean(left > right)
	case lexer.GREATER_EQUAL:
		return Boolean(left >= right)
	case lexer.PLUS:
		return left + right
	case lexer.MINUS:
		return left - right
	case lexer.SLASH:
		return left / right
	case lexer.STAR:
		return left * right
	}
	return nil
}

func stringBinaryOp(op lexer.TokenType, left, right String) Value {
	switch op {
	case lexer.EQUAL_EQUAL:
		return Boolean(left == right)
	case lexer.BANG_EQUAL:
		return Boolean(left != right)
	case lexer.LESS:
		return Boolean(left < right)
	case lexer.LESS_EQUAL:
		return Boolean(left <= right)
	case lexer.GREATER:
		return Boolean(left > right)
	case lexer.GREATER_EQUAL:
		return Boolean(left >= right)
	case lexer.PLUS:
		return left + right
	}
	return nil
}
