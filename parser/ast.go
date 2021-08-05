package parser

import "toe/lexer"

type NodeType uint8

type Node interface {
	Tok() lexer.Token
	Type() NodeType
	String() string
}

type Expr interface {
	Node
	expr()
}

type Stmt interface {
	Node
	stmt()
}

// Module represents a file containing the program.
type Module struct {
	Filename string
	Stmts    []Stmt
}
