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
	}
	if c.Root != nil {
		if start := c.Root.Start(); start != NoPosition {
			return start
		}
	}
	if c.CheckRoot != nil {
		return *c.CheckRoot
	}
	for _, node := range c.Chain {
		if node != nil {
			if start := node.Start(); start != NoPosition {
				return start
			}
		}
	}
	if c.Tilde != nil {
		return deltaPos(*c.Tilde, len("~"))
	}
	if c.Default != nil {
		if start := c.Default.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (c *ZeroCoalescing) End() Position {
	if c.Default != nil {
		if end := c.Default.End(); end != NoPosition {
			return end
		}
	}
	if c.Tilde != nil {
		return deltaPos(*c.Tilde, len("~"))
	}
	for _, node := range slices.Backward(c.Chain) {
		if node != nil {
			if end := node.End(); end != NoPosition {
				return end
			}
		}
	}
	if c.CheckRoot != nil {
		return deltaPos(*c.CheckRoot, len("?"))
	}
	if c.Root != nil {
		if end := c.Root.End(); end != NoPosition {
			return end
		}
	}
	if c.DerefPosition != nil {
		return deltaPos(*c.DerefPosition, c.DerefCount)
	}
	return NoPosition
}

func (c *ZeroCoalescing) Walk(w func(Node)) {
	if c.Root != nil {
		w(c.Root)
	}
	for _, node := range c.Chain {
		if node != nil {
			w(node)
		}
	}
	if c.Default != nil {
		w(c.Default)
	}
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
	}
	if e.Index != nil {
		if start := e.Index.Start(); start != NoPosition {
			return start
		}
	}
	switch {
	case e.CheckIndex != nil:
		return *e.CheckIndex
	case e.Comma != nil:
		return *e.Comma
	case e.RBracket != nil:
		return *e.RBracket
	case e.CheckValue != nil:
		return *e.CheckValue
	}
	return NoPosition
}

func (e *ZCIndexExpression) End() Position {
	switch {
	case e.CheckValue != nil:
		return deltaPos(*e.CheckValue, len("?"))
	case e.RBracket != nil:
		return deltaPos(*e.RBracket, len("]"))
	case e.Comma != nil:
		return deltaPos(*e.Comma, len(","))
	case e.CheckIndex != nil:
		return deltaPos(*e.CheckIndex, len("?"))
	}
	if e.Index != nil {
		if end := e.Index.End(); end != NoPosition {
			return end
		}
	}
	if e.LBracket != nil {
		return deltaPos(*e.LBracket, len("["))
	}
	return NoPosition
}

func (e *ZCIndexExpression) Walk(w func(Node)) {
	if e.Index != nil {
		w(e.Index)
	}
}

func (*ZCIndexExpression) _node()               {}
func (*ZCIndexExpression) _zeroCoalescingNode() {}

// ============================================================================
// Selector Expression
// ======================================================================================

// ZCSelectorExpression is a dot followed by a Go identifier.
type ZCSelectorExpression struct {
	Dot   *Position // of the dot
	Ident *Identifier
	Check *Position
}

var _ ZeroCoalescingNode = (*ZCSelectorExpression)(nil)

func (e *ZCSelectorExpression) Start() Position {
	if e.Dot != nil {
		return *e.Dot
	}
	if e.Ident != nil {
		if start := e.Ident.Start(); start != NoPosition {
			return start
		}
	}
	if e.Check != nil {
		return *e.Check
	}
	return NoPosition
}

func (e *ZCSelectorExpression) End() Position {
	if e.Check != nil {
		return deltaPos(*e.Check, len("?"))
	}
	if e.Ident != nil {
		if end := e.Ident.End(); end != NoPosition {
			return end
		}
	}
	if e.Dot != nil {
		return deltaPos(*e.Dot, len("."))
	}
	return NoPosition
}

func (e *ZCSelectorExpression) Walk(w func(Node)) {
	if e.Ident != nil {
		w(e.Ident)
	}
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
			if start := arg.Start(); start != NoPosition {
				return start
			}
		}
	}
	if e.RParen != nil {
		return *e.RParen
	}
	if e.Check != nil {
		return *e.Check
	}
	return NoPosition
}

func (e *ZCParenExpression) End() Position {
	if e.Check != nil {
		return deltaPos(*e.Check, len("?"))
	}
	if e.RParen != nil {
		return *e.RParen
	}
	for _, arg := range slices.Backward(e.Args) {
		if arg != nil {
			if end := arg.End(); end != NoPosition {
				return end
			}
		}
	}
	if e.LParen != nil {
		return deltaPos(*e.LParen, len("("))
	}
	return NoPosition
}

func (e *ZCParenExpression) Walk(w func(Node)) {
	for _, arg := range e.Args {
		if arg != nil {
			w(arg)
		}
	}
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
	Type         FullIdentifier
	CheckType    *Position
	RParen       *Position
	CheckValue   *Position
}

var _ ZeroCoalescingNode = (*ZCTypeAssertionExpression)(nil)

func (e *ZCTypeAssertionExpression) Start() Position {
	switch {
	case e.Dot != nil:
		return *e.Dot
	case e.LParen != nil:
		return *e.LParen
	}
	if e.Type != nil {
		if start := e.Type.Start(); start != NoPosition {
			return start
		}
	}
	switch {
	case e.CheckType != nil:
		return *e.CheckType
	case e.CheckValue != nil:
		return *e.CheckValue
	}
	return NoPosition
}

func (e *ZCTypeAssertionExpression) End() Position {
	switch {
	case e.CheckValue != nil:
		return deltaPos(*e.CheckValue, len("?"))
	case e.RParen != nil:
		return deltaPos(*e.RParen, len(")"))
	case e.CheckType != nil:
		return deltaPos(*e.CheckType, len("?"))
	}
	if e.Type != nil {
		if end := e.Type.End(); end != NoPosition {
			return end
		}
	}
	switch {
	case e.Type != nil:
		return e.Type.End()
	case e.LParen != nil:
		return deltaPos(*e.LParen, len("("))
	case e.Dot != nil:
		return deltaPos(*e.Dot, len("."))
	}
	return NoPosition
}

func (e *ZCTypeAssertionExpression) Walk(w func(Node)) {
	if e.Type != nil {
		e.Type.Walk(w)
	}
}

func (*ZCTypeAssertionExpression) _node()               {}
func (*ZCTypeAssertionExpression) _zeroCoalescingNode() {}
