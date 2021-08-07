package eval

type Environment struct {
	store map[string]Value
	outer *Environment
}

func newEnvironment(outer *Environment) *Environment {
	return &Environment{
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

// // Get gets the given name from the environment, traversing
// // the outer environments if it is not found.
// func (e *Environment) Get(name string) (Value, bool) {
// 	v, ok := e.store[name]
// 	if ok {
// 		return v, true
// 	}
// 	if e.outer != nil {
// 		return e.outer.Get(name)
// 	}
// 	return nil, false
// }
