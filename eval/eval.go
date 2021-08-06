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
	_Object   *Object  // Object
	_Nil      *Object  // Nil (prototype of nil)
	_nil      *Nil     // nil
	_Boolean  *Object  // Boolean
	_Function *Object  // Function
	_true     *Boolean // true
	_false    *Boolean // false
	_Number   *Object  // Number
	_String   *Object  // String
}

func NewContext() *Context {
	ctx := &Context{}
	ctx.env = newEnvironment(nil)
	// bootstrap the object system
	ctx._Object = newObject(nil)
	ctx._Nil = newObject(ctx._Object) // even nil is an Object
	ctx._nil = &Nil{}
	ctx._Boolean = newObject(ctx._Object)
	ctx._Function = newObject(ctx._Object)
	ctx._true = &Boolean{true}
	ctx._false = &Boolean{false}
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
	mod_obj.props["__file__"] = &String{module.Filename}
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
			return &Error{&String{fmt.Sprintf("unknown identifier %s", ident)}}
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
				return &Error{&String{fmt.Sprintf("unknown identifier %s", name)}}
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
			return &Number{node.Token.Literal.(float64)}
		case lexer.STRING:
			return &String{node.Token.Literal.(string)}
		case lexer.NIL:
			return ctx._nil
		case lexer.TRUE:
			return ctx._true
		case lexer.FALSE:
			return ctx._false
		}
	}
	return &Error{&String{fmt.Sprintf("not implemented yet: %#+v", node)}}
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

func isError(v Value) bool {
	_, ok := v.(*Error)
	return ok
}
