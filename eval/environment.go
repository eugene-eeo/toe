package eval

type Environment struct {
	filename string // the filename of the module this environment encloses
	store    map[string]Value
	outer    *Environment
}

func newEnvironment(filename string, outer *Environment) *Environment {
	return &Environment{
		filename: filename,
		store: map[string]Value{},
		outer: outer,
	}
}

// Ancestor returns the environment that is distance x
// away from the current environment.
func (e *Environment) Ancestor(distance int) *Environment {
	for distance > 0 {
		distance--
		e = e.outer
	}
	return e
}

// GetAt gets the variable name at the environment that is distance x
// away from the current environment.
func (e *Environment) GetAt(distance int, name string) (Value, bool) {
	val, ok := e.Ancestor(distance).store[name]
	return val, ok
}

// Define binds the given name to the given value.
func (e *Environment) Define(name string, value Value) {
	e.store[name] = value
}
