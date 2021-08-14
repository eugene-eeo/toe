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
		// fmt.Println(obj.(Stringer).String())
		fmt.Println(obj)
	}
	return NIL
}

// --------
// get_slot
// --------
func bi_get_slot(ctx *Context, this Value, args []Value) Value {
	if err := expectNArgs(ctx, args, 2); err != nil {
		return err
	}
	slot, err := expectArgType(ctx, "slot", args[1], VT_STRING)
	if err != nil {
		return err
	}
	return ctx.getSlot(args[0], string(slot.(String)), nil)
}

// --------
// set_slot
// --------
func bi_set_slot(ctx *Context, this Value, args []Value) Value {
	if err := expectNArgs(ctx, args, 3); err != nil {
		return err
	}
	slot, err := expectArgType(ctx, "slot", args[1], VT_STRING)
	if err != nil {
		return err
	}
	return ctx.setSlot(args[0], string(slot.(String)), args[2])
}

// ----------
// slot_names
// ----------
func bi_slot_names(ctx *Context, this Value, args []Value) Value {
	if err := expectNArgs(ctx, args, 1); err != nil {
		return err
	}
	slots := []Value{}
	if obj_slots, ok := args[0].(hasSlots); ok {
		for slot := range obj_slots.getSlots() {
			slots = append(slots, String(slot))
		}
	}
	return newArray(ctx, slots)
}

// ---------
// get_proto
// ---------
func bi_get_proto(ctx *Context, this Value, args []Value) Value {
	if err := expectNArgs(ctx, args, 1); err != nil {
		return err
	}
	rv := ctx.getPrototype(args[0])
	if rv == nil {
		return NIL
	}
	return rv
}

// ----
// is_a
// ----
func bi_is_a(ctx *Context, this Value, args []Value) Value {
	// Does the given object appear anywhere on my prototype chain?
	if err := expectNArgs(ctx, args, 2); err != nil {
		return err
	}
	obj := args[0]
	query := args[1]
	for obj != nil {
		if obj == query {
			return TRUE
		}
		obj = ctx.getPrototype(obj)
	}
	return FALSE
}

// ------
// Object
// ------

func bi_Object_clone(ctx *Context, this Value, args []Value) Value {
	return newObject(this)
}

func bi_Object_new(ctx *Context, this Value, args []Value) Value {
	rv := ctx.call_method(this, "clone", []Value{})
	if isError(rv) {
		ctx.addErrorStackBuiltin(rv.(*Error))
		return rv
	}
	// do we have an init slot?
	var whence Value
	if init := ctx.maybeGetSlot(rv, "init", &whence); init != nil {
		err := ctx.call(whence, init, rv, args);
		if isError(err) {
			ctx.addErrorStackBuiltin(err.(*Error))
			return err
		}
	}
	return rv
}

// ----------------
// Object Operators
// ----------------

func bi_Object_eq(ctx *Context, this Value, args []Value) Value {
	if err := expectNArgs(ctx, args, 1); err != nil {
		return err
	}
	return Boolean(this == args[0])
}

func bi_Object_neq(ctx *Context, this Value, args []Value) Value {
	rv := ctx.call_method(this, "==", args)
	if isError(rv) {
		ctx.addErrorStackBuiltin(rv.(*Error))
		return rv
	}
	return Boolean(!isTruthy(rv))
}

// --------
// Function
// --------

func bi_Function_bind(ctx *Context, this Value, args []Value) Value {
	if err := expectNArgs(ctx, args, 1); err != nil {
		return err
	}
	target := args[0]
	switch this := this.(type) {
	case *Function:
		return this.Bind(target)
	case *Builtin:
		return this.Bind(target)
	}
	return newError(ctx, String(fmt.Sprintf("invalid receiver type %s", this.Type())))
}

func bi_Function_call(ctx *Context, this Value, args []Value) Value {
	var call_this Value = NIL
	if len(args) > 0 {
		call_this = args[0]
		args = args[1:]
	}
	return ctx.call(NIL, this, call_this, args)
}

