package parser

import "toe/lexer"

type (
	unaryParser  func() Expr
	binaryParser func(Expr) Expr
)

type Parser struct {
	filename      string
	tokens        []lexer.Token
	Errors        []error
	curr          int // how many we have consumed.
	unaryParsers  map[lexer.TokenType]unaryParser
	binaryParsers map[lexer.TokenType]binaryParser
	precedences   map[lexer.TokenType]int
}

const (
	PREC_LOWEST  = iota
	PREC_ASSIGN  // =
	PREC_AND     // and, or
	PREC_EQ      // ==, !=
	PREC_CMP     // <=, <, >, >=
	PREC_SUM     // +, -
	PREC_PRODUCT // *, /
	PREC_UNARY   // !, -
	PREC_CALL    // (), ., ->
)

// ====
// init
// ====

func New(fn string, tokens []lexer.Token) *Parser {
	p := &Parser{
		filename: fn,
		tokens:   tokens,
		Errors:   []error{},
		curr:     0,
	}
	p.unaryParsers = map[lexer.TokenType]unaryParser{
		lexer.LEFT_PAREN: p.grouping,
		lexer.IDENTIFIER: p.identifier,
		lexer.NUMBER:     p.literal,
		lexer.STRING:     p.literal,
		lexer.TRUE:       p.literal,
		lexer.FALSE:      p.literal,
		lexer.NIL:        p.literal,
		lexer.BANG:       p.unary,
		lexer.MINUS:      p.unary,
		lexer.FN:         p.function,
		lexer.SUPER:      p.super,
	}
	// note: need to make sure that every entry in binaryParsers
	// has a corresponding entry in precedences.
	p.binaryParsers = map[lexer.TokenType]binaryParser{
		lexer.EQUAL:         p.assign,
		lexer.AND:           p.and,
		lexer.OR:            p.or,
		lexer.EQUAL_EQUAL:   p.binary,
		lexer.BANG_EQUAL:    p.binary,
		lexer.GREATER:       p.binary,
		lexer.GREATER_EQUAL: p.binary,
		lexer.LESS:          p.binary,
		lexer.LESS_EQUAL:    p.binary,
		lexer.PLUS:          p.binary,
		lexer.MINUS:         p.binary,
		lexer.STAR:          p.binary,
		lexer.SLASH:         p.binary,
		lexer.DOT:           p.get,
		lexer.MINUS_GREATER: p.get,
		lexer.LEFT_PAREN:    p.call,
	}
	p.precedences = map[lexer.TokenType]int{
		lexer.EQUAL:         PREC_ASSIGN,
		lexer.AND:           PREC_AND,
		lexer.OR:            PREC_AND,
		lexer.EQUAL_EQUAL:   PREC_EQ,
		lexer.BANG_EQUAL:    PREC_EQ,
		lexer.GREATER:       PREC_CMP,
		lexer.GREATER_EQUAL: PREC_CMP,
		lexer.LESS:          PREC_CMP,
		lexer.LESS_EQUAL:    PREC_CMP,
		lexer.PLUS:          PREC_SUM,
		lexer.MINUS:         PREC_SUM,
		lexer.STAR:          PREC_PRODUCT,
		lexer.SLASH:         PREC_PRODUCT,
		lexer.DOT:           PREC_CALL,
		lexer.MINUS_GREATER: PREC_CALL,
		lexer.LEFT_PAREN:    PREC_CALL,
	}
	return p
}

// =====
// utils
// =====

// consume consumes one token
func (p *Parser) consume() lexer.Token {
	if !p.isAtEnd() {
		p.curr++
	}
	return p.previous()
}

// previous returns the most recently consumed token
func (p *Parser) previous() lexer.Token { return p.tokens[p.curr-1] }

// peek returns the token to be consumed
func (p *Parser) peek() lexer.Token { return p.tokens[p.curr] }

// isAtEnd returns true if the current token is an EOF token
func (p *Parser) isAtEnd() bool { return p.peek().Type == lexer.EOF }

// check returns if the peek token matches the given type
func (p *Parser) check(t lexer.TokenType) bool {
	return !p.isAtEnd() && p.peek().Type == t
}

// match consumes the token if it matches any of the given types
func (p *Parser) match(types ...lexer.TokenType) bool {
	for _, t := range types {
		if p.check(t) {
			p.consume()
			return true
		}
	}
	return false
}

// ===========
// entry point
// ===========

// module → declaration*

func (p *Parser) Parse() *Module {
	module := &Module{Filename: p.filename, Stmts: []Stmt{}}
	for !p.isAtEnd() {
		module.Stmts = append(module.Stmts, p.declaration())
	}
	return module
}

// =================
// statement parsing
// =================
//
// to differentiate between a declaration and a statement, we
// use the following rules; this is to disallow e.g.:
//   if (expr) let x = 1; <-- what's the point?
//
// the main entry point is the declaration rule:
//
//   declaration → let | statement
//   statement   → for | while | if | block | break | continue | return | exprStmt
//   let      → "let" IDENT "=" expression ";"
//   for      → "for" "(" IDENT ":" expr ")" statement
//   while    → "while" "(" expr ")" statement
//   if       → "if" "(" expr ")" statement ( "else" statement )?
//   block    → "{" declaration* "}"
//   break    → "break" ";"
//   continue → "continue" ";"
//   return   → "return" ( expr )? ";"
//   exprStmt → expression ";"
//
// note: since most of the let,for,... are keywords in Go,
// they are named __Stmt().

