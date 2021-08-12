// Package resolver implements identifer resolution semantic analysis,
// as well as some syntax checks (e.g. ensuring that continues and breaks
// are within a loop construct). Identifier resolution works by recording
// the distance from the current environment where an identifier can be
// found.
package resolver

import (
	"errors"
	"fmt"
	"toe/lexer"
	"toe/parser"
)

var TooManyErrors = errors.New("too many errors")

type ResolverError struct {
	Filename string
	Token    lexer.Token
	Message  string
}

func (re ResolverError) Error() string { return re.String() }
func (re ResolverError) String() string {
	return fmt.Sprintf("%s:%d:%d: %s", re.Filename, re.Token.Line, re.Token.Column, re.Message)
}

type Scope map[string]bool

// Control flags -- whether we are in a loop, or a function.
const (
	LOOP = 1 << iota
	FUNC
)

type Resolver struct {
	module *parser.Module
	// each scope is a map from varname to a boolean, corresponding
	// to whether the variable was already initialised.
	scopes []Scope
	Errors []error
	ctrl   uint8
}

func New(module *parser.Module) *Resolver {
	r := &Resolver{
		module: module,
		scopes: []Scope{},
		Errors: []error{},
		ctrl:   0,
	}
	r.push() // the global scope.
	return r
}

func (r *Resolver) AddGlobals(globals []string) {
	for _, x := range globals {
		r.scopes[0][x] = true
	}
}

func (r *Resolver) curr() Scope { return r.scopes[len(r.scopes)-1] }
func (r *Resolver) push()       { r.scopes = append(r.scopes, Scope{}) }
func (r *Resolver) pop()        { r.scopes = r.scopes[:len(r.scopes)-1] }

func (r *Resolver) err(tok lexer.Token, msg string) {
	r.Errors = append(r.Errors, ResolverError{
		Filename: r.module.Filename,
		Token:    tok,
		Message:  msg,
	})
}

// Clean up frees memory used by the resolver -- this can only be done
// after reporting errors, as it clears the errors as well.
func (r *Resolver) Cleanup() {
	r.scopes = nil
	r.Errors = nil
}

// ResolveOne resolves the given node -- it is mainly for
// interactive usage.
func (r *Resolver) ResolveOne(node parser.Node) {
	r.resolve(node)
}

// Resolve resolves the given module.
// This method can only be called once.
func (r *Resolver) Resolve() {
	for _, stmt := range r.module.Stmts {
		r.resolve(stmt)
		if len(r.Errors) >= 10 {
			r.Errors = append(r.Errors, TooManyErrors)
			break
		}
	}
	if len(r.scopes) != 1 || r.ctrl != 0 {
		panic("something gone wrong!")
	}
}

func (r *Resolver) resolve(node parser.Node) {
	switch node := node.(type) {
	// Statements
	case *parser.Let:
		r.resolveLet(node)
	case *parser.Block:
		r.resolveBlock(node)
	case *parser.For:
		r.resolveFor(node)
	case *parser.While:
		r.resolveWhile(node)
	case *parser.If:
		r.resolveIf(node)
	case *parser.ExprStmt:
		r.resolveExprStmt(node)
	case *parser.Break:
		r.resolveBreak(node)
	case *parser.Continue:
		r.resolveContinue(node)
	case *parser.Return:
		r.resolveReturn(node)
	// Expressions
	case *parser.Binary:
		r.resolveBinary(node)
	case *parser.And:
		r.resolveAnd(node)
	case *parser.Or:
		r.resolveOr(node)
	case *parser.Assign:
		r.resolveAssign(node)
	case *parser.Unary:
		r.resolveUnary(node)
	case *parser.Get:
		r.resolveGet(node)
	case *parser.Set:
		r.resolveSet(node)
	case *parser.Method:
		r.resolveMethod(node)
	case *parser.Call:
		r.resolveCall(node)
	case *parser.Identifier:
		r.resolveIdentifier(node)
	case *parser.Literal:
		return
	case *parser.Array:
		r.resolveArray(node)
	case *parser.Hash:
		r.resolveHash(node)
	case *parser.Function:
		r.resolveFunction(node)
	case *parser.Super:
		r.resolveSuper(node)
	default:
		panic(fmt.Sprintf("unhandled node: %#+v", node))
	}
}

// ==========
// Statements
// ==========

func (r *Resolver) resolveLet(node *parser.Let) {
	name := node.Name.Lexeme
	curr := r.curr()
	if _, ok := curr[name]; ok {
		// is there already an existing let?
		r.err(node.Name, "already a variable with this name in scope.")
	}
	curr[name] = false
	r.resolve(node.Value)
	curr[name] = true
	addFunctionName(node.Value, name)
}

func (r *Resolver) resolveBlock(node *parser.Block) {
	r.push()
	for _, x := range node.Stmts {
		r.resolve(x)
	}
	r.pop()
}

func (r *Resolver) resolveFor(node *parser.For) {
	name := node.Name.Lexeme
	r.resolve(node.Iter)
	r.push()
	ctrl := r.ctrl
	r.ctrl |= LOOP
	r.curr()[name] = true
	r.resolve(node.Stmt)
	r.ctrl = ctrl
	r.pop()
}