// -----
// Error
// -----

func bi_Error_throw(ctx *Context, this Value, args []Value) Value {
	return newError(ctx, this)
}

// ------
// Number
// ------

func bi_Number_equal(ctx *Context, left, right Value) Value  { return Boolean(left.(Number) == right.(Number)) }
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

func bi_String_equal(ctx *Context, left, right Value) Value { return Boolean(left.(String) == right.(String)) }
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

func bi_Array_init(ctx *Context, this Value, args []Value) Value {
	if this.Type() == VT_OBJECT {
		obj := this.(*Object)
		obj.data = &Array{args}
		return NIL
	} else {
		return newError(ctx, String("'Array.init' called on non-object"))
	}
}

func bi_Array_equal(ctx *Context, a, b Value) Value {
	left := a.(*Array)
	right := b.(*Array)
	if len(left.values) != len(right.values) {
		return FALSE
	}
	for i, lhs := range left.values {
		rhs := right.values[i]
		rv := ctx.areObjectsEqual(lhs, rhs)
		if isError(rv) {
			ctx.addErrorStackBuiltin(rv.(*Error))
			return rv
		}
		if !isTruthy(rv) {
			return FALSE
		}
	}
	return TRUE
}

func bi_Array_plus(ctx *Context, a, b Value) Value {
	left := a.(*Array)
	right := b.(*Array)
	l_sz := len(left.values)
	r_sz := len(right.values)
	values := make([]Value, l_sz+r_sz)
	copy(values, left.values)
	copy(values[l_sz:], right.values)
	return newArray(ctx, values)
}

var bi_Array_concat = make_method(
	make_argspec(VT_ARRAY, make_argpair("arr", VT_ARRAY)),
	func(ctx *Context, this Value, args[]Value) Value {
		arr := this.(*Array)
		arr.values = append(arr.values, args[0].(*Array).values...)
		return NIL
	},
)

var bi_Array_size = make_method(
	make_argspec(VT_ARRAY),
	func(ctx *Context, this Value, args []Value) Value {
		return Number(len(this.(*Array).values))
	},
)

var bi_Array_get = make_method(
	make_argspec(VT_ARRAY, make_argpair("size", VT_NUMBER)),
	func(ctx *Context, this Value, args []Value) Value {
		arr := this.(*Array)
		idx := int(args[0].(Number))
		if 0 <= idx && idx < len(arr.values) {
			return arr.values[idx]
		}
		return newError(ctx, String("list index out of bounds"))
	},
)

var bi_Array_set = make_method(
	make_argspec(VT_ARRAY, make_argpair("index", VT_NUMBER), make_argpair("value", VT_ANY)),
	func (ctx *Context, this Value, args []Value) Value {
		arr := this.(*Array)
		idx := int(args[0].(Number))
		if 0 <= idx && idx < len(arr.values) {
			arr.values[idx] = args[1]
			return args[1]
		}
		return newError(ctx, String("list index out of bounds"))
	},
)

var bi_Array_push = make_method(
	make_argspec(VT_ARRAY, make_argpair("value", VT_ARRAY)),
	func (ctx *Context, this Value, args []Value) Value {
		arr := this.(*Array)
		arr.values = append(arr.values, args[0])
		return nil
	},
)

func bi_Array_pop(ctx *Context, this Value, args []Value) Value {
	arr, err := expectArgType(ctx, "this", this, VT_ARRAY)
	if err != nil {
		return err
	}
	me := arr.(*Array)
	sz := len(me.values)
	if sz == 0 {
		return newError(ctx, String("pop from empty string"))
	}
	idx := sz - 1
	if len(args) > 0 {
		num, err := expectArgType(ctx, "index", args[0], VT_NUMBER)
		if err != nil {
			return err
		}
		idx = int(num.(Number))
		if !(0 <= idx && idx < sz) {
			return newError(ctx, String("list index out of bounds"))
		}
	}
	rv := me.values[idx]
	me.values = append(me.values[idx:], me.values[idx+1:]...)
	return rv
}

// ----
// Hash
// ----

