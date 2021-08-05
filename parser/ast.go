package parser

import "toe/lexer"

type NodeType uint8

type Node interface {
	Tok() lexer.Token
	Type() NodeType
}

type Expr interface {
	Node
	expr()
}

type Stmt interface {
	Node
	stmt()
}
