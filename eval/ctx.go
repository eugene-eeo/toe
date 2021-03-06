package eval

import (
	"fmt"
	"toe/lexer"
	"toe/parser"
)

type Context struct {
	// stack contains the current call stack. we consult the call-stack to tell
	// us which function we're in, and augment that using an expression's token.
	stack []callStackEntry
	// the current environment we're executing.
	env *environment
	// the object in which the currently executing function is found,
	// and the current `this` -- these are required to implement super.
	whence, this Value
	// for hash tables
	ht_seed uint64
	// object model
	globals *Globals
}

func NewContext() *Context {
	return &Context{
		stack:   make([]callStackEntry, 0, 8),
		ht_seed: getNewHashTableSeed(),
		globals: newGlobals(),
	}
}

func (ctx *Context) pushEnv() { ctx.env = newEnv(ctx.env) }
func (ctx *Context) popEnv()  { ctx.env = ctx.env.outer }

func (ctx *Context) pushFunc(e callStackEntry) { ctx.stack = append(ctx.stack, e) }
func (ctx *Context) popFunc()                  { ctx.stack = ctx.stack[:len(ctx.stack)-1] }

func (ctx *Context) EvalStmt(node parser.Stmt) Value {
	switch node := node.(type) {
	case *parser.Module:
		return ctx.evalModule(node)
	case *parser.Let:
		return ctx.evalLet(node)
	case *parser.Block:
		return ctx.evalBlock(node)
	case *parser.For:
		return ctx.evalFor(node)
	case *parser.While:
		return ctx.evalWhile(node)
	case *parser.If:
		return ctx.evalIf(node)
	case *parser.ExprStmt:
		return ctx.EvalExpr(node.Expr)
	case *parser.Break:
		return BREAK
	case *parser.Continue:
		return CONTINUE
	case *parser.Return:
		return ctx.evalReturn(node)
	}
	panic(fmt.Sprintf("unhandled node %#+v", node))
}

func (ctx *Context) EvalExpr(node parser.Expr) Value {
	switch node := node.(type) {
	case *parser.Binary:
		return ctx.evalBinary(node)
	case *parser.And:
		return ctx.evalAnd(node)
	case *parser.Or:
		return ctx.evalOr(node)
	case *parser.Assign:
		return ctx.evalAssign(node)
	case *parser.Unary:
		return ctx.evalUnary(node)
	case *parser.Get:
		return ctx.evalGet(node)
	case *parser.Set:
		return ctx.evalSet(node)
	case *parser.Method:
		return ctx.evalMethod(node)
	case *parser.Call:
		return ctx.evalCall(node)
	case *parser.Identifier:
		return ctx.evalIdentifier(node)
	case *parser.Literal:
		switch node.Lit.Type {
		case lexer.STRING:
			return String(node.Lit.Literal.(string))
		case lexer.NUMBER:
			return Number(node.Lit.Literal.(float64))
		case lexer.NIL:
			return NIL
		case lexer.TRUE:
			return TRUE
		case lexer.FALSE:
			return FALSE
		}
	case *parser.Array:
		return ctx.evalArray(node)
	case *parser.Hash:
		return ctx.evalHash(node)
	case *parser.Function:
		return ctx.evalFunction(node)
	case *parser.Super:
		return ctx.evalSuper(node)
	}
	panic(fmt.Sprintf("unhandled node %#+v", node))
}

// ==========
// Statements
// ==========

func (ctx *Context) evalModule(module *parser.Module) Value {
	ctx.pushEnv()
	ctx.globals.addToEnv(ctx.env)
	ctx.pushFunc(&moduleCse{module.Filename})
	for _, stmt := range module.Stmts {
		rv := ctx.EvalStmt(stmt)
		if isError(rv) {
			return rv
		}
	}
	ctx.popFunc()
	ctx.popEnv()
	return NIL
}

func (ctx *Context) evalLet(node *parser.Let) Value {
	name := node.Name.Lexeme
	value := ctx.EvalExpr(node.Value)
	if isError(value) {
		return value
	}
	ctx.env.set(name, value)
	return NIL
}

func (ctx *Context) evalBlock(node *parser.Block) Value {
	var rv = Value(NIL)
	ctx.pushEnv()
	for _, stmt := range node.Stmts {
		rv = ctx.EvalStmt(stmt)
	}
	ctx.popEnv()
	return rv
}

