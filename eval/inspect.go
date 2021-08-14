package eval

import (
	"bytes"
	"fmt"
	"strconv"
)

// This file implements the inspect() protocol.
//
// inspect() calls inspect_visit(f) with a function used to call the
// next inspected object.

var bi_Object_inspect = make_method(
	make_argspec(VT_ANY),
	func(ctx *Context, this Value, args []Value) Value {
		outer_this := this
		m := map[Value]bool{}
		var visitor *Builtin
		visitor = newBuiltin("visitor", make_method(
			make_argspec(VT_ANY, make_argpair("value", VT_ANY)),
			func (ctx *Context, _ Value, args []Value) Value {
				v := args[0]
				if m[v] {
					return String("...")
				} else {
					m[v] = true
					// Check if we have an inspect_visit method; if so then we have
					// to call it; otherwise just call the normal inspect().
					var rv Value
					if ctx.maybeGetSlot(v, "inspect_visit", nil) != nil {
						rv = ctx.call_method(v, "inspect_visit", []Value{visitor})
					} else if v != outer_this {
						rv = ctx.call_method(v, "inspect", nil)
					} else {
						rv = String(fmt.Sprintf("[Object %p]", v))
					}
					if isError(rv) {
						ctx.addErrorStackBuiltin(rv.(*Error))
						return rv
					}
					str := ctx.getSpecial(rv, VT_STRING)
					if str == nil {
						err := newError(ctx, String("inspect should return a string"))
						return ctx.addErrorStackBuiltin(err)
					}
					return str
				}
			},
		))
		return ctx.call(NIL, visitor, NIL, []Value{outer_this})
	},
)

var bi_Function_inspect = make_method(
	make_argspec(VT_FUNCTION),
	func(ctx *Context, this Value, args []Value) Value {
		return String(fmt.Sprintf("[Function %p]", this))
	},
)

var bi_Boolean_inspect = make_method(
	make_argspec(VT_BOOLEAN),
	func(ctx *Context, this Value, args []Value) Value {
		if this == TRUE {
			return String("true")
		}
		return String("false")
	},
)

var bi_String_inspect = make_method(
	make_argspec(VT_STRING),
	func(ctx *Context, this Value, args []Value) Value {
		return String(fmt.Sprintf("%q", string(this.(String))))
	},
)

var bi_Number_inspect = make_method(
	make_argspec(VT_NUMBER),
	func(ctx *Context, this Value, args []Value) Value {
		return String(strconv.FormatFloat(
			float64(this.(Number)),
			'g', -1, 64,
		))
	},
)

var bi_Array_inspect_visit = make_method(
	make_argspec(VT_ARRAY, make_argpair("f", VT_CALL)),
	func(ctx *Context, this Value, args []Value) Value {
		f := args[0].(*Builtin)
		var buf bytes.Buffer
		buf.WriteString("[")
		arr := this.(*Array)
		sz := len(arr.values)
		for i, x := range arr.values {
			s := ctx.call(NIL, f, NIL, []Value{x})
			if isError(s) {
				ctx.addErrorStackBuiltin(s.(*Error))
				return s
			}
			buf.WriteString(string(s.(String)))
			if i != sz - 1 {
				buf.WriteString(", ")
			}
		}
		buf.WriteString("]")
		return String(buf.String())
	},
)
