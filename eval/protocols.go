package eval

import (
	"fmt"
	"toe/lexer"
	"unicode/utf8"
)

// =========
// Iterators
// =========
//
// For loops are equivalent to:
//
//                  | let it = obj.iter()
// for (x : obj)    | while (!it.done()) {
//     ...          |   let x = it.next()
//                  |   ...
//                  | }
//                  | it.close()

type Iterator interface {
	Close() Value
	Done() Value
	Next() Value
}

type StringIterator struct {
	i int
	s String
}

func (si *StringIterator) Close() Value { return NIL }
func (si *StringIterator) Done() Value  { return Boolean(si.i >= len(si.s)) }
func (si *StringIterator) Next() Value {
	r, w := utf8.DecodeRuneInString(string(si.s)[si.i:])
	si.i += w
	return String(r)
}

type ArrayIterator struct {
	i int
	a *Array
}

func (ai *ArrayIterator) Close() Value { return NIL }
func (ai *ArrayIterator) Done() Value  { return Boolean(ai.i >= len(ai.a.values)) }
func (ai *ArrayIterator) Next() Value {
	rv := ai.a.values[ai.i]
	ai.i++
	return rv
}

type HashIterator struct {
	curr  int
	valid uint64
	hash  *Hash
}

func (hi *HashIterator) Close() Value { return NIL }
func (hi *HashIterator) Done() Value  { return Boolean(hi.valid == hi.hash.table.size()) }
func (hi *HashIterator) Next() Value {
	ht := hi.hash.table
	for i := hi.curr; i < len(ht.entries); i++ {
		entry := &ht.entries[i]
		if entry.hasValue() {
			hi.valid++
			return *entry.key
		}
	}
	return NIL
}

func getIterator(v Value) (Iterator, bool) {
	switch v := v.(type) {
	case String:
		return &StringIterator{s: v}, true
	case *Array:
		return &ArrayIterator{a: v}, true
	case *Hash:
		return &HashIterator{hash: v}, true
	}
	return nil, false
}

// ============
// Object Model
// ============
//
// Toe is prototype-based; thus there are no classes, just objects (bag-of-values).
// Below we implement the functions needed to modify these objects. Note that some
// of them may return error values; you have to check for them!
//
//  1. Whenever we see a nil prototype, we can stop searching.
//  2. All built-in `types' have prototype == Object, except for nil.
//  3. Object's prototype == nil.

// -------
// Binding
// -------

func (ctx *Context) bind(obj Value, this Value) Value {
	switch obj := obj.(type) {
	case *Function:
		return obj.Bind(this)
	case *Builtin:
		return obj.Bind(this)
	}
	return obj
}

func (f *Function) Bind(this Value) *Function {
	// A bound function cannot be bound again. By checking if this == nil,
	// we allow for explicitly binding to NIL.
	if f.this != nil {
		return f
	}
	// Lightweight wrapper around f that binds it to `this'
	return &Function{
		Object: f.Object,
		node: f.node,
		this: this,
		closure: f.closure,
		filename: f.filename,
	}
}

func (b *Builtin) Bind(this Value) *Builtin {
	if b.this != nil {
		return b
	}
	return &Builtin{Object: b.Object, this: this, call: b.call}
}

// -------------
// Get Prototype
// -------------

func (ctx *Context) getPrototype(obj Value) Value {
	switch obj := obj.(type) {
	case *Object:
		return obj.proto
	case *Function:
		return obj.proto
	}
	return nil
}

// --------
// Get Slot
// --------

func (ctx *Context) getSlot(obj Value, name string) Value {
	for obj != nil {
		if obj_ho, ok := obj.(hasObject); ok {
			if v, ok := obj_ho.object().slots[name]; ok {
				return v
			}
		}
		obj = ctx.getPrototype(obj)
	}
	err := newError(String(fmt.Sprintf("object has no slot %q", name)))
	return err
}

type hasObject interface{ object() *Object }

func (o *Object) object() *Object { return o }

// --------
// Set Slot
// --------

func (ctx *Context) setSlot(obj Value, name string, val Value) Value {
	if obj_ho, ok := obj.(hasObject); ok {
		oo := obj_ho.object()
		oo.slots[name] = val
		return val
	}
	err := newError(String(fmt.Sprintf("cannot set slot %q on object", name)))
	return err
}

