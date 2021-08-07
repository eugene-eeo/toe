package eval

// Implements the actual evaluator for the language.

import (
	"fmt"
	"toe/lexer"
	"toe/parser"
)

var (
	NIL   = &Nil{}
	TRUE  = &Boolean{true}
	FALSE = &Boolean{false}
)

type Context struct {
	Env  *Environment        // current executing environment.
	locs map[parser.Expr]int // map of resolvable expressions to distance.
	// 'global' values
	_Object   *Object // Object
	_Nil      *Object // Nil (prototype of nil)
	_Boolean  *Object // Boolean
	_Function *Object // Function
	_Number   *Object // Number
	_String   *Object // String
	_Array    *Object
}

func NewContext(locs map[parser.Expr]int) *Context {
	ctx := &Context{}
	// bootstrap the object system
	ctx._Object = newObject(nil)
	ctx._Nil = newObject(ctx._Object) // even nil is an Object
	ctx._Boolean = newObject(ctx._Object)
	ctx._Function = newObject(ctx._Object)
	ctx._Number = newObject(ctx._Object)
	ctx._String = newObject(ctx._Object)
	ctx._Array = newObject(ctx._Object)
	// todo: add methods here.
	return ctx
}

func (ctx *Context) popEnv() { ctx.Env = ctx.Env.outer }
func (ctx *Context) pushEnv() *Environment {
	ctx.Env = newEnvironment(ctx.Env)
	return ctx.Env
}

func (ctx *Context) NewModuleEnv(filename string) (*Environment, Value) {
	new_env := newEnvironment(nil)
	mod_obj := newObject(ctx._Object)
	mod_obj.props["filename"] = &String{filename}
	mod_obj.props["exports"] = newObject(ctx._Object)
	// new_env.Define("puts", &Builtin{})
	new_env.Define("module", mod_obj)
	new_env.Define("Object", ctx._Object)
	new_env.Define("Boolean", ctx._Boolean)
	new_env.Define("Number", ctx._Number)
	new_env.Define("String", ctx._String)
	new_env.Define("Array", ctx._Array)
	return new_env, mod_obj
}

func (ctx *Context) EvalModule(module *parser.Module) Value {
	new_env, mod_obj := ctx.NewModuleEnv(module.Filename)
	og_env := ctx.Env
	ctx.Env = new_env
	for _, stmt := range module.Stmts {
		v := ctx.Eval(stmt)
		if isError(v) {
			return v
		}
	}
	ctx.Env = og_env
	return mod_obj
}

func (ctx *Context) Eval(node parser.Node) Value {
	switch node := node.(type) {
	// Statements
	case *parser.Let:
		name := node.Name.Lexeme
		value := ctx.Eval(node.Value)
		if isError(value) {
			return value
		}
		ctx.Env.Define(name, value)
		return NIL
	case *parser.Block:
		return ctx.evalBlock(node)
	case *parser.For:
		return ctx.evalFor(node)
	case *parser.While:
		return ctx.evalWhile(node)
	case *parser.If:
		return ctx.evalIf(node)
	case *parser.ExprStmt:
		return ctx.Eval(node.Expr)
	case *parser.Break:
		return &Break{}
	case *parser.Continue:
		return &Continue{}
	// Expressions
	case *parser.Binary:
		return ctx.evalBinary(node)
	case *parser.And:
		return ctx.evalAnd(node)
	case *parser.Or:
		return ctx.evalOr(node)
	case *parser.Assign:
		name := node.Name.Lexeme
		value := ctx.Eval(node.Right)
		if isError(value) {
			return value
		}
		ctx.Env.Ancestor(ctx.locs[node]).Define(name, value)
		return value
	case *parser.Unary:
		return ctx.evalUnary(node)
	case *parser.Get:
		return ctx.evalGet(node)
	case *parser.Set:
		return ctx.evalSet(node)
	case *parser.Identifier:
		return ctx.evalIdentifier(node)
	// Literals
	case *parser.Literal:
		switch node.Tok().Type {
		case lexer.NUMBER:
			return &Number{node.Token.Literal.(float64)}
		case lexer.STRING:
			return &String{node.Token.Literal.(string)}
		case lexer.NIL:
			return NIL
		case lexer.TRUE:
			return TRUE
		case lexer.FALSE:
			return FALSE
		}
	}
	return &Error{&String{fmt.Sprintf("not implemented yet: %#+v", node)}}
}

// ===========
// Expressions
// ===========

func (ctx *Context) evalBinary(node *parser.Binary) Value {
	left := ctx.Eval(node.Left)
	if isError(left) {
		return left
	}
	right := ctx.Eval(node.Right)
	if isError(right) {
		return right
	}
	return ctx.evalBinaryValues(node.Token.Type, left, right)
}