func bi_Hash_init(ctx *Context, this Value, args []Value) Value {
	if this.Type() == VT_OBJECT {
		obj := this.(*Object)
		obj.data = &Hash{table: newHashTable(ctx)}
		return NIL
	} else {
		return newError(ctx, String("'Hash.init' called on non-object"))
	}
}

func bi_Hash_get(ctx *Context, this Value, args []Value) Value {
	hash, err := expectArgType(ctx, "this", this, VT_HASH)
	if err != nil {
		return err
	}
	if err := expectNArgs(ctx, args, 1); err != nil {
		return err
	}
	rv, found, err := hash.(*Hash).table.get(args[0])
	if err != nil {
		return err
	}
	if !found {
		return newError(ctx, String("key not in hash"))
	}
	return rv
}

func bi_Hash_set(ctx *Context, this Value, args []Value) Value {
	hash, err := expectArgType(ctx, "this", this, VT_HASH)
	if err != nil {
		return err
	}
	if err := expectNArgs(ctx, args, 2); err != nil {
		return err
	}
	if err := hash.(*Hash).table.insert(args[0], args[1]); err != nil {
		return err
	}
	return NIL
}

func bi_Hash_delete(ctx *Context, this Value, args []Value) Value {
	hash, err := expectArgType(ctx, "this", this, VT_HASH)
	if err != nil {
		return err
	}
	if err := expectNArgs(ctx, args, 1); err != nil {
		return err
	}
	found, err := hash.(*Hash).table.delete(args[0])
	if err != nil {
		return err
	}
	if !found {
		return newError(ctx, String("key not in hash"))
	}
	return NIL
}

func bi_Hash_size(ctx *Context, this Value, args []Value) Value {
	hash, err := expectArgType(ctx, "this", this, VT_HASH)
	if err != nil {
		return err
	}
	return Number(hash.(*Hash).table.size())
}

func bi_Hash_equal(ctx *Context, a, b Value) Value {
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
				ctx.addErrorStackBuiltin(rv.(*Error))
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
	puts       *Builtin
	set_slot   *Builtin
	get_slot   *Builtin
	slot_names *Builtin
	get_proto  *Builtin
	is_a       *Builtin
	Object     *Object
	Function   *Object
	Error      *Object
	Boolean    *Object
	Number     *Object
	String     *Object
	Array      *Object
	Hash       *Object
}

