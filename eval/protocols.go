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

func (f *Function) Bind(this Value) *Function {
	// A bound function cannot be bound again. By checking if this == nil,
	// we allow for explicitly binding to NIL.
	if f.this != nil {
		return f
	}
	g := newFunction(f.filename, f.node, f.closure)
	g.this = this
	return g
}

func (b *Builtin) Bind(this Value) *Builtin {
	if b.this != nil {
		return b
	}
	return &Builtin{slots: newSlots(), this: this, call: b.call}
}

// -------------
// Get Prototype
// -------------

func (ctx *Context) getPrototype(obj Value) Value {
	switch obj := obj.(type) {
	case Nil:
		return nil
	case Boolean:
		return ctx.globals.Boolean
	case Number:
		return ctx.globals.Number
	case String:
		return ctx.globals.String
	case *Array:
		return ctx.globals.Array
	case *Hash:
		return ctx.globals.Hash
	case *Object:
		return obj.proto
	case *Function:
		return ctx.globals.Function
	case *Builtin:
		return ctx.globals.Function
	}
	return nil
}

// getSpecial is used to search the prototype chain for an object
// matching the given internal Value type. This has the effect of
// treating builtin objects as some special field. This is useful
// for e.g. if users want to subtype builtin objects. If the return
// value is nil, then the search is unsuccessful.
func (ctx *Context) getSpecial(obj Value, typ ValueType) Value {
	for obj != nil {
		objTyp := obj.Type()
		if objTyp == typ {
			return obj
		}
		if objTyp == VT_OBJECT {
			data := obj.(*Object).data
			if data != nil && data.Type() == typ {
				return data
			}
		}
		obj = ctx.getPrototype(obj)
	}
	return nil
}

// --------
// Get Slot
// --------

// maybeGetSlot fetches the slot $obj.$name, traversing the prototype chain, and
// returning nil if not found.
func (ctx *Context) maybeGetSlot(obj Value, name string, whence *Value) Value {
	for obj != nil {
		if obj_slots, ok := obj.(hasSlots); ok {
			if v, ok := obj_slots.getSlots()[name]; ok {
				if whence != nil {
					*whence = obj
				}
				return v
			}
		}
		obj = ctx.getPrototype(obj)
	}
	return nil
}

// getSlot uses maybeGetSlot internally, but returns an error if the slot is
// not found.
func (ctx *Context) getSlot(obj Value, name string, whence *Value) Value {
	rv := ctx.maybeGetSlot(obj, name, whence)
	if rv == nil {
		err := newError(String(fmt.Sprintf("object has no slot %q", name)))
		return err
	}
	return rv
}

type hasSlots interface{ getSlots() map[string]Value }

func (o *Object) getSlots() map[string]Value   { return o.slots }
func (f *Function) getSlots() map[string]Value { return f.slots }
func (b *Builtin) getSlots() map[string]Value  { return b.slots }

// --------
// Set Slot
// --------

func (ctx *Context) setSlot(obj Value, name string, val Value) Value {
	if obj_slots, ok := obj.(hasSlots); ok {
		slots := obj_slots.getSlots()
		slots[name] = val
		return val
	}
	err := newError(String(fmt.Sprintf("cannot set slot %q on object", name)))
	return err
}

// ==============
// Function Calls
// ==============

func (ctx *Context) call(whence Value, callee Value, this Value, args []Value) (rv Value) {
	old_whence := ctx.whence
	ctx.whence = whence
	switch callee := callee.(type) {
	case *Function:
		rv = callee.Call(ctx, this, args)
	case *Builtin:
		rv = callee.Call(ctx, this, args)
	default:
		rv = newError(String("not a function"))
	}
	ctx.whence = old_whence
	return
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
	old_this := ctx.this
	if b.this != nil {
		this = b.this
	}
	ctx.pushFunc(&builtinCse{b})
	ctx.this = this
	rv := b.call(ctx, this, args)
	ctx.this = old_this
	ctx.popFunc()
	return rv
}

// =========
// Operators
// =========

func (ctx *Context) binary(op string, left, right Value) Value {
	// Fast case for ==
	if op == "==" && left == right {
		return TRUE
	}
	return ctx.call_method(left, op, []Value{right})
}

// areObjectsEqual is a shortcut for binary(==, ...)
func (ctx *Context) areObjectsEqual(left, right Value) Value {
	return ctx.binary("==", left, right)
}

func (ctx *Context) unary(op lexer.TokenType, right Value) Value {
	switch {
	case op == lexer.BANG:
		return Boolean(!isTruthy(right))
	case op == lexer.MINUS && right.Type() == VT_NUMBER:
		return Number(-right.(Number))
	}
	return newError(String(fmt.Sprintf(
		"unsupported operand for %q: %s",
		op.String(),
		right.Type().String(),
	)))
}

// =========
// Utilities
// =========

// call_method is a utility method to call $obj.$name, binding
// it at the same time -- this can be used as the base for future
// protocols (e.g. object iter).
func (ctx *Context) call_method(obj Value, name string, args []Value) Value {
	var whence Value
	fn := ctx.getSlot(obj, name, &whence)
	if isError(fn) {
		return fn
	}
	return ctx.call(whence, fn, obj, args)
}

// forward forwards the call `name` up the prototype chain.
func (ctx *Context) forward(obj Value, name string, args []Value) Value {
	var whence Value
	fn := ctx.getSlot(ctx.getPrototype(ctx.whence), name, &whence)
	if isError(fn) {
		return fn
	}
	return ctx.call(whence, fn, obj, args)
}