// ==============
// Function Calls
// ==============

func (ctx *Context) call(callee Value, this Value, args []Value) Value {
	switch callee := callee.(type) {
	case *Function:
		return callee.Call(ctx, this, args)
	case *Builtin:
		return callee.Call(ctx, this, args)
	}
	err := newError(String("not a function"))
	return err
}

func (f *Function) Call(ctx *Context, this Value, args []Value) Value {
	old_env := ctx.env
	old_this := ctx.this
	if f.this != nil {
		this = f.this
	}

	ctx.env = f.closure
	ctx.this = this
	ctx.pushEnv()
	ctx.pushFunc(&functionCse{f})

	ctx.env.set("this", this)
	for i, id := range f.node.Params {
		name := id.Lexeme
		if len(args) <= i {
			ctx.env.set(name, NIL)
		} else {
			ctx.env.set(name, args[i])
		}
	}

	// Remember to unwrap return values.
	rv := ctx.evalBlock(f.node.Body)
	if isReturn(rv) {
		rv = rv.(Return).value
	}

	ctx.popFunc()
	ctx.popEnv()
	ctx.this = old_this
	ctx.env = old_env
	return rv
}

func (b *Builtin) Call(ctx *Context, this Value, args []Value) Value {
	if b.this != nil {
		this = b.this
	}
	ctx.pushFunc(&builtinCse{b})
	rv := b.call(ctx, this, args)
	ctx.popFunc()
	return rv
}

// =========
// Operators
// =========

func (ctx *Context) binary(op lexer.TokenType, left, right Value) Value {
	// Fast case for ==, != cannot be fast-cased...
	if op == lexer.EQUAL_EQUAL && left == right {
		return TRUE
	}
	// Search the operator table.
	rv, ok := ctx.binaryDispatch(op, left, right)
	if !ok {
		// Try to see if we can find a logical alternative.
		// For example, a != b === !(a == b). Note that here we _cannot_
		// use binary(...) again.
		switch op {
		// == and != falls back to pointer equality
		case lexer.EQUAL_EQUAL:
			altRv, altOk := ctx.binaryDispatch(lexer.BANG_EQUAL, left, right)
			if altOk {
				if isError(altRv) {
					return altRv
				} else {
					return Boolean(!isTruthy(altRv))
				}
			}
			return Boolean(left == right)
		case lexer.BANG_EQUAL:
			altRv, altOk := ctx.binaryDispatch(lexer.EQUAL_EQUAL, left, right)
			if altOk {
				if isError(altRv) {
					return altRv
				} else {
					return Boolean(!isTruthy(altRv))
				}
			}
			return Boolean(left != right)
		}
		// There really is no implementation.
		return newError(String(fmt.Sprintf(
			"unsupported operands for %q: %s and %s",
			op.String(),
			left.Type().String(),
			right.Type().String(),
		)))
	}
	return rv
}

// binaryDispatch searches the binary operations table for a corresponding
// handler for the given values.
func (ctx *Context) binaryDispatch(op lexer.TokenType, left, right Value) (Value, bool) {
	info := binOpInfo{op, left.Type(), right.Type()}
	if impl, ok := binOpTable[info]; ok {
		rv := impl(ctx, left, right)
		return rv, true
	}
	return nil, false
}

// areObjectsEqual returns TRUE if the two objects are equal, FALSE otherwise.
func (ctx *Context) areObjectsEqual(left, right Value) Value {
	return ctx.binary(lexer.EQUAL_EQUAL, left, right)
}

func (ctx *Context) unary(op lexer.TokenType, right Value) Value {
	switch {
	case op == lexer.BANG:
		return Boolean(!isTruthy(right))
	case op == lexer.MINUS && right.Type() == VT_NUMBER:
		return Number(-right.(Number))
	}
	err := newError(String(fmt.Sprintf(
		"unsupported operand for %q: %s",
		op.String(),
		right.Type().String(),
	)))
	return err
}

// =========
// Utilities
// =========

// bind_and_call is a utility method to call $obj.$name, binding
// it at the same time -- this can be used as the base for future
// protocols (e.g. object iter).
func (ctx *Context) bind_and_call(obj Value, name string, args []Value) Value {
	fn := ctx.getSlot(obj, name)
	if isError(fn) {
		return fn
	}
	return ctx.call(fn, obj, args)
}
