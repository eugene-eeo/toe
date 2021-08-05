package eval

// Implements the actual evaluator for the language.

import (
	"fmt"
	"toe/lexer"
	"toe/parser"
)

type Context struct {
	env *Environment // current executing environment.
	// 'global' values
	_Object  Value // Object
	_nil     Value // nil
	_Boolean Value // Boolean
	_true    Value // true
	_false   Value // false
	_Number  Value // Number
	_String  Value // String
	_module  Value // module
}

func NewContext() *Context {
	ctx := &Context{}
	ctx.env = newEnvironment(nil)
	// bootstrap the object system
	ctx._Object = newObject(nil)
	ctx._nil = newNil(ctx._Object)
	ctx._Boolean = newObject(ctx._Object)
	ctx._true = newBoolean(ctx._Boolean, true)
	ctx._false = newBoolean(ctx._Boolean, false)
	ctx._Number = newObject(ctx._Object)
	ctx._String = newObject(ctx._Object)
	// todo: add any methods here.
	// define globals
	ctx.env.Define("Object", ctx._Object)
	ctx.env.Define("Boolean", ctx._Boolean)
	ctx.env.Define("Number", ctx._Number)
	ctx.env.Define("String", ctx._String)
	return ctx
}

func (ctx *Context) pushEnv() { ctx.env = newEnvironment(ctx.env) }
func (ctx *Context) popEnv()  { ctx.env = ctx.env.outer }

func (ctx *Context) EvalModule(module *parser.Module) Value {
	// First push a new environment with a module object.
	mod_obj := newObject(ctx._Object)
	ctx.pushEnv()
	ctx.env.Define("module", mod_obj)
	mod_obj.props["__file__"] = newString(ctx._String, module.Filename)
	for _, stmt := range module.Stmts {
		ctx.Eval(stmt)
	}
	ctx.popEnv()
	return mod_obj
}

func (ctx *Context) Eval(node parser.Node) Value {
	switch node := node.(type) {
	// Statements
	case *parser.ExprStmt:
		return ctx.Eval(node.Expr)
	case *parser.Let:
		switch left := node.Name.(type) {
		case *parser.Identifier:
			name := left.Token.Lexeme
			value := ctx.Eval(node.Value)
			if isError(value) {
				return value
			}
			ctx.env.Define(name, value)
			return ctx._nil
		}
	// Expressions
	case *parser.Identifier:
		ident := node.Tok().Lexeme
		rv, ok := ctx.env.Get(ident)
		if !ok {
			return &Error{ctx.newString(fmt.Sprintf("unknown identifier %s", ident))}
		}
		return rv
	case *parser.Assign:
		switch left := node.Left.(type) {
		case *parser.Identifier:
			name := left.Token.Lexeme
			value := ctx.Eval(node.Right)
			if isError(value) {
				return value
			}
			env := ctx.env.Resolve(name)
			if env == nil {
				return &Error{ctx.newString(fmt.Sprintf("unknown identifier %s", name))}
			}
			env.Define(name, value)
			return value
		}
	case *parser.Binary:
		left := ctx.Eval(node.Left)
		if isError(left) {
			return left
		}
		right := ctx.Eval(node.Right)
		if isError(right) {
			return right
		}
		return ctx.evalBinary(node.Token.Type, left, right)
	// Literals
	case *parser.Literal:
		switch node.Tok().Type {
		case lexer.NUMBER:
			return newNumber(ctx._Number, node.Token.Literal.(float64))
		case lexer.STRING:
			return newString(ctx._String, node.Token.Literal.(string))
		case lexer.NIL:
			return ctx._nil
		case lexer.TRUE:
			return ctx._true
		case lexer.FALSE:
			return ctx._false
		}
	}
	panic(fmt.Sprintf("not implemented yet: %#+v", node))
}

// Operators

func (ctx *Context) evalBinary(op lexer.TokenType, left Value, right Value) Value {
	switch op {
	case lexer.EQUAL_EQUAL:
		switch {
		case left == right: // fast pointer equality
			return ctx._true
		case left.Type() != right.Type():
			return ctx._false
		case left.Type() == NUMBER && right.Type() == NUMBER:
			return ctx.newBool(
				left.(*Number).value == right.(*Number).value,
			)
		case left.Type() == STRING && right.Type() == STRING:
			return ctx.newBool(
				left.(*String).value == right.(*String).value,
			)
		default:
			return ctx._false
		}
	}
	panic(fmt.Sprintf("not implemented yet: %#+v", op))
}

// =====
// Utils
// =====

func (ctx *Context) newBool(b bool) Value {
	if b {
		return ctx._true
	} else {
		return ctx._false
	}
}

func (ctx *Context) newString(lit string) Value {
	return newString(ctx._String, lit)
}

func isError(v Value) bool {
	_, ok := v.(*Error)
	return ok
}
