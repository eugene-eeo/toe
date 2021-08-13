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

// --------
// get_slot
// --------
func bi_get_slot(ctx *Context, this Value, args []Value) Value {
	if err := expectNArgs(args, 2); err != nil {
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
	if err := expectNArgs(args, 3); err != nil {
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
	if err := expectNArgs(args, 1); err != nil {
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
	if err := expectNArgs(args, 1); err != nil {
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
	if err := expectNArgs(args, 2); err != nil {
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
		return rv
	}
	// do we have an init slot?
	var whence Value
	if init := ctx.maybeGetSlot(rv, "init", &whence); init != nil {
		err := ctx.call(whence, init, rv, args);
		if isError(err) {
			return err
		}
	}
	return rv
}

func bi_Object_inspect(ctx *Context, this Value, args []Value) Value {
	typ := ctx.maybeGetSlot(ctx.getPrototype(this), "type", nil)
	if typ == nil {
		typ = String("Object")
	}
	str := ctx.getSpecial(typ, VT_STRING)
	if str == nil {
		str = String("Object")
	}
	return String(fmt.Sprintf("[%s %p]", string(str.(String)), this))
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

func bi_Number_init(ctx *Context, this Value, args []Value) Value {
	if this.Type() == VT_OBJECT {
		if len(args) > 0 && args[0] != NIL {
			num, err := expectArgType(ctx, "s", args[0], VT_NUMBER)
			if err != nil {
				return err
			}
			obj := this.(*Object)
			obj.data = num.(Number)
		}
		return NIL
	} else {
		return newError(String("'Number.init' called on non-object"))
	}
}

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

func bi_String_init(ctx *Context, this Value, args []Value) Value {
	if this.Type() == VT_OBJECT {
		if len(args) > 0 && args[0] != NIL {
			str, err := expectArgType(ctx, "n", args[0], VT_STRING)
			if err != nil {
				return err
			}
			obj := this.(*Object)
			obj.data = str.(String)
		}
		return NIL
	} else {
		return newError(String("'Number.init' called on non-object"))
	}
}

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
		return newError(String("'Array.init' called on non-object"))
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

func bi_Array_concat(ctx *Context, this Value, args []Value) Value {
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

func bi_Array_get(ctx *Context, this Value, args []Value) Value {
	arr := ctx.getSpecial(this, VT_ARRAY)
	if arr == nil {
		return newError(String("no VT_ARRAY in prototype chain"))
	}
	if err := expectNArgs(args, 1); err != nil {
		return err
	}
	num := ctx.getSpecial(args[0], VT_NUMBER)
	if num == nil {
		return newError(String("index: no VT_NUMBER in prototype chain"))
	}
	me := arr.(*Array)
	idx := int(num.(Number))
	if 0 <= idx && idx < len(me.values) {
		return me.values[idx]
	}
	return newError(String("list index out of bounds"))
}

func bi_Array_set(ctx *Context, this Value, args []Value) Value {
	arr, err := expectArgType(ctx, "this", this, VT_HASH)
	if err != nil {
		return err
	}
	if err := expectNArgs(args, 2); err != nil {
		return err
	}
	num := ctx.getSpecial(args[0], VT_NUMBER)
	if num == nil {
		return newError(String("index: no VT_NUMBER in prototype chain"))
	}
	me := arr.(*Array)
	idx := int(num.(Number))
	if 0 <= idx && idx < len(me.values) {
		me.values[idx] = args[1]
		return args[1]
	}
	return newError(String("list index out of bounds"))
}

func bi_Array_push(ctx *Context, this Value, args []Value) Value {
	arr, err := expectArgType(ctx, "this", this, VT_HASH)
	if err != nil {
		return err
	}
	me := arr.(*Array)
	for _, x := range args {
		me.values = append(me.values, x)
	}
	return NIL
}

func bi_Array_pop(ctx *Context, this Value, args []Value) Value {
	arr, err := expectArgType(ctx, "this", this, VT_HASH)
	if err != nil {
		return err
	}
	me := arr.(*Array)
	sz := len(me.values)
	if sz == 0 {
		return newError(String("pop from empty string"))
	}
	idx := sz - 1
	if len(args) > 0 {
		num, err := expectArgType(ctx, "index", args[0], VT_NUMBER)
		if err != nil {
			return err
		}
		idx = int(num.(Number))
		if !(0 <= idx && idx < sz) {
			return newError(String("list index out of bounds"))
		}
	}
	me.values = append(me.values[idx:], me.values[idx+1:]...)
	return NIL
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
		return newError(String("'Hash.init' called on non-object"))
	}
}

func bi_Hash_get(ctx *Context, this Value, args []Value) Value {
	hash, err := expectArgType(ctx, "this", this, VT_HASH)
	if err != nil {
		return err
	}
	if err := expectNArgs(args, 1); err != nil {
		return err
	}
	rv, found, err := hash.(*Hash).table.get(args[0])
	if err != nil {
		return err
	}
	if !found {
		return newError(String("key not in hash"))
	}
	return rv
}

func bi_Hash_set(ctx *Context, this Value, args []Value) Value {
	hash, err := expectArgType(ctx, "this", this, VT_HASH)
	if err != nil {
		return err
	}
	if err := expectNArgs(args, 2); err != nil {
		return err
	}
	if err := hash.(*Hash).table.insert(args[0], args[1]); err != nil {
		return err
	}
	return NIL
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
	g.Object.slots["type"] = String("Object")
	g.Object.slots["clone"] = newBuiltin("clone", bi_Object_clone)
	g.Object.slots["new"] = newBuiltin("clone", bi_Object_new)
	g.Object.slots["inspect"] = newBuiltin("inspect", bi_Object_inspect)
	g.Object.slots["=="] = newBuiltin("==", bi_Object_eq)
	g.Object.slots["!="] = newBuiltin("!=", bi_Object_neq)

	g.Function = newObject(g.Object)
	g.Function.slots["type"] = String("Function")
	g.Function.slots["bind"] = newBuiltin("bind", bi_Function_bind)

	g.Error = newObject(g.Object)
	g.Error.slots["type"] = String("Error")
	g.Error.slots["throw"] = newBuiltin("throw", bi_Error_throw)

	g.Boolean = newObject(g.Object)

	g.Number = newObject(g.Object)
	g.Number.slots["type"] = String("Number")
	g.Number.slots["init"] = newBuiltin("init", bi_Number_init)
	g.Number.slots["=="] = binOp2Builtin("==", bi_Number_equal, VT_NUMBER, VT_NUMBER)
	g.Number.slots["+"] = binOp2Builtin("+", bi_Number_plus, VT_NUMBER, VT_NUMBER)
	g.Number.slots["-"] = binOp2Builtin("-", bi_Number_minus, VT_NUMBER, VT_NUMBER)
	g.Number.slots["*"] = binOp2Builtin("*", bi_Number_times, VT_NUMBER, VT_NUMBER)
	g.Number.slots["/"] = binOp2Builtin("/", bi_Number_divide, VT_NUMBER, VT_NUMBER)
	g.Number.slots[">"] = binOp2Builtin(">", bi_Number_gt, VT_NUMBER, VT_NUMBER)
	g.Number.slots[">="] = binOp2Builtin(">=", bi_Number_geq, VT_NUMBER, VT_NUMBER)
	g.Number.slots["<"] = binOp2Builtin("<", bi_Number_lt, VT_NUMBER, VT_NUMBER)
	g.Number.slots["<="] = binOp2Builtin("<=", bi_Number_leq, VT_NUMBER, VT_NUMBER)

	g.String = newObject(g.Object)
	g.String.slots["type"] = String("String")
	g.String.slots["init"] = newBuiltin("init", bi_String_init)
	g.String.slots["=="] = binOp2Builtin("==", bi_String_equal, VT_STRING, VT_STRING)
	g.String.slots["+"] = binOp2Builtin("+", bi_String_plus, VT_STRING, VT_STRING)
	g.String.slots[">"] = binOp2Builtin(">", bi_String_gt, VT_STRING, VT_STRING)
	g.String.slots[">="] = binOp2Builtin(">=", bi_String_geq, VT_STRING, VT_STRING)
	g.String.slots["<"] = binOp2Builtin("<", bi_String_lt, VT_STRING, VT_STRING)
	g.String.slots["<="] = binOp2Builtin("<=", bi_String_leq, VT_STRING, VT_STRING)

	g.Array = newObject(g.Object)
	g.Array.slots["type"] = String("Array")
	g.Array.slots["init"] = newBuiltin("init", bi_Array_init)
	g.Array.slots["=="] = binOp2Builtin("==", bi_Array_equal, VT_ARRAY, VT_ARRAY)
	g.Array.slots["+"] = binOp2Builtin("+", bi_Array_plus, VT_ARRAY, VT_ARRAY)
	g.Array.slots["concat"] = newBuiltin("concat", bi_Array_concat)
	g.Array.slots["get"] = newBuiltin("get", bi_Array_get)
	g.Array.slots["set"] = newBuiltin("set", bi_Array_set)
	g.Array.slots["pop"] = newBuiltin("pop", bi_Array_pop)

	g.Hash = newObject(g.Object)
	g.Hash.slots["type"] = String("Hash")
	g.Hash.slots["init"] = newBuiltin("init", bi_Hash_init)
	g.Hash.slots["get"] = newBuiltin("get", bi_Hash_get)
	g.Hash.slots["set"] = newBuiltin("set", bi_Hash_set)
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

func expectNArgs(args []Value, n int) Value {
	if len(args) == n {
		return nil
	}
	return newError(String(fmt.Sprintf("expected %d argument(s), got=%d", n, len(args))))
}

func expectArgType(ctx *Context, argName string, value Value, expectedType ValueType) (rv Value, err Value) {
	rv = ctx.getSpecial(value, expectedType)
	if rv == nil {
		return nil, newError(String(fmt.Sprintf(
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