func (p *Parser) declaration() (stmt Stmt) {
	defer func() {
		// This will be called repeatedly as we parse statements, so
		// this is a good place to synchronize(). We have to make
		// sure that all top-level calls to parse statements/expressions
		// have a recover.
		if rv := recover(); rv != nil {
			if _, ok := rv.(ParserError); ok {
				p.synchronize()
				stmt = nil
				return
			}
			panic(rv)
		}
	}()
	if p.check(lexer.LET) {
		stmt = p.letStmt()
	} else {
		stmt = p.statement()
	}
	return
}

func (p *Parser) statement() Stmt {
	switch {
	case p.check(lexer.FOR):
		return p.forStmt()
	case p.check(lexer.WHILE):
		return p.whileStmt()
	case p.check(lexer.IF):
		return p.ifStmt()
	case p.check(lexer.LEFT_BRACE):
		return p.blockStmt()
	case p.check(lexer.CONTINUE):
		return p.continueStmt()
	case p.check(lexer.BREAK):
		return p.breakStmt()
	case p.check(lexer.RETURN):
		return p.returnStmt()
	}
	return p.exprStmt()
}

func (p *Parser) letStmt() Stmt {
	p.consume()
	ident := p.expect(lexer.IDENTIFIER, "expect an identifier")
	p.expect(lexer.EQUAL, "expect '=' after identifier")
	expr := p.expression()
	p.expect(lexer.SEMICOLON, "expect ';' after variable declaration")
	return newLet(ident, expr)
}

func (p *Parser) forStmt() Stmt {
	p.consume()
	p.expect(lexer.LEFT_PAREN, "expect '(' after 'for'")
	ident := p.expect(lexer.IDENTIFIER, "expect an identifier after '('")
	p.expect(lexer.COLON, "expect ':' after identifier")
	iter := p.expression()
	p.expect(lexer.RIGHT_PAREN, "unclosed '('")
	stmt := p.statement()
	return newFor(ident, iter, stmt)
}

func (p *Parser) whileStmt() Stmt {
	p.consume()
	p.expect(lexer.LEFT_PAREN, "expect '(' after 'while'")
	cond := p.expression()
	p.expect(lexer.RIGHT_PAREN, "unclosed '('")
	stmt := p.statement()
	return newWhile(cond, stmt)
}

func (p *Parser) ifStmt() Stmt {
	p.consume()
	p.expect(lexer.LEFT_PAREN, "expect '(' after 'if'")
	cond := p.expression()
	p.expect(lexer.RIGHT_PAREN, "unclosed '('")
	then := p.statement()
	var elseStmt Stmt = nil
	if p.match(lexer.ELSE) {
		elseStmt = p.statement()
	}
	return newIf(cond, then, elseStmt)
}

func (p *Parser) blockStmt() Stmt {
	p.consume()
	stmts := []Stmt{}
	for !p.isAtEnd() && !p.check(lexer.RIGHT_BRACE) {
		stmts = append(stmts, p.declaration())
	}
	p.expect(lexer.RIGHT_BRACE, "unclosed '{'")
	return newBlock(stmts)
}

func (p *Parser) continueStmt() Stmt {
	token := p.consume()
	p.expect(lexer.SEMICOLON, "expect ';' after 'continue'")
	return newContinue(token)
}

func (p *Parser) breakStmt() Stmt {
	token := p.consume()
	p.expect(lexer.SEMICOLON, "expect ';' after 'break'")
	return newBreak(token)
}

func (p *Parser) returnStmt() Stmt {
	token := p.consume()
	expr := Expr(nil)
	if !p.check(lexer.SEMICOLON) {
		expr = p.expression()
	}
	p.expect(lexer.SEMICOLON, "expect ';' after 'return'")
	return newReturn(token, expr)
}

func (p *Parser) exprStmt() Stmt {
	expr := p.expression()
	p.expect(lexer.SEMICOLON, "expect ';' after expression statement")
	if expr == nil {
		return nil
	}
	return newExprStmt(expr)
}

// ==================
// expression parsing
// ==================
//
// Ambiguity is resolved via a Pratt parser; below we give the rules,
// from those with the least precedence to that with the most.
//
// expression → assign
//            | and | or
//            | binary
//            | unary
//            | call | get
//            | literal
//            | super
// assign   → ( get | IDENTIFIER ) "=" expression
// and      → expression "and" expression
// or       → expression "or" expression
// binary   → expression ( "==" | "!=" | "<=" | ">=" | "<" | ">" | "+" | "-" | "*" | "/" ) expression
// unary    → ( "!" | "-" ) expression
// get      → expression ( "." | "->" ) ( IDENTIFIER | "nil" | "true" | "false" )
// call     → expression "(" args ")"
// args     → expression ( "," args )? | ε
// literal  → STRING | IDENTIFIER | NUMBER | TRUE | FALSE | NIL | func
// function → "fn" "(" params ")" block
// params   → IDENTIFIER ( "," params )? | ε
// super    → "super" "." IDENTIFIER

