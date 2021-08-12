package eval

import (
	"bytes"
	"fmt"
	"strconv"
)

// This file implements the inspect() protocol.
// Most of the complication comes from ensuring that recursive objects,
// e.g. a list containing itself does not crash the process.

type valueInspector func(v Value) string

type inspectable interface {
	inspect(valueInspector) string
}

func inspect(v Value) string {
	seen := map[inspectable]bool{}
	var visit valueInspector
	visit = func(v Value) string {
		switch v := v.(type) {
		case Stringer:
			return v.String()
		case inspectable:
			if seen[v] {
				return "(...)"
			}
			seen[v] = true
			return v.inspect(visit)
		}
		panic(fmt.Sprintf("cannot inspect: %#+v", v))
	}
	return visit(v)
}

// Container types (or String -- which needs special formatting).

func (v String) inspect(f valueInspector) string {
	return fmt.Sprintf("%q", string(v))
}

func (v *Array) inspect(f valueInspector) string {
	var buf bytes.Buffer
	last_idx := len(v.values) - 1
	buf.WriteString("[")
	for i, x := range v.values {
		buf.WriteString(f(x))
		if i != last_idx {
			buf.WriteString(", ")
		}
	}
	buf.WriteString("]")
	return buf.String()
}

func (v *Hash) inspect(f valueInspector) string {
	var buf bytes.Buffer
	buf.WriteString("{")
	j := uint64(0)
	size := v.table.size()
	for i := 0; i < len(v.table.entries); i++ {
		ref := &v.table.entries[i]
		if ref.hasValue() {
			j++
			buf.WriteString(f(*ref.key))
			buf.WriteString(": ")
			buf.WriteString(f(*ref.value))
			if j < size {
				buf.WriteString(", ")
			}
		}
	}
	buf.WriteString("}")
	return buf.String()
}

// =========
// Stringify
// =========

type Stringer interface {
	String() string
}

func (v Nil) String() string { return "nil" }
func (v Boolean) String() string {
	if v {
		return "true"
	}
	return "false"
}
func (v Number) String() string { return strconv.FormatFloat(float64(v), 'g', -1, 64) }
func (v String) String() string { return string(v) }

func (v *Function) String() string {
	name := v.node.Name
	if name != "" {
		name = " " + name
	}
	isBound := ""
	if v.this != nil {
		isBound = " bound"
	}
	return fmt.Sprintf("[Function%s%s %p]", isBound, name, v)
}

func (v *Object) String() string {
	return fmt.Sprintf("[Object %p]", v)
}

func (v *Builtin) String() string {
	isBound := ""
	if v.this != nil {
		isBound = " bound"
	}
	return fmt.Sprintf("[Function%s %s]", isBound, v.name)
}

func (v *Array) String() string { return inspect(v) }
func (v *Hash)  String() string { return inspect(v) }
