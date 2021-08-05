package parser

import (
	"bytes"
	"strings"
)

// TODO: find out some way we could automate this.

func (node *Module) String() string {
	stmts := []string{}
	for _, stmt := range node.Stmts {
		stmts = append(stmts, stmt.String())
	}
	return strings.Join(stmts, "\n")
}

// Statements

func (node *Let) String() string {
	var buf bytes.Buffer
	buf.WriteString(node.Tok().Lexeme)
	buf.WriteString(" ")
	buf.WriteString(node.Name.String())
	buf.WriteString(" = ")
	buf.WriteString(node.Value.String())
	buf.WriteString(";")
	return buf.String()
}

func (node *For) String() string {
	var buf bytes.Buffer
	buf.WriteString(node.Tok().Lexeme)
	buf.WriteString(" (")
	buf.WriteString(node.Name.String())
	buf.WriteString(" : ")
	buf.WriteString(node.Iter.String())
	buf.WriteString(") ")
	buf.WriteString(node.Stmt.String())
	return buf.String()
}

func (node *While) String() string {
	var buf bytes.Buffer
	buf.WriteString(node.Tok().Lexeme)
	buf.WriteString(" (")
	buf.WriteString(node.Cond.String())
	buf.WriteString(") ")
	buf.WriteString(node.Stmt.String())
	return buf.String()
}

func (node *If) String() string {
	var buf bytes.Buffer
	buf.WriteString(node.Tok().Lexeme)
	buf.WriteString(" (")
	buf.WriteString(node.Cond.String())
	buf.WriteString(") ")
	buf.WriteString(node.Then.String())
	if node.Else != nil {
		buf.WriteString(" else ")
		buf.WriteString(node.Else.String())
	}
	return buf.String()
}

func (node *Block) String() string {
	var buf bytes.Buffer
	buf.WriteString(node.Tok().Lexeme)
	for _, stmt := range node.Statements {
		buf.WriteString(stmt.String())
	}
	buf.WriteString("}")
	return buf.String()
}

func (node *ExprStmt) String() string {
	var buf bytes.Buffer
	buf.WriteString(node.Expr.String())
	buf.WriteString(";")
	return buf.String()
}

func (node *Break) String() string    { return node.Tok().Lexeme + ";" }
func (node *Continue) String() string { return node.Tok().Lexeme + ";" }

// Expressions

func (node *Assign) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(node.Left.String())
	buf.WriteString(" ")
	buf.WriteString(node.Tok().Lexeme)
	buf.WriteString(" ")
	buf.WriteString(node.Right.String())
	buf.WriteString(")")
	return buf.String()
}

func (node *Binary) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(node.Left.String())
	buf.WriteString(" ")
	buf.WriteString(node.Tok().Lexeme)
	buf.WriteString(" ")
	buf.WriteString(node.Right.String())
	buf.WriteString(")")
	return buf.String()
}

func (node *And) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(node.Left.String())
	buf.WriteString(" ")
	buf.WriteString(node.Tok().Lexeme)
	buf.WriteString(" ")
	buf.WriteString(node.Right.String())
	buf.WriteString(")")
	return buf.String()
}

func (node *Or) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(node.Left.String())
	buf.WriteString(" ")
	buf.WriteString(node.Tok().Lexeme)
	buf.WriteString(" ")
	buf.WriteString(node.Right.String())
	buf.WriteString(")")
	return buf.String()
}

func (node *Unary) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(node.Tok().Lexeme)
	buf.WriteString(node.Right.String())
	buf.WriteString(")")
	return buf.String()
}

func (node *Identifier) String() string { return node.Tok().Lexeme }
func (node *Literal) String() string    { return node.Tok().Lexeme }