// expression matches a single expression.
func (p *Parser) expression() Expr { return p.precedence(PREC_LOWEST) }
func (p *Parser) precedence(prec int) Expr {
	unary, ok := p.unaryParsers[p.peek().Type]
	if !ok {
		panic(p.error(p.peek(), "not an expression: %s", p.peek().Type))
	}
	expr := unary()
	for !p.check(lexer.SEMICOLON) && prec < p.peekPrecedence() {
		expr = p.binaryParsers[p.peek().Type](expr)
	}
	return expr
}

func (p *Parser) peekPrecedence() int {
	if prec, ok := p.precedences[p.peek().Type]; ok {
		return prec
	}
	return PREC_LOWEST
}

func (p *Parser) unary() Expr {
	tok := p.consume()
	return newUnary(tok, p.precedence(PREC_UNARY-1))
}

func (p *Parser) grouping() Expr {
	p.consume()
	expr := p.expression()
	p.expect(lexer.RIGHT_PAREN, "unclosed '('")
	return expr
}

func (p *Parser) assign(left Expr) Expr {
	tok := p.consume()
	right := p.precedence(PREC_ASSIGN-1)
	switch left := left.(type) {
	case *Get:
		return newSet(left.Object, left.Name, left.Bound, right)
	case *Identifier:
		return newAssign(left.Id, right)
	default:
		p.error(tok, "invalid assignment target")
		return nil
	}
}

func (p *Parser) get(left Expr) Expr {
	// Be careful that are two kinds of attribute accesses.
	// The typical syntax defaults to `binding' mode, since most of the
	// time I expect that users would want to set callbacks; moreover
	// this makes the syntax much more uniform (IMO) and explicit.
	//
	//     obj.x = other.fn   (attr 'fn' of other is bound to other)
	//     obj.x = other->fn  (attr 'fn' of other is _not_ bound to other)
	//
	// a weird side-effect is that now the semantics for assignment has
	// to be changed as well. This should be fine, since most of the time
	// we don't use assignment results as an expression:
	//
	//     obj.a  = x  returns x bound to obj
	//     obj->a = x  returns x
	//
	tok := p.consume()
	bound := tok.Type == lexer.DOT // is this a binding access or not?
	name := p.consume()
	switch name.Type {
	case lexer.IDENTIFIER, lexer.NIL, lexer.TRUE, lexer.FALSE:
		return newGet(left, name, bound)
	}
	panic(p.error(name, "expected a name after %q", tok.Lexeme))
}

func (p *Parser) binary(left Expr) Expr {
	opToken := p.consume()
	return newBinary(left, opToken, p.precedence(p.precedences[opToken.Type]))
}

func (p *Parser) and(left Expr) Expr {
	opToken := p.consume()
	return newAnd(left, opToken, p.precedence(PREC_AND))
}

func (p *Parser) or(left Expr) Expr {
	opToken := p.consume()
	return newOr(left, opToken, p.precedence(PREC_AND))
}

func (p *Parser) identifier() Expr {
	return newIdentifier(p.consume())
}

func (p *Parser) literal() Expr {
	return newLiteral(p.consume())
}

func (p *Parser) function() Expr {
	fnTok := p.consume()
	p.expect(lexer.LEFT_PAREN, "expected a '(' after 'fn'")
	params := []lexer.Token{}
	// params
	for !p.isAtEnd() && !p.check(lexer.RIGHT_PAREN) {
		tok := p.expect(lexer.IDENTIFIER, "expect an identifier or ')' after '('")
		params = append(params, tok)
		if !p.match(lexer.COMMA) {
			break
		}
	}
	p.expect(lexer.RIGHT_PAREN, "unclosed '('")
	block := p.blockStmt().(*Block)
	return newFunction(fnTok, params, block)
}

func (p *Parser) call(left Expr) Expr {
	lParenTok := p.consume()
	args := []Expr{}
	for !p.isAtEnd() && !p.check(lexer.RIGHT_PAREN) {
		args = append(args, p.expression())
		if !p.match(lexer.COMMA) {
			break
		}
	}
	p.expect(lexer.RIGHT_PAREN, "unclosed '('")
	return newCall(left, lParenTok, args)
}

func (p *Parser) super() Expr {
	superToken := p.consume()
	if !p.match(lexer.DOT, lexer.MINUS_GREATER) {
		panic(p.error(p.peek(), "expect '.' or '->' after 'super'"))
	}
	dot := p.previous()
	name := p.consume()
	switch name.Type {
	case lexer.IDENTIFIER, lexer.NIL, lexer.TRUE, lexer.FALSE:
		return newSuper(superToken, name, dot.Type == lexer.DOT)
	}
	panic(p.error(name, "expected a name after %q", dot.Lexeme))
}
