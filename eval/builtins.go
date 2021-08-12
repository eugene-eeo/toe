package eval

import (
	"fmt"
	"toe/resolver"
)

// =================
// Builtin functions
// =================

// ----
// puts
// ----

func bi_puts(ctx *Context, this Value, args []Value) Value {
	for _, obj := range args {
		fmt.Println(obj.(Stringer).String())
	}
	return NIL
}

// ------
// Object
// ------

func bi_Object_proto(ctx *Context, this Value, args []Value) Value {
	rv := ctx.getPrototype(this)
	if rv == nil {
		return NIL
	}
	return rv
}

func bi_Object_clone(ctx *Context, this Value, args []Value) Value {
	var rv Value
	switch this {
	// If we're cloning some builtin, then remember to return the
	// default values for each builtin.
	case ctx.globals.Boolean:
		return FALSE
	case ctx.globals.Number:
		return Number(0)
	case ctx.globals.String:
		return String("")
	case ctx.globals.Array:
		rv = newArray([]Value{})
	case ctx.globals.Hash:
		rv = newHash(ctx)
	default:
		rv = newObject(this)
	}
	// do we have an init slot?
	var whence Value
	slot := ctx.maybeGetSlot(rv, "init", &whence)
	if slot != nil {
		// yes: forward all arguments to the init method.
		res := ctx.call(whence, slot, rv, args)
		if isError(res) {
			return res
		}
	}
	return rv
}

func bi_Object_is_a(ctx *Context, this Value, args []Value) Value {
	// Does the given object appear anywhere on my prototype chain?
	if err := expectNArgs(args, 1); err != nil {
		return err
	}
	query := args[0]
	for this != nil {
		if this == query {
			return TRUE
		}
		this = ctx.getPrototype(this)
	}
	return FALSE
}

// ----------------
// Object Operators
// ----------------

func bi_Object_eq(ctx *Context, this Value, args []Value) Value {
	if err := expectNArgs(args, 1); err != nil {
		return err
	}
	return Boolean(this == args[0])
}

func bi_Object_neq(ctx *Context, this Value, args []Value) Value {
	rv := ctx.call_method(this, "==", args)
	if isError(rv) {
		return rv
	}
	return Boolean(!isTruthy(rv))
}

// --------
// Function
// --------

func bi_Function_bind(ctx *Context, this Value, args []Value) Value {
	if err := expectNArgs(args, 1); err != nil {
		return err
	}
	target := args[0]
	switch this := this.(type) {
	case *Function:
		return this.Bind(target)
	case *Builtin:
		return this.Bind(target)
	}
	return newError(String(fmt.Sprintf("invalid receiver type %s", this.Type())))
}

// -----
// Error
// -----

func bi_Error_throw(ctx *Context, this Value, args []Value) Value {
	return newError(this)
}

// ------
// Number
// ------

func bi_Number_plus(ctx *Context, left, right Value) Value   { return left.(Number) + right.(Number) }
func bi_Number_minus(ctx *Context, left, right Value) Value  { return left.(Number) - right.(Number) }
func bi_Number_times(ctx *Context, left, right Value) Value  { return left.(Number) * right.(Number) }
func bi_Number_divide(ctx *Context, left, right Value) Value { return left.(Number) / right.(Number) }
func bi_Number_gt(ctx *Context, left, right Value) Value {
	return Boolean(left.(Number) > right.(Number))
}
func bi_Number_geq(ctx *Context, left, right Value) Value {
	return Boolean(left.(Number) >= right.(Number))
}
func bi_Number_lt(ctx *Context, left, right Value) Value {
	return Boolean(left.(Number) < right.(Number))
}
func bi_Number_leq(ctx *Context, left, right Value) Value {
	return Boolean(left.(Number) <= right.(Number))
}

// ------
// String
// ------

func bi_String_plus(ctx *Context, left, right Value) Value { return left.(String) + right.(String) }
func bi_String_gt(ctx *Context, left, right Value) Value {
	return Boolean(left.(String) > right.(String))
}
func bi_String_geq(ctx *Context, left, right Value) Value {
	return Boolean(left.(String) >= right.(String))
}
func bi_String_lt(ctx *Context, left, right Value) Value {
	return Boolean(left.(String) < right.(String))
}
func bi_String_leq(ctx *Context, left, right Value) Value {
	return Boolean(left.(String) <= right.(String))
}

// -----
// Array
// -----

func bi_array_equal(ctx *Context, a, b Value) Value {
	left := a.(*Array)
	right := b.(*Array)
	if len(left.values) != len(right.values) {
		return FALSE
	}
	for i, lhs := range left.values {
		rhs := right.values[i]
		rv := ctx.areObjectsEqual(lhs, rhs)
		if isError(rv) {
			return rv
		}
		if !isTruthy(rv) {
			return FALSE
		}
	}
	return TRUE
}

func bi_array_concat_new(ctx *Context, a, b Value) Value {
	left := a.(*Array)
	right := b.(*Array)
	l_sz := len(left.values)
	r_sz := len(right.values)
	values := make([]Value, l_sz+r_sz)
	copy(values, left.values)
	copy(values[l_sz:], right.values)
	return newArray(values)
}

func bi_array_concat(ctx *Context, this Value, args []Value) Value {
	arr := ctx.getSpecial(this, VT_ARRAY)
	if arr == nil {
		return newError(String("no VT_ARRAY in prototype chain"))
	}
	me := arr.(*Array)
	for i, x := range args {
		ext := ctx.getSpecial(x, VT_ARRAY)
		if ext == nil {
			return newError(String(fmt.Sprintf("args[%d]: no VT_ARRAY in prototype chain", i)))
		}
		me.values = append(me.values, ext.(*Array).values...)
	}
	return NIL
}