func (ctx *Context) evalAnd(node *parser.And) Value {
	left := ctx.Eval(node.Left)
	if isError(left) || !isTruthy(left) {
		return left
	}
	return ctx.Eval(node.Right)
}

func (ctx *Context) evalOr(node *parser.Or) Value {
	left := ctx.Eval(node.Left)
	if isError(left) || isTruthy(left) {
		return left
	}
	return ctx.Eval(node.Right)
}

func (ctx *Context) evalUnary(node *parser.Unary) Value {
	right := ctx.Eval(node.Right)
	if isError(right) {
		return right
	}
	return ctx.evalUnaryValues(node.Tok().Type, right)
}

func (ctx *Context) evalGet(node *parser.Get) Value {
	object := ctx.Eval(node.Object)
	if isError(object) {
		return object
	}
	attr := node.Name.Lexeme
	v, ok := ctx.getAttr(object, attr)
	if !ok {
		return &Error{&String{
			fmt.Sprintf("attribute not found: %q", attr),
		}}
	}
	if node.Bound {
		v = ctx.bind(v, object)
	}
	return v
}

func (ctx *Context) evalSet(node *parser.Set) Value {
	object := ctx.Eval(node.Object)
	if isError(object) {
		return object;
	}
	attr := node.Name.Lexeme
	value := ctx.Eval(node.Right)
	if isError(value) {
		return value
	}
	rv, ok := ctx.setAttr(object, attr, value)
	if !ok {
		return &Error{&String{
			fmt.Sprintf("cannot set attribute %q", attr),
		}}
	}
	if node.Bound {
		rv = ctx.bind(rv, object)
	}
	return rv
}

func (ctx *Context) evalIdentifier(node *parser.Identifier) Value {
	ident := node.Tok().Lexeme
	rv, ok := ctx.Env.GetAt(ctx.locs[node], ident)
	if !ok {
		return &Error{&String{fmt.Sprintf("unknown identifier %s", ident)}}
	}
	return rv
}

// ==========
// Statements
// ==========

func (ctx *Context) evalBlock(node *parser.Block) Value {
	// Blocks evaluate to the return-value of the last statement in the block.
	// Where we encounter continue / break, we will return that signal.
	var rv Value = NIL
	ctx.pushEnv()
	defer ctx.popEnv()
	for _, stmt := range node.Statements {
		rv = ctx.Eval(stmt)
		if isError(rv) || isBreak(rv) || isContinue(rv) {
			return rv
		}
	}
	return rv
}

func (ctx *Context) evalFor(node *parser.For) Value {
	// In theory we would need a stack for iterators,
	// but the call stack helps us handle it.
	it := ctx.Eval(node.Iter)
	if isError(it) {
		return it
	}
	iter, ok := ctx.getIterator(it)
	if !ok {
		// havent found an iterator?
		return &Error{&String{fmt.Sprintf("not iterable")}}
	}
	loopRv := Value(NIL)
	loopVar := node.Name.Tok().Lexeme
	ctx.pushEnv()
	defer ctx.popEnv()
	for {
		done := iter.Done()
		if isError(done) {
			loopRv = done
			break
		}
		if isTruthy(done) {
			break
		}
		next := iter.Next()
		if isError(next) {
			loopRv = next
			break
		}
		ctx.Env.Define(loopVar, next)
		res := ctx.Eval(node.Stmt)
		switch {
		case isError(res):
			loopRv = res
			break
		case isBreak(res):
			break
		}
	}
	// we _always_ evaluate iter.End()
	if v := iter.End(); isError(v) {
		return v
	}
	return loopRv
}

func (ctx *Context) evalWhile(node *parser.While) Value {
	for {
		cond := ctx.Eval(node.Cond)
		if isError(cond) {
			return cond
		}
		if !isTruthy(cond) {
			break
		}
		round := ctx.Eval(node.Stmt)
		switch {
		case isError(round):
			return round
		case isBreak(round):
			break
		}
	}
	return NIL
}

func (ctx *Context) evalIf(node *parser.If) Value {
	cond := ctx.Eval(node.Cond)
	if isError(cond) {
		return cond
	}
	if isTruthy(cond) {
		return ctx.Eval(node.Then)
	} else {
		if node.Else != nil {
			return ctx.Eval(node.Else)
		}
		return NIL
	}
}

// =====
// Utils
// =====

func newBool(b bool) *Boolean {
	if b {
		return TRUE
	} else {
		return FALSE
	}
}

func isBreak(v Value) bool    { return v.Type() == BREAK }
func isContinue(v Value) bool { return v.Type() == CONTINUE }
func isError(v Value) bool    { return v.Type() == ERROR }
func isTruthy(v Value) bool   { return v != FALSE && v != NIL }
