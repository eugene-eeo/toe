package eval

// callStackEntry contains partial information about the function call;
// only including the filename and the string.
type callStackEntry interface {
	Filename() string
	Context() string
}

type moduleCse struct{ filename string }

func (m moduleCse) Filename() string { return m.filename }
func (m moduleCse) Context() string  { return "[Module]" }

type functionCse struct {
	function *Function
}

func (f functionCse) Filename() string { return f.function.filename }
func (f functionCse) Context() string  { return f.function.String() }

type builtinCse struct {
	builtin *Builtin
}

func (b builtinCse) Filename() string { return "[builtin]" }
func (b builtinCse) Context() string  { return b.builtin.String() }
