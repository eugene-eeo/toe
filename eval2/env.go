package eval2

type environment struct {
	store map[string]Value
	outer *environment
}

func newEnv(outer *environment) *environment {
	return &environment{
		store: map[string]Value{},
		outer: outer,
	}
}

func (e *environment) ancestor(d int) *environment {
	for d > 0 {
		d--
		e = e.outer
	}
	return e
}

func (e *environment) get(name string) (Value, bool) {
	v, ok := e.store[name]
	return v, ok
}

func (e *environment) set(name string, v Value) {
	e.store[name] = v
}