func newGlobals() *Globals {
	g := &Globals{}
	g.puts = newBuiltin("puts", bi_puts)
	g.set_slot = newBuiltin("set_slot", bi_set_slot)
	g.get_slot = newBuiltin("set_slot", bi_get_slot)
	g.slot_names = newBuiltin("slot_names", bi_slot_names)
	g.get_proto = newBuiltin("get_proto", bi_get_proto)
	g.is_a = newBuiltin("is_a", bi_is_a)

	g.Object = newObject(nil)
	g.Object.slots["clone"] = newBuiltin("clone", bi_Object_clone)
	g.Object.slots["new"] = newBuiltin("clone", bi_Object_new)
	g.Object.slots["inspect"] = newBuiltin("inspect", bi_Object_inspect)
	g.Object.slots["=="] = newBuiltin("==", bi_Object_eq)
	g.Object.slots["!="] = newBuiltin("!=", bi_Object_neq)

	g.Function = newObject(g.Object)
	g.Function.slots["bind"] = newBuiltin("bind", bi_Function_bind)
	g.Function.slots["call"] = newBuiltin("call", bi_Function_call)
	g.Function.slots["inspect"] = newBuiltin("inspect", bi_Function_inspect)

	g.Error = newObject(g.Object)
	g.Error.slots["throw"] = newBuiltin("throw", bi_Error_throw)

	g.Boolean = newObject(g.Object)
	g.Boolean.slots["init"] = newBuiltin("init", builtin_init(VT_BOOLEAN, FALSE))
	g.Boolean.slots["inspect"] = newBuiltin("inspect", bi_Boolean_inspect)

	g.Number = newObject(g.Object)
	g.Number.slots["init"] = newBuiltin("init", builtin_init(VT_NUMBER, Number(0)))
	g.Number.slots["=="] = binOp2Builtin("==", bi_Number_equal, VT_NUMBER, VT_NUMBER)
	g.Number.slots["+"] = binOp2Builtin("+", bi_Number_plus, VT_NUMBER, VT_NUMBER)
	g.Number.slots["-"] = binOp2Builtin("-", bi_Number_minus, VT_NUMBER, VT_NUMBER)
	g.Number.slots["*"] = binOp2Builtin("*", bi_Number_times, VT_NUMBER, VT_NUMBER)
	g.Number.slots["/"] = binOp2Builtin("/", bi_Number_divide, VT_NUMBER, VT_NUMBER)
	g.Number.slots[">"] = binOp2Builtin(">", bi_Number_gt, VT_NUMBER, VT_NUMBER)
	g.Number.slots[">="] = binOp2Builtin(">=", bi_Number_geq, VT_NUMBER, VT_NUMBER)
	g.Number.slots["<"] = binOp2Builtin("<", bi_Number_lt, VT_NUMBER, VT_NUMBER)
	g.Number.slots["<="] = binOp2Builtin("<=", bi_Number_leq, VT_NUMBER, VT_NUMBER)
	g.Number.slots["inspect"] = newBuiltin("inspect", bi_Number_inspect)

	g.String = newObject(g.Object)
	g.String.slots["init"] = newBuiltin("init", builtin_init(VT_STRING, String("")))
	g.String.slots["=="] = binOp2Builtin("==", bi_String_equal, VT_STRING, VT_STRING)
	g.String.slots["+"] = binOp2Builtin("+", bi_String_plus, VT_STRING, VT_STRING)
	g.String.slots[">"] = binOp2Builtin(">", bi_String_gt, VT_STRING, VT_STRING)
	g.String.slots[">="] = binOp2Builtin(">=", bi_String_geq, VT_STRING, VT_STRING)
	g.String.slots["<"] = binOp2Builtin("<", bi_String_lt, VT_STRING, VT_STRING)
	g.String.slots["<="] = binOp2Builtin("<=", bi_String_leq, VT_STRING, VT_STRING)
	g.String.slots["inspect"] = newBuiltin("inspect", bi_String_inspect)

	g.Array = newObject(g.Object)
	g.Array.slots["init"] = newBuiltin("init", bi_Array_init)
	g.Array.slots["=="] = binOp2Builtin("==", bi_Array_equal, VT_ARRAY, VT_ARRAY)
	g.Array.slots["+"] = binOp2Builtin("+", bi_Array_plus, VT_ARRAY, VT_ARRAY)
	g.Array.slots["concat"] = newBuiltin("concat", bi_Array_concat)
	g.Array.slots["size"] = newBuiltin("size", bi_Array_size)
	g.Array.slots["get"] = newBuiltin("get", bi_Array_get)
	g.Array.slots["set"] = newBuiltin("set", bi_Array_set)
	g.Array.slots["push"] = newBuiltin("push", bi_Array_push)
	g.Array.slots["pop"] = newBuiltin("pop", bi_Array_pop)
	g.Array.slots["inspect_visit"] = newBuiltin("inspect_visit", bi_Array_inspect_visit)

	g.Hash = newObject(g.Object)
	g.Hash.slots["init"] = newBuiltin("init", bi_Hash_init)
	g.Hash.slots["size"] = newBuiltin("size", bi_Hash_size)
	g.Hash.slots["get"] = newBuiltin("get", bi_Hash_get)
	g.Hash.slots["set"] = newBuiltin("set", bi_Hash_set)
	g.Hash.slots["delete"] = newBuiltin("delete", bi_Hash_delete)
	g.Hash.slots["=="] = binOp2Builtin("==", bi_Hash_equal, VT_HASH, VT_HASH)

	return g
}

