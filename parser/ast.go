package parser

type Node interface {
	String() string
	node()
}

type Expr interface {
	Node
	expr()
}

type Stmt interface {
	Node
	stmt()
}

// Resolvable implements the interface required by the resolver.
// The resolver will add distance information (integers) onto
// _resolvable_ nodes.
type Resolvable interface {
	AddLocation(loc int)
}

func (node *Identifier) AddLocation(loc int) { node.Loc = loc }
func (node *Assign) AddLocation(loc int)     { node.Loc = loc }