func (ctx *Context) evalFor(node *parser.For) Value {
	iter_obj := ctx.EvalExpr(node.Iter)
	if isError(iter_obj) {
		return iter_obj
	}
	iterator, ok := getIterator(iter_obj)
	if !ok {
		e := newError(ctx, String("not an iterable"))
		return ctx.addErrorStack(e, node.Keyword)
	}
	loop_var := node.Name.Lexeme
	loop_rv := Value(NIL)
	ctx.pushEnv()
	env := ctx.env
	for {
		// while (!it.done())
		done := iterator.Done()
		if isError(done) {
			loop_rv = done
			break
		}
		if isTruthy(done) {
			break
		}
		// let ? = it.next()
		next := iterator.Next()
		if isError(next) {
			loop_rv = next
			break
		}
		env.set(loop_var, next)
		signal := ctx.EvalStmt(node.Stmt)
		if isBreak(signal) {
			break
		}
		if isError(signal) || isReturn(signal) {
			loop_rv = signal
			break
		}
	}
	ctx.popEnv()
	// always call the .Close method, to allow for cleanup
	if v := iterator.Done(); isError(v) {
		return v
	}
	return loop_rv
}

func (ctx *Context) evalWhile(node *parser.While) Value {
	for {
		cond := ctx.EvalExpr(node.Cond)
		if isError(cond) {
			return cond
		}
		if !isTruthy(cond) {
			break
		}
		rv := ctx.EvalStmt(node.Stmt)
		if isBreak(rv) {
			break
		}
		if isError(rv) || isReturn(rv) {
			return rv
		}
	}
	return NIL
}

func (ctx *Context) evalIf(node *parser.If) Value {
	cond := ctx.EvalExpr(node.Cond)
	if isError(cond) {
		return cond
	} else if isTruthy(cond) {
		return ctx.EvalStmt(node.Then)
	} else {
		if node.Else != nil {
			return ctx.EvalStmt(node.Else)
		}
		return NIL
	}
}

func (ctx *Context) evalReturn(node *parser.Return) Value {
	if node.Expr == nil {
		return Return{NIL}
	}
	v := ctx.EvalExpr(node.Expr)
	if isError(v) {
		return v
	}
	return Return{v}
}

// ===========
// Expressions
// ===========

func (ctx *Context) evalBinary(node *parser.Binary) Value {
	left := ctx.EvalExpr(node.Left)
	if isError(left) {
		return left
	}
	right := ctx.EvalExpr(node.Right)
	if isError(right) {
		return right
	}
	rv := ctx.binary(node.Op.Lexeme, left, right)
	if isError(rv) {
		return ctx.addErrorStack(rv.(*Error), node.Op)
	}
	return rv
}

func (ctx *Context) evalAnd(node *parser.And) Value {
	left := ctx.EvalExpr(node.Left)
	if isError(left) || !isTruthy(left) {
		return left
	}
	return ctx.EvalExpr(node.Right)
}

func (ctx *Context) evalOr(node *parser.Or) Value {
	left := ctx.EvalExpr(node.Left)
	if isError(left) || isTruthy(left) {
		return left
	}
	return ctx.EvalExpr(node.Right)
}

func (ctx *Context) evalAssign(node *parser.Assign) Value {
	right := ctx.EvalExpr(node.Right)
	if isError(right) {
		return right
	}
	name := node.Name.Lexeme
	ctx.env.ancestor(node.Loc).set(name, right)
	return right
}

func (ctx *Context) evalUnary(node *parser.Unary) Value {
	right := ctx.EvalExpr(node.Right)
	if isError(right) {
		return right
	}
	rv := ctx.unary(node.Op.Type, right)
	if isError(rv) {
		return ctx.addErrorStack(rv.(*Error), node.Op)
	}
	return rv
}

func (ctx *Context) evalGet(node *parser.Get) Value {
	object := ctx.EvalExpr(node.Object)
	if isError(object) {
		return object
	}
	if isSuper(object) {
		object = object.(Super).proto
	}
	var tmp Value
	rv := ctx.getSlot(object, node.Name.Lexeme, &tmp)
	if isError(rv) {
		return ctx.addErrorStack(rv.(*Error), node.Name)
	}
	return rv
}

func (ctx *Context) evalSet(node *parser.Set) Value {
	right := ctx.EvalExpr(node.Right)
	if isError(right) {
		return right
	}
	object := ctx.EvalExpr(node.Object)
	if isError(object) {
		return object
	}
	if isSuper(object) {
		object = object.(Super).proto
	}
	rv := ctx.setSlot(object, node.Name.Lexeme, right)
	if isError(rv) {
		return ctx.addErrorStack(rv.(*Error), node.Name)
	}
	return rv
}