func (g *Globals) addToEnv(env *environment) {
	env.set("puts", g.puts)
	env.set("set_slot", g.set_slot)
	env.set("get_slot", g.get_slot)
	env.set("slot_names", g.slot_names)
	env.set("get_proto", g.get_proto)
	env.set("is_a", g.is_a)
	env.set("Object", g.Object)
	env.set("Function", g.Function)
	env.set("Error", g.Error)
	env.set("Boolean", g.Boolean)
	env.set("Number", g.Number)
	env.set("String", g.String)
	env.set("Array", g.Array)
	env.set("Hash", g.Hash)
}

func (g *Globals) addToResolver(r *resolver.Resolver) {
	r.AddGlobals([]string{
		"puts",
		"set_slot", "get_slot", "slot_names", "get_proto", "is_a",
		"Object", "Function", "Error", "Number", "String", "Array", "Hash",
	})
}

// ========
// Utilties
// ========

func expectNArgs(ctx *Context, args []Value, n int) *Error {
	if len(args) == n {
		return nil
	}
	return newError(ctx, String(fmt.Sprintf("expected %d argument(s), got=%d", n, len(args))))
}

func expectArgType(ctx *Context, argName string, value Value, expectedType ValueType) (rv Value, err *Error) {
	if expectedType == VT_ANY {
		return value, nil
	}
	if expectedType == VT_CALL {
		rv = ctx.getSpecial(value, VT_FUNCTION)
		if rv != nil {
			return rv, nil
		}
		rv = ctx.getSpecial(value, VT_BUILTIN)
		if rv != nil {
			return rv, nil
		}
		return nil, newError(ctx, String(fmt.Sprintf(
			"argument '%s' is not callable",
			argName,
		)))
	}
	rv = ctx.getSpecial(value, expectedType)
	if rv == nil {
		return nil, newError(ctx, String(fmt.Sprintf(
			"argument '%s' has no %s in prototype chain",
			argName,
			expectedType,
		)))
	}
	return rv, nil
}

type binOpFunc func(*Context, Value, Value) Value

func binOp2Builtin(name string, f binOpFunc, ltype, rtype ValueType) *Builtin {
	return newBuiltin(name, func(ctx *Context, this Value, args []Value) Value {
		if err := expectNArgs(ctx, args, 1); err != nil {
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

// builtin_init generates an init method for immutable builtins, i.e. String,
// Number, Boolean.
func builtin_init(typ ValueType, zero Value) builtinFunc {
	return func(ctx *Context, this Value, args []Value) Value {
		fmt.Println(this)
		if this.Type() == VT_OBJECT {
			obj := this.(*Object)
			if obj.data != nil {
				return newError(ctx, String("object already initialised"))
			}
			obj.data = zero
			// Check for NIL here: this makes init methods easier to write.
			if len(args) > 0 && args[0] != NIL {
				value, err := expectArgType(ctx, "x", args[0], typ)
				if err != nil {
					return err
				}
				obj.data = value
			}
			return NIL
		} else {
			return newError(ctx, String("init called on non-object"))
		}
	}
}

type argPair struct {
	name string
	typ ValueType
}

type argSpec struct {
	this ValueType
	args []argPair
}

func make_argpair(name string, typ ValueType) argPair { return argPair{name, typ} }
func make_argspec(this ValueType, args ...argPair) argSpec {
	return argSpec{this, args}
}

func make_method(spec argSpec, fn builtinFunc) builtinFunc {
	return func(ctx *Context, this Value, args []Value) Value {
		// check the argspec.
		newThis, err := expectArgType(ctx, "this", this, spec.this)
		if err != nil {
			ctx.addErrorStackBuiltin(err)
			return err
		}
		if err := expectNArgs(ctx, args, len(spec.args)); err != nil {
			ctx.addErrorStackBuiltin(err)
			return err
		}
		newArgs := make([]Value, len(args))
		for i, arg := range args {
			value, err := expectArgType(ctx, spec.args[i].name, arg, spec.args[i].typ)
			if err != nil {
				ctx.addErrorStackBuiltin(err)
				return err
			}
			newArgs[i] = value
		}
		return fn(ctx, newThis, newArgs)
	}
}
