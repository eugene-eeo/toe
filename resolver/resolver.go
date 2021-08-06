// Package resolver implements the identifer resolution semantic analysis.
package resolver

import (
	"fmt"
	"toe/lexer"
	"toe/parser"
)

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

const (
	LOOP = 1 << iota
	FUNC
)

type Resolver struct {
	// each scope is a map from varname to a boolean, corresponding
	// to whether the variable was already initialised.
	module   *parser.Module
	scopes   []Scope
	Errors   []ResolverError
	Locs     map[*parser.Identifier]int
	ctrl     uint8 // control block -- whether we're in a loop / function
}

func New(module *parser.Module) *Resolver {
	r := &Resolver{
		module: module,
		scopes: []Scope{},
		Errors: []ResolverError{},
		Locs:   map[*parser.Identifier]int{},
		ctrl:   0,
	}
	r.push() // the global scope.
	scope := r.curr()
	scope["module"] = true
	scope["Object"] = true
	scope["Boolean"] = true
	scope["Number"] = true
	scope["String"] = true
	scope["Array"] = true
	return r
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

// Cleanup is used to clean up the resolver.
func (r *Resolver) Cleanup() {
	r.scopes = nil
}

func (r *Resolver) Resolve() {
	for _, stmt := range r.module.Stmts {
		r.resolve(stmt)
	}
	if len(r.scopes) != 1 || r.ctrl != 0 {
		panic("something gone wrong!")
	}
}

func (r *Resolver) ResolveOne(node parser.Node) {
	r.resolve(node)
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
	case *parser.Identifier:
		r.resolveIdentifier(node)
	case *parser.Literal:
		// nothing to resolve.
		return
	}
}

// ==========
// Statements
// ==========

func (r *Resolver) resolveLet(node *parser.Let) {
	name := node.Name.Tok().Lexeme
	curr := r.curr()
	curr[name] = false
	r.resolve(node.Value)
	curr[name] = true
}

func (r *Resolver) resolveBlock(node *parser.Block) {
	r.push()
	for _, x := range node.Statements {
		r.resolve(x)
	}
	r.pop()
}

func (r *Resolver) resolveFor(node *parser.For) {
	name := node.Name.Tok().Lexeme
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
		r.err(node.Token, "break outside of loop")
	}
}

func (r *Resolver) resolveContinue(node *parser.Continue) {
	if r.ctrl&LOOP == 0 {
		r.err(node.Token, "continue outside of loop")
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
	r.resolve(node.Left)
	r.resolve(node.Right)
}

func (r *Resolver) resolveUnary(node *parser.Unary) {
	r.resolve(node.Right)
}

func (r *Resolver) resolveGet(node *parser.Get) {
	r.resolve(node.Left)
}

func (r *Resolver) resolveIdentifier(node *parser.Identifier) {
	name := node.Token.Lexeme
	curr := len(r.scopes) - 1
	// loop until we find a closest scope containing the name.
	for i := curr; i >= 0; i-- {
		initialised, ok := r.scopes[i][name]
		if ok {
			// if we're referring to an uninitialised variable, e.g.
			// let a = a, then we can return an error -- unless:
			//  1. we're in a function AND
			//  2. we didn't find the name in the current scope.
			if !initialised {
				if ((r.ctrl & FUNC) != 0) && i != curr {
					r.Locs[node] = i
					return
				}
				r.err(node.Token, fmt.Sprintf("cannot access %q before initialization", name))
				return
			}
			r.Locs[node] = i
			return
		}
	}
	if (r.ctrl & FUNC) != 0 {
		// if we're in a function, then we find variables
		// in the global scope -- this is to allow things
		// like:
		//
		//      let x = fn(b) {
		//         return a + b;  // <-- `a' is found outside.
		//      }
		//      let a = 1;
		//
		r.Locs[node] = curr
	} else {
		// otherwise, this is an error.
		r.err(node.Token, fmt.Sprintf("undefined variable %q", name))
	}
}