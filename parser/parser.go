package parser

import "toe/lexer"

type (
	unaryParser  func() Expr
	binaryParser func(Expr) Expr
)

type Parser struct {
	filename      string
	tokens        []lexer.Token
	Errors        []ParserError
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
	PREC_CALL    // (), .
)

// ====
// init
// ====

func New(fn string, tokens []lexer.Token) *Parser {
	p := &Parser{
		filename: fn,
		tokens:   tokens,
		Errors:   []ParserError{},
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
//   statement   → for | while | if | block | break | continue | exprStmt
//   let      → "let" IDENT "=" expression ";"
//   for      → "for" "(" IDENT ":" expr ")" statement
//   while    → "while" "(" expr ")" statement
//   if       → "if" "(" expr ")" statement ( "else" statement )?
//   block    → "{" declaration* "}"
//   break    → "break" ";"
//   continue → "continue" ";"
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
	}
	return p.exprStmt()
}

func (p *Parser) letStmt() Stmt {
	token := p.consume()
	ident := p.expect(lexer.IDENTIFIER, "expected an identifier")
	p.expect(lexer.EQUAL, "expected =")
	expr := p.expression()
	p.expect(lexer.SEMICOLON, "expected ; after variable declaration")
	return newLet(token, ident, expr)
}

func (p *Parser) forStmt() Stmt {
	token := p.consume() // the 'for' token
	p.expect(lexer.LEFT_PAREN, "expected (")
	ident := p.expect(lexer.IDENTIFIER, "expected an identifier")
	p.expect(lexer.COLON, "expected :")
	iter := p.expression()
	p.expect(lexer.RIGHT_PAREN, "unclosed (")
	stmt := p.statement()
	return newFor(token, newIdentifier(ident), iter, stmt)
}

func (p *Parser) whileStmt() Stmt {
	token := p.consume()
	p.expect(lexer.LEFT_PAREN, "expected (")
	cond := p.expression()
	p.expect(lexer.RIGHT_PAREN, "unclosed (")
	stmt := p.statement()
	return newWhile(token, cond, stmt)
}

func (p *Parser) ifStmt() Stmt {
	token := p.consume()
	p.expect(lexer.LEFT_PAREN, "expected (")
	cond := p.expression()
	p.expect(lexer.RIGHT_PAREN, "unclosed (")
	then := p.statement()
	var elseStmt Stmt = nil
	if p.match(lexer.ELSE) {
		elseStmt = p.statement()
	}
	return newIf(token, cond, then, elseStmt)
}

func (p *Parser) blockStmt() Stmt {
	token := p.consume()
	stmts := []Stmt{}
	for !p.isAtEnd() && !p.check(lexer.RIGHT_BRACE) {
		stmts = append(stmts, p.declaration())
	}
	p.expect(lexer.RIGHT_BRACE, "unmatched {")
	return newBlock(token, stmts)
}

func (p *Parser) continueStmt() Stmt {
	token := p.consume()
	p.expect(lexer.SEMICOLON, "expected ; after continue")
	return newContinue(token)
}

func (p *Parser) breakStmt() Stmt {
	token := p.consume()
	p.expect(lexer.SEMICOLON, "expected ; after break")
	return newBreak(token)
}

func (p *Parser) exprStmt() Stmt {
	expr := p.expression()
	p.expect(lexer.SEMICOLON, "expected ; after expression statement")
	if expr == nil {
		return nil
	}
	return newExprStmt(expr.Tok(), expr)
}

// ==================
// expression parsing
// ==================

// expression matches a single expression.
func (p *Parser) expression() Expr { return p.precedence(PREC_LOWEST) }
func (p *Parser) precedence(prec int) Expr {
	unary, ok := p.unaryParsers[p.peek().Type]
	if !ok {
		panic(p.error(p.peek(), "not an expression: %s", p.previous().Type))
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
	p.expect(lexer.RIGHT_PAREN, "unmatched (")
	return expr
}

func (p *Parser) assign(left Expr) Expr {
	tok := p.consume()
	right := p.precedence(PREC_ASSIGN-1)
	switch left.Type() {
	case IDENTIFIER:
		return newAssign(tok, left.Tok(), right)
	default:
		// this is not an error worth panicking over.
		// just move along -- we will put it in `.errors'.
		p.error(left.Tok(), "invalid assignment target")
		return nil
	}
}

func (p *Parser) get(left Expr) Expr {
	tok := p.consume()
	right := p.precedence(PREC_CALL)
	// we allow any names
	switch right.Type() {
	case IDENTIFIER:
		return newGet(tok, left, right.Tok())
	case LITERAL:
		switch right.(*Literal).Tok().Type {
		case lexer.NIL:
			fallthrough
		case lexer.TRUE:
			fallthrough
		case lexer.FALSE:
			return newGet(tok, left, right.Tok())
		}
	}
	panic(p.error(right.Tok(), "expected an identifier after ."))
}

func (p *Parser) binary(left Expr) Expr {
	tok := p.consume()
	return newBinary(tok, left, p.precedence(p.precedences[tok.Type]))
}

func (p *Parser) and(left Expr) Expr {
	tok := p.consume()
	return newAnd(tok, left, p.precedence(PREC_AND))
}

func (p *Parser) or(left Expr) Expr {
	tok := p.consume()
	return newOr(tok, left, p.precedence(PREC_AND))
}

func (p *Parser) identifier() Expr {
	tok := p.consume()
	return newIdentifier(tok)
}

func (p *Parser) literal() Expr {
	tok := p.consume()
	return newLiteral(tok)
}
