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

type globals struct {
	Object   *Object // Object
	Nil      *Object // Nil (prototype of nil)
	Boolean  *Object // Boolean
	Function *Object // Function
	Number   *Object // Number
	String   *Object // String
	Array    *Object // Array
	Error    *Object // Error
}

type Context struct {
	Env     *Environment        // current executing environment.
	locs    map[parser.Expr]int // map of resolvable expressions to distance.
	funcs   []string            // list of function calls we're currently executing, for error tracing.
	this    Value               // `this' of the currently executing function, if any.
	globals *globals            // 'global' values
}

func NewContext(locs map[parser.Expr]int) *Context {
	ctx := &Context{
		locs:    locs,
		funcs:   []string{},
		globals: &globals{},
	}
	// bootstrap the object system
	ctx.globals.Object = newObject(nil)
	ctx.globals.Nil = newObject(ctx.globals.Object) // even nil is an Object
	ctx.globals.Boolean = newObject(ctx.globals.Object)
	ctx.globals.Function = newObject(ctx.globals.Object)
	ctx.globals.Number = newObject(ctx.globals.Object)
	ctx.globals.String = newObject(ctx.globals.Object)
	ctx.globals.Array = newObject(ctx.globals.Object)
	// methods
	ctx.globals.Object.props["clone"] = &Builtin{fn: _Object_clone}
	ctx.globals.Object.props["inspect"] = &Builtin{fn: _Object_inspect}

	ctx.globals.Nil.props["inspect"] = &Builtin{fn: _Nil_inspect}

	ctx.globals.Boolean.props["inspect"] = &Builtin{fn: _Boolean_inspect}

	ctx.globals.Function.props["bind"] = &Builtin{fn: _Function_bind}
	ctx.globals.Function.props["inspect"] = &Builtin{fn: _Function_inspect}

	ctx.globals.Number.props["inspect"] = &Builtin{fn: _Number_inspect}

	ctx.globals.String.props["length"] = &Builtin{fn: _String_length}
	ctx.globals.String.props["inspect"] = &Builtin{fn: _String_inspect}
	return ctx
}

func (ctx *Context) popEnv()  { ctx.Env = ctx.Env.outer }
func (ctx *Context) pushEnv() { ctx.Env = newEnvironment(ctx.Env.filename, ctx.Env) }

func (ctx *Context) NewModuleEnv(filename string) (*Environment, Value) {
	new_env := newEnvironment(filename, nil)
	mod_obj := newObject(ctx.globals.Object)
	mod_obj.props["filename"] = &String{filename}
	mod_obj.props["exports"] = newObject(ctx.globals.Object)
	new_env.Define("puts", &Builtin{fn: func(ctx *Context, this Value, args []Value) Value {
		for _, x := range args {
			fmt.Println(x)
		}
		return NIL
	}})
	new_env.Define("module", mod_obj)
	new_env.Define("Object", ctx.globals.Object)
	new_env.Define("Boolean", ctx.globals.Boolean)
	new_env.Define("Number", ctx.globals.Number)
	new_env.Define("String", ctx.globals.String)
	new_env.Define("Array", ctx.globals.Array)
	new_env.Define("Function", ctx.globals.Function)
	return new_env, mod_obj
}

func (ctx *Context) EvalModule(module *parser.Module) Value {
	new_env, mod_obj := ctx.NewModuleEnv(module.Filename)
	og_env := ctx.Env
	ctx.Env = new_env
	ctx.pushFunc("<module>")
	for _, stmt := range module.Stmts {
		v := ctx.Eval(stmt)
		if isError(v) {
			return v
		}
	}
	ctx.popFunc()
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
		ctx.Eval(node.Expr)
		return NIL
	case *parser.Break:
		return &Break{}
	case *parser.Continue:
		return &Continue{}
	case *parser.Return:
		return ctx.evalReturn(node)
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
	case *parser.Call:
		return ctx.evalCall(node)
	case *parser.Identifier:
		return ctx.evalIdentifier(node)
	case *parser.Function:
		return ctx.evalFunction(node)
	case *parser.Super:
		return ctx.evalSuper(node)
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
	return ctx.err(&String{fmt.Sprintf("not implemented yet: %#+v", node)})
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
	rv := ctx.evalBinaryValues(node.Token.Type, left, right)
	if isError(rv) {
		rv.(*Error).addContext(node.Token)
	}
	return rv
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
	rv := ctx.evalUnaryValues(node.Tok().Type, right)
	if isError(rv) {
		rv.(*Error).addContext(node.Token)
	}
	return rv
}

