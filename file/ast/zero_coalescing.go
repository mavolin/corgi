package ast

import "slices"

type ZeroCoalescing struct {
	DerefPosition *Position
	DerefCount    int // number of leading '*' pointer derefs
	Root          *Expression
	CheckRoot     *Position
	Chain         []ZeroCoalescingNode // chain behind root

	Tilde   *Position   // only set if we have a default
	Default *Expression // optional, not a ZeroCoalescing
}

var _ CodeNode = (*ZeroCoalescing)(nil)

func (c *ZeroCoalescing) Start() Position {
	if c.DerefPosition != nil {
		return *c.DerefPosition
	} else if c.Root != nil {
		return c.Root.Start()
	} else if c.CheckRoot != nil {
		return *c.CheckRoot
	}
	for _, node := range c.Chain {
		if node != nil {
			return node.Start()
		}
	}
	if c.Tilde != nil {
		return deltaPos(*c.Tilde, len("~"))
	} else if c.Default != nil {
		return c.Default.Start()
	}
	return Position{}
}
func (c *ZeroCoalescing) End() Position {
	if c.Default != nil {
		return c.Default.End()
	} else if c.Tilde != nil {
		return deltaPos(*c.Tilde, len("~"))
	}
	for _, node := range slices.Backward(c.Chain) {
		if node != nil {
			return node.End()
		}
	}
	if c.CheckRoot != nil {
		return deltaPos(*c.CheckRoot, len("?"))
	} else if c.Root != nil {
		return c.Root.End()
	} else if c.DerefPosition != nil {
		return deltaPos(*c.DerefPosition, c.DerefCount)
	}
	return Position{}
}

func (*ZeroCoalescing) _node()     {}
func (*ZeroCoalescing) _codeNode() {}

// ============================================================================
// Zero Coalescing Node
// ======================================================================================

// ZeroCoalescingNode is a node in a chain expression.
//
// It is either a [ZCIndexExpression], [ZCParenExpression],
// [ZCTypeAssertionExpression], or a [ZCSelectorExpression].
type ZeroCoalescingNode interface {
	Node
	_zeroCoalescingNode()
}

// if this is changed, change the comment above
var (
	_ ZeroCoalescingNode = (*ZCIndexExpression)(nil)
	_ ZeroCoalescingNode = (*ZCParenExpression)(nil)
	_ ZeroCoalescingNode = (*ZCTypeAssertionExpression)(nil)
	_ ZeroCoalescingNode = (*ZCSelectorExpression)(nil)
)

// ============================================================================
// Index Expression
// ======================================================================================

// ZCIndexExpression is either a map or slice index expression.
type ZCIndexExpression struct {
	LBracket   *Position
	Index      *Expression // not a ZeroCoalescing
	CheckIndex *Position
	Comma      *Position // optional
	RBracket   *Position
	CheckValue *Position
}

var _ ZeroCoalescingNode = (*ZCIndexExpression)(nil)

func (e *ZCIndexExpression) Start() Position {
	if e.LBracket != nil {
		return *e.LBracket
	} else if e.Index != nil {
		return e.Index.Start()
	} else if e.CheckIndex != nil {
		return *e.CheckIndex
	} else if e.Comma != nil {
		return *e.Comma
	} else if e.RBracket != nil {
		return *e.RBracket
	} else if e.CheckValue != nil {
		return *e.CheckValue
	}
	return Position{}
}
func (e *ZCIndexExpression) End() Position {
	if e.CheckValue != nil {
		return deltaPos(*e.CheckValue, len("?"))
	} else if e.RBracket != nil {
		return deltaPos(*e.RBracket, len("]"))
	} else if e.Comma != nil {
		return deltaPos(*e.Comma, len(","))
	} else if e.CheckIndex != nil {
		return deltaPos(*e.CheckIndex, len("?"))
	} else if e.Index != nil {
		return e.Index.End()
	} else if e.LBracket != nil {
		return deltaPos(*e.LBracket, len("["))
	}
	return Position{}
}