func (ctx *Context) evalMethod(node *parser.Method) Value {
	object := ctx.EvalExpr(node.Object)
	if isError(object) {
		return object
	}
	this := object
	var whence Value
	if isSuper(object) {
		object = object.(Super).proto
		this = ctx.this
	}
	fn := ctx.getSlot(object, node.Name.Lexeme, &whence)
	if isError(fn) {
		return ctx.addErrorStack(fn.(*Error), node.Name)
	}
	args := make([]Value, len(node.Args))
	for i, expr_node := range node.Args {
		expr := ctx.EvalExpr(expr_node)
		if isError(expr) {
			return expr
		}
		args[i] = expr
	}
	rv := ctx.call(whence, fn, this, args)
	if isError(rv) {
		ctx.addErrorStack(rv.(*Error), node.LParen)
	}
	return rv
}

func (ctx *Context) evalCall(node *parser.Call) Value {
	callee := ctx.EvalExpr(node.Callee)
	if isError(callee) {
		return callee
	}
	args := make([]Value, len(node.Args))
	for i, expr_node := range node.Args {
		expr := ctx.EvalExpr(expr_node)
		if isError(expr) {
			return expr
		}
		args[i] = expr
	}
	rv := ctx.call(nil, callee, NIL, args)
	if isError(rv) {
		return ctx.addErrorStack(rv.(*Error), node.LParen)
	}
	return rv
}

func (ctx *Context) evalIdentifier(node *parser.Identifier) Value {
	name := node.Id.Lexeme
	value, ok := ctx.env.ancestor(node.Loc).get(name)
	if !ok {
		e := newError(ctx, String(fmt.Sprintf("%q is not defined", name)))
		return ctx.addErrorStack(e, node.Id)
	}
	return value
}

func (ctx *Context) evalArray(node *parser.Array) Value {
	values := make([]Value, len(node.Exprs))
	for i, expr := range node.Exprs {
		val := ctx.EvalExpr(expr)
		if isError(val) {
			return val
		}
		values[i] = val
	}
	return newArray(ctx, values)
}

func (ctx *Context) evalHash(node *parser.Hash) Value {
	obj := newHash(ctx)
	hash := obj.data.(*Hash)
	for _, pair := range node.Pairs {
		k := ctx.EvalExpr(pair.Key)
		if isError(k) {
			return k
		}
		v := ctx.EvalExpr(pair.Value)
		if isError(v) {
			return v
		}
		err := hash.table.insert(k, v)
		if err != nil {
			ctx.addErrorStack(err, node.LBrace)
			return err
		}
	}
	return obj
}

func (ctx *Context) evalFunction(node *parser.Function) Value {
	fn := ctx.stack[len(ctx.stack)-1].Filename()
	return newFunction(fn, node, ctx.env)
}

func (ctx *Context) evalSuper(node *parser.Super) Value {
	proto := ctx.getPrototype(ctx.whence)
	if proto == nil {
		e := newError(ctx, String("object has nil prototype"))
		return ctx.addErrorStack(e, node.Tok)
	}
	return Super{proto}
}

// =========
// Utilities
// =========

func (ctx *Context) addErrorStack(err *Error, token lexer.Token) *Error {
	cse := ctx.stack[len(ctx.stack)-1]
	err.stack = append(err.stack, context{
		fn:  cse.Filename(),
		ln:  token.Line,
		col: token.Column,
		ctx: cse.Context(),
	})
	return err
}

func (ctx *Context) addErrorStackBuiltin(err *Error) *Error {
	cse := ctx.stack[len(ctx.stack)-1]
	err.stack = append(err.stack, context{
		fn:  cse.Filename(),
		ln:  0, col: 0,
		ctx: cse.Context(),
	})
	return err
}

func isSuper(s Value) bool    { return s.Type() == VT_SUPER }
func isError(s Value) bool    { return s.Type() == VT_ERROR }
func isBreak(s Value) bool    { return s.Type() == VT_BREAK }
func isContinue(s Value) bool { return s.Type() == VT_CONTINUE }
func isReturn(s Value) bool   { return s.Type() == VT_RETURN }
func isTruthy(s Value) bool   { return s != FALSE && s != NIL }