func (ctx *Context) evalGet(node *parser.Get) Value {
	object := ctx.Eval(node.Object)
	if isError(object) {
		return object
	}
	attr := node.Name.Lexeme
	v, ok := ctx.getAttr(object, attr)
	if !ok {
		e := ctx.err(&String{fmt.Sprintf("attribute not found: %q", attr)})
		e.addContext(node.Token)
		return e
	}
	if node.Bound {
		v = ctx.bind(v, object)
	}
	return v
}

func (ctx *Context) evalSet(node *parser.Set) Value {
	object := ctx.Eval(node.Object)
	if isError(object) {
		return object
	}
	attr := node.Name.Lexeme
	value := ctx.Eval(node.Right)
	if isError(value) {
		return value
	}
	rv, ok := ctx.setAttr(object, attr, value)
	if !ok {
		e := ctx.err(&String{fmt.Sprintf("cannot set attribute %q", attr)})
		e.addContext(node.Token)
		return e
	}
	if node.Bound {
		rv = ctx.bind(rv, object)
	}
	return rv
}

func (ctx *Context) evalCall(node *parser.Call) Value {
	fn := ctx.Eval(node.Fn)
	if isError(fn) {
		return fn
	}
	args := make([]Value, len(node.Args))
	for i, arg := range node.Args {
		args[i] = ctx.Eval(arg)
		if isError(args[i]) {
			return args[i]
		}
	}
	rv, ok := ctx.callFunction(fn, args)
	if !ok {
		e := ctx.err(&String{fmt.Sprintf("not a function")})
		e.addContext(node.Token)
		return e
	}
	// unwrap
	if isReturn(rv) {
		rv = rv.(*Return).value
	}
	if isError(rv) {
		rv.(*Error).addContext(node.Token)
	}
	return rv
}

func (ctx *Context) evalIdentifier(node *parser.Identifier) Value {
	ident := node.Tok().Lexeme
	rv, ok := ctx.Env.GetAt(ctx.locs[node], ident)
	if !ok {
		e := ctx.err(&String{fmt.Sprintf("unknown identifier %s", ident)})
		e.addContext(node.Token)
		return e
	}
	return rv
}

func (ctx *Context) evalFunction(node *parser.Function) Value {
	return &Function{
		this:    nil,
		node:    node,
		closure: ctx.Env,
	}
}

func (ctx *Context) evalSuper(node *parser.Super) Value {
	proto := ctx.getPrototype(ctx.this)
	v, ok := ctx.getAttr(proto, node.Name.Lexeme)
	if !ok {
		e := ctx.err(&String{fmt.Sprintf("attribute not found: %q", node.Name.Lexeme)})
		e.addContext(node.Name)
		return e
	}
	if node.Bound {
		v = ctx.bind(v, ctx.this)
	}
	return v
}

// ==========
// Statements
// ==========

func (ctx *Context) evalBlock(node *parser.Block) Value {
	// Blocks evaluate to the return-value of the last statement in the block.
	// Where we encounter continue / break, we will return that signal.
	ctx.pushEnv()
	defer ctx.popEnv()
	for _, stmt := range node.Statements {
		rv := ctx.Eval(stmt)
		if isError(rv) || isBreak(rv) || isContinue(rv) || isReturn(rv) {
			return rv
		}
	}
	return NIL
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
		return ctx.err(&String{fmt.Sprintf("not iterable")})
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
		if isReturn(res) || isError(res) {
			loopRv = res
			break
		} else if isBreak(res) {
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
		if isReturn(round) || isError(round) {
			return round
		} else if isBreak(round) {
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

func (ctx *Context) evalReturn(node *parser.Return) Value {
	rv := Value(NIL)
	if node.Expr != nil {
		rv = ctx.Eval(node.Expr)
	}
	if isError(rv) {
		return rv
	}
	return &Return{rv}
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

func isTruthy(v Value) bool   { return v != FALSE && v != NIL }
func isReturn(v Value) bool   { return v.Type() == RETURN }
func isBreak(v Value) bool    { return v.Type() == BREAK }
func isContinue(v Value) bool { return v.Type() == CONTINUE }
func isError(v Value) bool    { return v.Type() == ERROR }
