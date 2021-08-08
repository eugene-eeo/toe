package eval

import (
	"bytes"
	"fmt"
	"strconv"
)

// this package implements the built-in functions.

// ==============
// Object.clone()
// ==============

// Object.clone() returns a new object with `this' as the prototype.
func _Object_clone(ctx *Context, this Value, args []Value) Value {
	return newObject(this)
}

// ===============
// Function.bind()
// ===============

// Function.bind(x) binds the function to x.
func _Function_bind(ctx *Context, this Value, args []Value) Value {
	if len(args) == 0 {
		return ctx.err(String(fmt.Sprintf("expected 1 argument, got=%d", len(args))))
	}
	newThis := args[0]
	switch this.(type) {
	case *Function:
		return this.(*Function).Bind(newThis)
	case *Builtin:
		return this.(*Builtin).Bind(newThis)
	default:
		return ctx.err(String(fmt.Sprintf(".bind() called on a non-function")))
	}
}

// =========
// .length()
// =========

// String.length() returns the length of the string.
func _String_length(ctx *Context, this Value, args []Value) Value {
	switch this.(type) {
	case String:
		return Number(len(this.(String)))
	default:
		return ctx.err(String(fmt.Sprintf(".length() called on a non-string")))
	}
}

// ==========
// .inspect()
// ==========

// Object.inspect() returns the `repr' value of the object.
func _Object_inspect(ctx *Context, this Value, args []Value) Value {
	if obj, ok := this.(*Object); ok && len(obj.props) > 0 {
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf("<object %p", this))
		buf.WriteString("(\n")
		for key, value := range obj.props {
			// to prevent .inspect()-ing forever
			value_inspect := "(...)"
			if value != this {
				value_inspect = ctx.Inspect(value)
			}
			buf.WriteString(fmt.Sprintf("  %s=%s,\n", key, value_inspect))
		}
		buf.WriteString(")>")
		return String(buf.String())
	}
	return String(fmt.Sprintf("<object %p>", this))
}

func _Nil_inspect(ctx *Context, this Value, args []Value) Value {
	switch this {
	case NIL:
		return String("nil")
	}
	return _Object_inspect(ctx, this, args)
}

func _Boolean_inspect(ctx *Context, this Value, args []Value) Value {
	switch this {
	case TRUE:
		return String("true")
	case FALSE:
		return String("false")
	}
	return _Object_inspect(ctx, this, args)
}

func _String_inspect(ctx *Context, this Value, args []Value) Value {
	if str, ok := this.(String); ok {
		return String(fmt.Sprintf("%q", str))
	}
	return _Object_inspect(ctx, this, args)
}

func _Number_inspect(ctx *Context, this Value, args []Value) Value {
	if n, ok := this.(Number); ok {
		return String(strconv.FormatFloat(float64(n), 'g', -1, 64))
	}
	return _Object_inspect(ctx, this, args)
}

func _Function_inspect(ctx *Context, this Value, args []Value) Value {
	if this.Type() == BUILTIN || this.Type() == FUNCTION {
		return String(fmt.Sprintf("<function %p>", this))
	}
	return _Object_inspect(ctx, this, args)
}

func (ctx *Context) Inspect(v Value) string {
	fn, ok := ctx.getAttr(v, "inspect")
	if !ok {
		return fmt.Sprintf("<go value %#v>", v)
	}
	fn = ctx.bind(fn, v)
	rv, ok := ctx.callFunction(fn, []Value{})
	if !ok {
		return fmt.Sprintf("<go value %#v>", v)
	}
	str, ok := rv.(String)
	if !ok {
		return fmt.Sprintf("<go value %#v>", v)
	}
	return string(str)
}