// ----
// Hash
// ----

func bi_hash_equal(ctx *Context, a, b Value) Value {
	left := a.(*Hash)
	right := b.(*Hash)
	if left.table.size() != right.table.size() {
		return FALSE
	}
	for _, entry := range left.table.entries {
		if entry.hasValue() {
			key := *entry.key
			lhs := *entry.value
			rhs, found, err := right.table.get(key)
			if err != nil {
				return err
			}
			if !found {
				return FALSE
			}
			rv := ctx.areObjectsEqual(lhs, rhs)
			if isError(rv) {
				return rv
			}
			if !isTruthy(rv) {
				return FALSE
			}
		}
	}
	return TRUE
}

// =======
// Globals
// =======

type Globals struct {
	Object   *Object
	Function *Object
	Error    *Object
	Boolean  *Object
	Number   *Object
	String   *Object
	Array    *Object
	Hash     *Object
}

func newGlobals() *Globals {
	g := &Globals{}

	g.Object = newObject(nil)
	g.Object.slots["proto"] = newBuiltin("proto", bi_Object_proto)
	g.Object.slots["clone"] = newBuiltin("clone", bi_Object_clone)
	g.Object.slots["is_a"] = newBuiltin("is_a", bi_Object_is_a)
	g.Object.slots["=="] = newBuiltin("==", bi_Object_eq)
	g.Object.slots["!="] = newBuiltin("!=", bi_Object_neq)

	g.Function = newObject(g.Object)
	g.Function.slots["bind"] = newBuiltin("bind", bi_Function_bind)

	g.Error = newObject(g.Object)
	g.Error.slots["throw"] = newBuiltin("throw", bi_Error_throw)

	g.Boolean = newObject(g.Object)

	g.Number = newObject(g.Object)
	g.Number.slots["+"] = binOp2Builtin("+", bi_Number_plus, VT_NUMBER, VT_NUMBER)
	g.Number.slots["-"] = binOp2Builtin("-", bi_Number_minus, VT_NUMBER, VT_NUMBER)
	g.Number.slots["*"] = binOp2Builtin("*", bi_Number_times, VT_NUMBER, VT_NUMBER)
	g.Number.slots["/"] = binOp2Builtin("/", bi_Number_divide, VT_NUMBER, VT_NUMBER)
	g.Number.slots[">"] = binOp2Builtin(">", bi_Number_gt, VT_NUMBER, VT_NUMBER)
	g.Number.slots[">="] = binOp2Builtin(">=", bi_Number_geq, VT_NUMBER, VT_NUMBER)
	g.Number.slots["<"] = binOp2Builtin("<", bi_Number_lt, VT_NUMBER, VT_NUMBER)
	g.Number.slots["<="] = binOp2Builtin("<=", bi_Number_leq, VT_NUMBER, VT_NUMBER)

	g.String = newObject(g.Object)
	g.String.slots["+"] = binOp2Builtin("+", bi_String_plus, VT_STRING, VT_STRING)
	g.String.slots[">"] = binOp2Builtin(">", bi_String_gt, VT_STRING, VT_STRING)
	g.String.slots[">="] = binOp2Builtin(">=", bi_String_geq, VT_STRING, VT_STRING)
	g.String.slots["<"] = binOp2Builtin("<", bi_String_lt, VT_STRING, VT_STRING)
	g.String.slots["<="] = binOp2Builtin("<=", bi_String_leq, VT_STRING, VT_STRING)

	g.Array = newObject(g.Object)
	g.Array.slots["=="] = binOp2Builtin("==", bi_array_equal, VT_ARRAY, VT_ARRAY)
	g.Array.slots["+"] = binOp2Builtin("+", bi_array_concat_new, VT_ARRAY, VT_ARRAY)
	g.Array.slots["concat"] = newBuiltin("concat", bi_array_concat)

	g.Hash = newObject(g.Object)
	g.Hash.slots["=="] = binOp2Builtin("==", bi_hash_equal, VT_HASH, VT_HASH)

	return g
}

func (g *Globals) addToEnv(env *environment) {
	env.set("Object", g.Object)
	env.set("Function", g.Function)
	env.set("Error", g.Error)
	env.set("Number", g.Number)
	env.set("String", g.String)
	env.set("Array", g.Array)
	env.set("Hash", g.Hash)
}

func (g *Globals) addToResolver(r *resolver.Resolver) {
	r.AddGlobals([]string{"Object", "Function", "Error", "Number", "String", "Array", "Hash"})
}

// ========
// Utilties
// ========

func expectNArgs(args []Value, n int) Value {
	if len(args) == n {
		return nil
	}
	return newError(String(fmt.Sprintf("expected %d argument(s), got=%d", n, len(args))))
}

type binOpFunc func(*Context, Value, Value) Value

func binOp2Builtin(name string, f binOpFunc, ltype, rtype ValueType) *Builtin {
	return newBuiltin(name, func(ctx *Context, this Value, args []Value) Value {
		if err := expectNArgs(args, 1); err != nil {
			return err
		}
		left := ctx.getSpecial(this, ltype)
		if left == nil {
			return ctx.forward(this, name, args)
		}
		right := ctx.getSpecial(args[0], rtype)
		if right == nil {
			return ctx.forward(this, name, args)
		}
		return f(ctx, left, right)
	})
}
