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

// Resolve finds the environment where `name' is found.
// If no such env is found, then the return value is nil.
func (e *Environment) Resolve(name string) *Environment {
	if _, ok := e.store[name]; ok {
		return e
	}
	if e.outer != nil {
		return e.outer.Resolve(name)
	}
	return nil
}

// Define binds the given name to the given value.
func (e *Environment) Define(name string, value Value) {
	e.store[name] = value
}

// Get gets the given name from the environment, traversing
// the outer environments if it is not found.
func (e *Environment) Get(name string) (Value, bool) {
	v, ok := e.store[name]
	if ok {
		return v, true
	}
	if e.outer != nil {
		return e.outer.Get(name)
	}
	return nil, false
}
