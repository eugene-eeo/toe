// Code generated by "stringer -type=TokenType"; DO NOT EDIT.

package lexer

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[LEFT_PAREN-1]
	_ = x[RIGHT_PAREN-2]
	_ = x[LEFT_BRACE-3]
	_ = x[RIGHT_BRACE-4]
	_ = x[COMMA-5]
	_ = x[COLON-6]
	_ = x[DOT-7]
	_ = x[MINUS-8]
	_ = x[PLUS-9]
	_ = x[SEMICOLON-10]
	_ = x[SLASH-11]
	_ = x[STAR-12]
	_ = x[BANG-13]
	_ = x[BANG_EQUAL-14]
	_ = x[EQUAL-15]
	_ = x[EQUAL_EQUAL-16]
	_ = x[GREATER-17]
	_ = x[GREATER_EQUAL-18]
	_ = x[LESS-19]
	_ = x[LESS_EQUAL-20]
	_ = x[IDENTIFIER-21]
	_ = x[STRING-22]
	_ = x[NUMBER-23]
	_ = x[LET-24]
	_ = x[AND-25]
	_ = x[OR-26]
	_ = x[ELSE-27]
	_ = x[FALSE-28]
	_ = x[FN-29]
	_ = x[FOR-30]
	_ = x[IF-31]
	_ = x[NIL-32]
	_ = x[RETURN-33]
	_ = x[SUPER-34]
	_ = x[TRUE-35]
	_ = x[WHILE-36]
	_ = x[BREAK-37]
	_ = x[CONTINUE-38]
	_ = x[EOF-39]
}

const _TokenType_name = "LEFT_PARENRIGHT_PARENLEFT_BRACERIGHT_BRACECOMMACOLONDOTMINUSPLUSSEMICOLONSLASHSTARBANGBANG_EQUALEQUALEQUAL_EQUALGREATERGREATER_EQUALLESSLESS_EQUALIDENTIFIERSTRINGNUMBERLETANDORELSEFALSEFNFORIFNILRETURNSUPERTRUEWHILEBREAKCONTINUEEOF"

var _TokenType_index = [...]uint8{0, 10, 21, 31, 42, 47, 52, 55, 60, 64, 73, 78, 82, 86, 96, 101, 112, 119, 132, 136, 146, 156, 162, 168, 171, 174, 176, 180, 185, 187, 190, 192, 195, 201, 206, 210, 215, 220, 228, 231}

func (i TokenType) String() string {
	i -= 1
	if i >= TokenType(len(_TokenType_index)-1) {
		return "TokenType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