func (*ZCIndexExpression) _node()               {}
func (*ZCIndexExpression) _zeroCoalescingNode() {}

// ============================================================================
// Dot Ident Expression
// ======================================================================================

// ZCSelectorExpression is a dot followed by a Go identifier.
type ZCSelectorExpression struct {
	Dot   *Position // of the dot
	Ident *Ident
	Check *Position
}

var _ ZeroCoalescingNode = (*ZCSelectorExpression)(nil)

func (e *ZCSelectorExpression) Start() Position {
	if e.Dot != nil {
		return *e.Dot
	} else if e.Ident != nil {
		return e.Ident.Start()
	} else if e.Check != nil {
		return *e.Check
	}
	return Position{}
}
func (e *ZCSelectorExpression) End() Position {
	if e.Check != nil {
		return deltaPos(*e.Check, len("?"))
	} else if e.Ident != nil {
		return e.Ident.End()
	} else if e.Dot != nil {
		return deltaPos(*e.Dot, len("."))
	}
	return Position{}
}
func (*ZCSelectorExpression) _node()               {}
func (*ZCSelectorExpression) _zeroCoalescingNode() {}

// ============================================================================
// Paren Expression
// ======================================================================================

// ZCParenExpression is the paren part of a function call or a type cast.
type ZCParenExpression struct {
	LParen *Position
	Args   []*Expression // not a ZeroCoalescing
	RParen *Position
	Check  *Position
}

var _ ZeroCoalescingNode = (*ZCParenExpression)(nil)

func (e *ZCParenExpression) Start() Position {
	if e.LParen != nil {
		return *e.LParen
	}
	for _, arg := range e.Args {
		if arg != nil {
			return arg.Start()
		}
	}
	if e.RParen != nil {
		return *e.RParen
	} else if e.Check != nil {
		return *e.Check
	}
	return Position{}
}
func (e *ZCParenExpression) End() Position {
	if e.Check != nil {
		return deltaPos(*e.Check, len("?"))
	} else if e.RParen != nil {
		return *e.RParen
	}
	for _, arg := range slices.Backward(e.Args) {
		if arg != nil {
			return arg.End()
		}
	}
	if e.LParen != nil {
		return deltaPos(*e.LParen, len("("))
	}
	return Position{}
}

func (*ZCParenExpression) _node()               {}
func (*ZCParenExpression) _zeroCoalescingNode() {}

// ============================================================================
// Type Assertion Expression
// ======================================================================================

// ZCTypeAssertionExpression is a type assertion.
type ZCTypeAssertionExpression struct {
	Dot          *Position
	LParen       *Position
	PointerCount int
	Type         FullIdent
	CheckType    *Position
	RParen       *Position
	CheckValue   *Position
}

var _ ZeroCoalescingNode = (*ZCTypeAssertionExpression)(nil)

func (e *ZCTypeAssertionExpression) Start() Position {
	if e.Dot != nil {
		return *e.Dot
	} else if e.LParen != nil {
		return *e.LParen
	} else if e.CheckType != nil {
		return *e.CheckType
	} else if e.Type != nil {
		return e.Type.Start()
	} else if e.CheckValue != nil {
		return *e.CheckValue
	}
	return Position{}
}
func (e *ZCTypeAssertionExpression) End() Position {
	if e.CheckValue != nil {
		return deltaPos(*e.CheckValue, len("?"))
	} else if e.RParen != nil {
		return deltaPos(*e.RParen, len(")"))
	} else if e.CheckType != nil {
		return deltaPos(*e.CheckType, len("?"))
	} else if e.Type != nil {
		return e.Type.End()
	} else if e.LParen != nil {
		return deltaPos(*e.LParen, len("("))
	} else if e.Dot != nil {
		return deltaPos(*e.Dot, len("."))
	}
	return Position{}
}

func (*ZCTypeAssertionExpression) _node()               {}
func (*ZCTypeAssertionExpression) _zeroCoalescingNode() {}
