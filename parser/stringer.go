package parser

import (
	"bytes"
	"strings"
)

func (node *Module) String() string {
	stmts := make([]string, len(node.Stmts))
	for i, stmt := range node.Stmts {
		stmts[i] = stmt.String()
	}
	return strings.Join(stmts, "\n")
}

// Statements

func (node *Let) String() string {
	var buf bytes.Buffer
	buf.WriteString("let ")
	buf.WriteString(node.Name.Lexeme)
	buf.WriteString(" = ")
	buf.WriteString(node.Value.String())
	buf.WriteString(";")
	return buf.String()
}

func (node *For) String() string {
	var buf bytes.Buffer
	buf.WriteString("for (")
	buf.WriteString(node.Name.Lexeme)
	buf.WriteString(" : ")
	buf.WriteString(node.Iter.String())
	buf.WriteString(") ")
	buf.WriteString(node.Stmt.String())
	return buf.String()
}

func (node *While) String() string {
	var buf bytes.Buffer
	buf.WriteString("while (")
	buf.WriteString(node.Cond.String())
	buf.WriteString(") ")
	buf.WriteString(node.Stmt.String())
	return buf.String()
}

func (node *If) String() string {
	var buf bytes.Buffer
	buf.WriteString("if (")
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
	buf.WriteString("{")
	for _, stmt := range node.Stmts {
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

func (node *Break) String() string    { return node.Keyword.Lexeme + ";" }
func (node *Continue) String() string { return node.Keyword.Lexeme + ";" }

func (node *Return) String() string {
	var buf bytes.Buffer
	buf.WriteString(node.Keyword.Lexeme)
	if node.Expr != nil {
		buf.WriteString(" ")
		buf.WriteString(node.Expr.String())
	}
	buf.WriteString(";")
	return buf.String()
}

// Expressions

func (node *Assign) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(node.Name.Lexeme)
	buf.WriteString(" = ")
	buf.WriteString(node.Right.String())
	buf.WriteString(")")
	return buf.String()
}

func (node *Binary) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(node.Left.String())
	buf.WriteString(" ")
	buf.WriteString(node.Op.Lexeme)
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
	buf.WriteString(node.Op.Lexeme)
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
	buf.WriteString(node.Op.Lexeme)
	buf.WriteString(" ")
	buf.WriteString(node.Right.String())
	buf.WriteString(")")
	return buf.String()
}

func (node *Get) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(node.Object.String())
	if node.Bound {
		buf.WriteString(".")
	} else {
		buf.WriteString("->")
	}
	buf.WriteString(node.Name.Lexeme)
	buf.WriteString(")")
	return buf.String()
}

func (node *Set) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(node.Object.String())
	if node.Bound {
		buf.WriteString(".")
	} else {
		buf.WriteString("->")
	}
	buf.WriteString(node.Name.Lexeme)
	buf.WriteString(" = ")
	buf.WriteString(node.Right.String())
	buf.WriteString(")")
	return buf.String()
}

func (node *Unary) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(node.Op.Lexeme)
	buf.WriteString(node.Right.String())
	buf.WriteString(")")
	return buf.String()
}

func (node *Call) String() string {
	var buf bytes.Buffer
	args := make([]string, len(node.Args))
	for i, arg := range node.Args {
		args[i] = arg.String()
	}
	buf.WriteString("(")
	buf.WriteString(node.Callee.String())
	buf.WriteString(node.LParen.Lexeme)
	buf.WriteString(strings.Join(args, ", "))
	buf.WriteString(")")
	buf.WriteString(")")
	return buf.String()
}

func (node *Identifier) String() string { return node.Id.Lexeme }
func (node *Literal) String() string    { return node.Lit.Lexeme }

func (node *Array) String() string {
	var buf bytes.Buffer
	exprs := make([]string, len(node.Exprs))
	for i, expr := range node.Exprs {
		exprs[i] = expr.String()
	}
	buf.WriteString("[")
	buf.WriteString(strings.Join(exprs, ", "))
	buf.WriteString("]")
	return buf.String()
}

func (node *Function) String() string {
	var buf bytes.Buffer
	params := make([]string, len(node.Params))
	for i, tok := range node.Params {
		params[i] = tok.Lexeme
	}
	buf.WriteString(node.Fn.Lexeme)
	buf.WriteString("(")
	buf.WriteString(strings.Join(params, ", "))
	buf.WriteString(")")
	buf.WriteString(node.Body.String())
	return buf.String()
}

func (node *Super) String() string {
	var buf bytes.Buffer
	buf.WriteString("(")
	buf.WriteString(node.Tok.Lexeme)
	if node.Bound {
		buf.WriteString(".")
	} else {
		buf.WriteString("->")
	}
	buf.WriteString(node.Name.Lexeme)
	buf.WriteString(")")
	return buf.String()
}
