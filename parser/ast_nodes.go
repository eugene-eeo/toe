// generated by tool/make_ast.py, do not modify!
package parser

import "toe/lexer"

const (
	_ = NodeType(iota)
	MODULE
	LET
	BLOCK
	FOR
	EXPR_STMT
	BINARY
	AND
	OR
	ASSIGN
	UNARY
	IDENTIFIER
	LITERAL
)

type Module struct {
	Token      lexer.Token
	Filename   string
	Statements []Stmt
}

func (node *Module) Tok() lexer.Token { return node.Token }
func (node *Module) Type() NodeType   { return MODULE }
func (node *Module) stmt()            {}

type Let struct {
	Token lexer.Token
	Name  Expr
	Value Expr
}

func (node *Let) Tok() lexer.Token { return node.Token }
func (node *Let) Type() NodeType   { return LET }
func (node *Let) stmt()            {}

type Block struct {
	Token      lexer.Token
	Statements []Stmt
}

func (node *Block) Tok() lexer.Token { return node.Token }
func (node *Block) Type() NodeType   { return BLOCK }
func (node *Block) stmt()            {}

type For struct {
	Token lexer.Token
	Name  Expr
	Iter  Expr
	Stmt  Stmt
}

func (node *For) Tok() lexer.Token { return node.Token }
func (node *For) Type() NodeType   { return FOR }
func (node *For) stmt()            {}

type ExprStmt struct {
	Token lexer.Token
	Expr  Expr
}

func (node *ExprStmt) Tok() lexer.Token { return node.Token }
func (node *ExprStmt) Type() NodeType   { return EXPR_STMT }
func (node *ExprStmt) stmt()            {}

type Binary struct {
	Token lexer.Token
	Left  Expr
	Right Expr
}

func (node *Binary) Tok() lexer.Token { return node.Token }
func (node *Binary) Type() NodeType   { return BINARY }
func (node *Binary) expr()            {}

type And struct {
	Token lexer.Token
	Left  Expr
	Right Expr
}

func (node *And) Tok() lexer.Token { return node.Token }
func (node *And) Type() NodeType   { return AND }
func (node *And) expr()            {}

type Or struct {
	Token lexer.Token
	Left  Expr
	Right Expr
}

func (node *Or) Tok() lexer.Token { return node.Token }
func (node *Or) Type() NodeType   { return OR }
func (node *Or) expr()            {}

type Assign struct {
	Token lexer.Token
	Left  Expr
	Right Expr
}

func (node *Assign) Tok() lexer.Token { return node.Token }
func (node *Assign) Type() NodeType   { return ASSIGN }
func (node *Assign) expr()            {}

type Unary struct {
	Token lexer.Token
	Right Expr
}

func (node *Unary) Tok() lexer.Token { return node.Token }
func (node *Unary) Type() NodeType   { return UNARY }
func (node *Unary) expr()            {}

type Identifier struct {
	Token lexer.Token
}

func (node *Identifier) Tok() lexer.Token { return node.Token }
func (node *Identifier) Type() NodeType   { return IDENTIFIER }
func (node *Identifier) expr()            {}

type Literal struct {
	Token lexer.Token
}

func (node *Literal) Tok() lexer.Token { return node.Token }
func (node *Literal) Type() NodeType   { return LITERAL }
func (node *Literal) expr()            {}

func newModule(Token lexer.Token, Filename string, Statements []Stmt) *Module {
	return &Module{Token: Token, Filename: Filename, Statements: Statements}
}
func newLet(Token lexer.Token, Name Expr, Value Expr) *Let {
	return &Let{Token: Token, Name: Name, Value: Value}
}
func newBlock(Token lexer.Token, Statements []Stmt) *Block {
	return &Block{Token: Token, Statements: Statements}
}
func newFor(Token lexer.Token, Name Expr, Iter Expr, Stmt Stmt) *For {
	return &For{Token: Token, Name: Name, Iter: Iter, Stmt: Stmt}
}
func newExprStmt(Token lexer.Token, Expr Expr) *ExprStmt { return &ExprStmt{Token: Token, Expr: Expr} }
func newBinary(Token lexer.Token, Left Expr, Right Expr) *Binary {
	return &Binary{Token: Token, Left: Left, Right: Right}
}
func newAnd(Token lexer.Token, Left Expr, Right Expr) *And {
	return &And{Token: Token, Left: Left, Right: Right}
}
func newOr(Token lexer.Token, Left Expr, Right Expr) *Or {
	return &Or{Token: Token, Left: Left, Right: Right}
}
func newAssign(Token lexer.Token, Left Expr, Right Expr) *Assign {
	return &Assign{Token: Token, Left: Left, Right: Right}
}
func newUnary(Token lexer.Token, Right Expr) *Unary { return &Unary{Token: Token, Right: Right} }
func newIdentifier(Token lexer.Token) *Identifier   { return &Identifier{Token: Token} }
func newLiteral(Token lexer.Token) *Literal         { return &Literal{Token: Token} }