func (r *Resolver) resolveWhile(node *parser.While) {
	r.resolve(node.Cond)
	ctrl := r.ctrl
	r.ctrl |= LOOP
	r.resolve(node.Stmt)
	r.ctrl = ctrl
}

func (r *Resolver) resolveIf(node *parser.If) {
	r.resolve(node.Cond)
	r.resolve(node.Then)
	if node.Else != nil {
		r.resolve(node.Else)
	}
}

func (r *Resolver) resolveExprStmt(node *parser.ExprStmt) {
	r.resolve(node.Expr)
}

func (r *Resolver) resolveBreak(node *parser.Break) {
	if r.ctrl&LOOP == 0 {
		r.err(node.Keyword, "break outside of loop")
	}
}

func (r *Resolver) resolveContinue(node *parser.Continue) {
	if r.ctrl&LOOP == 0 {
		r.err(node.Keyword, "continue outside of loop")
	}
}

func (r *Resolver) resolveReturn(node *parser.Return) {
	if r.ctrl&FUNC == 0 {
		r.err(node.Keyword, "return outside of function")
	}
	if node.Expr != nil {
		r.resolve(node.Expr)
	}
}

// ===========
// Expressions
// ===========

func (r *Resolver) resolveBinary(node *parser.Binary) {
	r.resolve(node.Left)
	r.resolve(node.Right)
}

func (r *Resolver) resolveAnd(node *parser.And) {
	r.resolve(node.Left)
	r.resolve(node.Right)
}

func (r *Resolver) resolveOr(node *parser.Or) {
	r.resolve(node.Left)
	r.resolve(node.Right)
}

func (r *Resolver) resolveAssign(node *parser.Assign) {
	r.resolve(node.Right)
	addFunctionName(node.Right, node.Name.Lexeme)
	r.lookup(node, node.Name)
}

func (r *Resolver) resolveUnary(node *parser.Unary) {
	r.resolve(node.Right)
}

func (r *Resolver) resolveGet(node *parser.Get) {
	r.resolve(node.Object)
}

func (r *Resolver) resolveSet(node *parser.Set) {
	r.resolve(node.Right)
	addFunctionName(node.Right, node.Name.Lexeme)
	r.resolve(node.Object)
}

func (r *Resolver) resolveMethod(node *parser.Method) {
	r.resolve(node.Object)
	for _, arg := range node.Args {
		r.resolve(arg)
	}
}

func (r *Resolver) resolveCall(node *parser.Call) {
	r.resolve(node.Callee)
	for _, arg := range node.Args {
		r.resolve(arg)
	}
}

func (r *Resolver) resolveIdentifier(node *parser.Identifier) {
	r.lookup(node, node.Id)
}

func (r *Resolver) resolveArray(node *parser.Array) {
	for _, expr := range node.Exprs {
		r.resolve(expr)
	}
}

func (r *Resolver) resolveHash(node *parser.Hash) {
	for _, pair := range node.Pairs {
		r.resolve(pair.Key)
		r.resolve(pair.Value)
	}
}

func (r *Resolver) resolveFunction(node *parser.Function) {
	// Function expressions -- we first push a new scope containing all
	// of the parameters, and then we resolve the body.
	ctrl := r.ctrl
	r.ctrl |= FUNC
	r.push()
	scope := r.curr()
	scope["this"] = true
	for _, name := range node.Params {
		scope[name.Lexeme] = true
	}
	r.resolveBlock(node.Body)
	r.pop()
	r.ctrl = ctrl
}

func (r *Resolver) resolveSuper(node *parser.Super) {
	if r.ctrl&FUNC == 0 {
		r.err(node.Tok, "super outside of function")
	}
}

func (r *Resolver) lookup(node parser.Expr, token lexer.Token) {
	name := token.Lexeme
	curr := len(r.scopes) - 1
	// loop until we find a closest scope containing the name.
	for i := curr; i >= 0; i-- {
		initialised, ok := r.scopes[i][name]
		if ok {
			// if we're referring to an uninitialised variable, e.g.
			// let a = a, then we can return an error -- unless:
			//  1. we're in a function AND
			//  2. we didn't find the name in the current scope.
			// this is to allow functions to refer to themselves.
			if !initialised && !((r.ctrl&FUNC) != 0 || i != curr) {
				r.err(token, fmt.Sprintf("cannot access %q before initialization", name))
				return
			}
			addLocation(node, curr-i)
			return
		}
	}
	if (r.ctrl & FUNC) != 0 {
		// if we're in a function, then we find variables in the global scope.
		// this allows things like:
		//
		//      let x = fn(b) {
		//         return a + b;  // <-- `a' is found outside.
		//      }
		//      let a = 1;
		//      x(2);
		//
		addLocation(node, curr)
	} else {
		r.err(token, fmt.Sprintf("undefined variable %q", name))
	}
}

// =========
// Utilities
// =========

func addLocation(node parser.Expr, loc int) {
	node.(parser.Resolvable).AddLocation(loc)
}

func addFunctionName(node parser.Expr, name string) {
	if fn, ok := node.(*parser.Function); ok {
		fn.Name = name
	}
}
