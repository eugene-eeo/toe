package parser

import "toe/lexer"

type NodeType uint8

type Node interface {
	Tok() lexer.Token
	Type() NodeType
}

type Expr interface { expr() }
type Stmt interface { stmt() }
