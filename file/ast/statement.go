package ast

import "slices"

// ============================================================================
// Statement
// ======================================================================================

// A Statement represents a line of Go code with corgi enhancements.
type Statement struct {
	Nodes  Code
	Parsed ParsedStatement // may be nil; see doc of ParsedStatement
}

var _ Node = (*Statement)(nil)

func (s *Statement) Start() Position {
	if s.Nodes != nil {
		if start := s.Nodes.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (s *Statement) End() Position {
	if s.Nodes != nil {
		if end := s.Nodes.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (s *Statement) Walk(w func(Node)) {
	if s.Nodes != nil {
		w(s.Nodes)
	}
	if s.Parsed != nil {
		w(s.Parsed)
	}
}

func (*Statement) _node() {}

// ============================================================================
// Parsed Statement
// ======================================================================================

// ParsedStatement is a subset of valid Go statements with corgi enhancements
// that we have an AST representation for.
// This is usually for the subset of statements that we need to properly
// identify later on.
//
// It is a pointer to either [Return], [Break], [Continue], [Fallthrough],
// [Defer], [IncDec], [Label], [ZeroCoalescingAssignment], [ConstDeclaration],
// [VarDeclaration], [ShortVarDeclaration], [Label], or [Assignment].
type ParsedStatement interface {
	Node
	_parsedStatement()
}

// if this is changed, change the comment above
var (
	_ ParsedStatement = (*Return)(nil)
	_ ParsedStatement = (*Break)(nil)
	_ ParsedStatement = (*Continue)(nil)
	_ ParsedStatement = (*Fallthrough)(nil)
	_ ParsedStatement = (*Defer)(nil)
	_ ParsedStatement = (*IncDec)(nil)
	_ ParsedStatement = (*Label)(nil)
	_ ParsedStatement = (*ZeroCoalescingAssignment)(nil)
	_ ParsedStatement = (*ConstDeclaration)(nil)
	_ ParsedStatement = (*VarDeclaration)(nil)
	_ ParsedStatement = (*ShortVarDeclaration)(nil)
	_ ParsedStatement = (*Label)(nil)
	_ ParsedStatement = (*Assignment)(nil)
)

// ============================================================================
// Simple Statement
// ======================================================================================

// A SimpleStatement represents the Go spec equivalent with corgi enhancements.
type SimpleStatement struct {
	Nodes  Code
	Parsed ParsedSimpleStatement // may be nil; see doc of ParsedStatement
}

var _ Node = (*SimpleStatement)(nil)

func (s *SimpleStatement) Start() Position {
	if s.Nodes != nil {
		if start := s.Nodes.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (s *SimpleStatement) End() Position {
	if s.Nodes != nil {
		if end := s.Nodes.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (s *SimpleStatement) Walk(w func(Node)) {
	if s.Nodes != nil {
		w(s.Nodes)
	}
	if s.Parsed != nil {
		w(s.Parsed)
	}
}

func (*SimpleStatement) _node() {}

// ParsedSimpleStatement is a subset of valid Go simple statements with corgi
// enhancements that we have an AST representation for.
// This is usually for the subset of statements that we need to properly
// identify later on.
//
// It is a pointer to either [IncDec], [ZeroCoalescingAssignment],
// [ShortVarDeclaration], or [Assignment].
type ParsedSimpleStatement interface {
	ParsedStatement
	_parsedSimpleStatement()
}

// if this is changed, change the comment above
var (
	_ ParsedSimpleStatement = (*IncDec)(nil)
	_ ParsedSimpleStatement = (*ZeroCoalescingAssignment)(nil)
	_ ParsedSimpleStatement = (*ShortVarDeclaration)(nil)
	_ ParsedSimpleStatement = (*Assignment)(nil)
)

// ============================================================================
// Return
// ======================================================================================

type Return struct {
	Return *Position
	Error  *Expression
}

var _ ParsedStatement = (*Return)(nil)

func (r *Return) Start() Position {
	if r.Return != nil {
		return *r.Return
	}
	if r.Error != nil {
		if start := r.Error.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (r *Return) End() Position {
	if r.Error != nil {
		if end := r.Error.End(); end != NoPosition {
			return end
		}
	}
	if r.Return != nil {
		return deltaPos(*r.Return, len("return"))
	}
	return NoPosition
}

func (r *Return) Walk(w func(Node)) {
	if r.Error != nil {
		w(r.Error)
	}
}

func (*Return) _node()            {}
func (*Return) _parsedStatement() {}

// ============================================================================
// Break
// ======================================================================================

type Break struct {
	Break *Position
	Label *Identifier // optional
}

var _ ParsedStatement = (*Break)(nil)

func (b *Break) Start() Position {
	if b.Break != nil {
		return *b.Break
	}
	if b.Label != nil {
		if start := b.Label.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (b *Break) End() Position {
	if b.Label != nil {
		if end := b.Label.End(); end != NoPosition {
			return end
		}
	}
	if b.Break != nil {
		return deltaPos(*b.Break, len("break"))
	}
	return NoPosition
}

func (b *Break) Walk(w func(Node)) {
	if b.Label != nil {
		w(b.Label)
	}
}

func (*Break) _node()            {}
func (*Break) _parsedStatement() {}

// ============================================================================
// Continue
// ======================================================================================

type Continue struct {
	Continue *Position
	Label    *Identifier // optional
}

var _ ParsedStatement = (*Continue)(nil)

func (c *Continue) Start() Position {
	if c.Label != nil {
		if start := c.Label.Start(); start != NoPosition {
			return start
		}
	}
	if c.Continue != nil {
		return *c.Continue
	}
	return NoPosition
}

func (c *Continue) End() Position {
	if c.Label != nil {
		if end := c.Label.End(); end != NoPosition {
			return end
		}
	}
	if c.Continue != nil {
		return deltaPos(*c.Continue, len("continue"))
	}
	return NoPosition
}

func (c *Continue) Walk(w func(Node)) {
	if c.Label != nil {
		w(c.Label)
	}
}

func (*Continue) _node()            {}
func (*Continue) _parsedStatement() {}

// ============================================================================
// Fallthrough
// ======================================================================================

type Fallthrough struct {
	Fallthrough *Position
	Label       *Identifier // optional
}

var _ ParsedStatement = (*Fallthrough)(nil)

func (f *Fallthrough) Start() Position {
	if f.Fallthrough != nil {
		return *f.Fallthrough
	}
	if f.Label != nil {
		if start := f.Label.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (f *Fallthrough) End() Position {
	if f.Label != nil {
		if end := f.Label.End(); end != NoPosition {
			return end
		}
	}
	if f.Fallthrough != nil {
		return deltaPos(*f.Fallthrough, len("fallthrough"))
	}
	return NoPosition
}

func (f *Fallthrough) Walk(w func(Node)) {
	if f.Label != nil {
		w(f.Label)
	}
}

func (*Fallthrough) _node()            {}
func (*Fallthrough) _parsedStatement() {}

// ============================================================================
// Defer
// ======================================================================================

type Defer struct {
	Defer      *Position
	Expression *Expression
}

var _ ParsedStatement = (*Defer)(nil)

func (d *Defer) Start() Position {
	if d.Expression != nil {
		if start := d.Expression.Start(); start != NoPosition {
			return start
		}
	}
	if d.Defer != nil {
		return *d.Defer
	}
	return NoPosition
}

func (d *Defer) End() Position {
	if d.Expression != nil {
		if end := d.Expression.End(); end != NoPosition {
			return end
		}
	}
	if d.Defer != nil {
		return deltaPos(*d.Defer, len("defer"))
	}
	return NoPosition
}

func (d *Defer) Walk(w func(Node)) {
	if d.Expression != nil {
		w(d.Expression)
	}
}

func (*Defer) _node()            {}
func (*Defer) _parsedStatement() {}

// ============================================================================
// Const Declaration
// ======================================================================================

type ConstDeclaration struct {
	Const  *Position
	LParen *Position // nil if single spec
	Specs  []*ConstSpec
	RParen *Position
}

var _ ParsedStatement = (*ConstDeclaration)(nil)

func (c *ConstDeclaration) Start() Position {
	if c.Const != nil {
		return *c.Const
	}
	if c.LParen != nil {
		return *c.LParen
	}
	for _, spec := range c.Specs {
		if spec != nil {
			if start := spec.Start(); start != NoPosition {
				return start
			}
		}
	}
	if c.RParen != nil {
		return *c.RParen
	}
	return NoPosition
}

func (c *ConstDeclaration) End() Position {
	if c.RParen != nil {
		return deltaPos(*c.RParen, len(")"))
	}
	for _, spec := range slices.Backward(c.Specs) {
		if spec != nil {
			if end := spec.End(); end != NoPosition {
				return end
			}
		}
	}
	if c.LParen != nil {
		return deltaPos(*c.LParen, len("("))
	}
	if c.Const != nil {
		return deltaPos(*c.Const, len("const"))
	}
	return NoPosition
}

func (c *ConstDeclaration) Walk(w func(Node)) {
	for _, spec := range c.Specs {
		if spec != nil {
			w(spec)
		}
	}
}

func (*ConstDeclaration) _node()            {}
func (*ConstDeclaration) _parsedStatement() {}

// ===================================== Const Spec =====================================

type ConstSpec struct {
	Names     []*Identifier
	Type      *Type // optional
	EqualSign *Position
	Values    []*Expression
}

var _ Node = (*ConstSpec)(nil)

func (c *ConstSpec) Start() Position {
	for _, name := range c.Names {
		if name != nil {
			if start := name.Start(); start != NoPosition {
				return start
			}
		}
	}
	if c.Type != nil {
		if start := c.Type.Start(); start != NoPosition {
			return start
		}
	}
	if c.EqualSign != nil {
		return *c.EqualSign
	}
	for _, value := range slices.Backward(c.Values) {
		if value != nil {
			if start := value.Start(); start != NoPosition {
				return start
			}
		}
	}
	return NoPosition
}

func (c *ConstSpec) End() Position {
	for _, value := range slices.Backward(c.Values) {
		if value != nil {
			if end := value.End(); end != NoPosition {
				return end
			}
		}
	}
	if c.EqualSign != nil {
		return deltaPos(*c.EqualSign, len("="))
	}
	if c.Type != nil {
		if end := c.Type.End(); end != NoPosition {
			return end
		}
	}
	for _, name := range slices.Backward(c.Names) {
		if name != nil {
			if end := name.End(); end != NoPosition {
				return end
			}
		}
	}
	return NoPosition
}

func (c *ConstSpec) Walk(w func(Node)) {
	for _, name := range c.Names {
		if name != nil {
			w(name)
		}
	}
	if c.Type != nil {
		w(c.Type)
	}
	for _, value := range c.Values {
		if value != nil {
			w(value)
		}
	}
}

func (*ConstSpec) _node() {}

// ============================================================================
// Var Declaration
// ======================================================================================

type VarDeclaration struct {
	Var    *Position
	LParen *Position // nil if single spec
	Specs  []*VarSpec
	RParen *Position
}

var _ ParsedStatement = (*VarDeclaration)(nil)

func (c *VarDeclaration) Start() Position {
	if c.Var != nil {
		return *c.Var
	}
	if c.LParen != nil {
		return *c.LParen
	}
	for _, spec := range c.Specs {
		if spec != nil {
			if start := spec.Start(); start != NoPosition {
				return start
			}
		}
	}
	if c.RParen != nil {
		return *c.RParen
	}
	return NoPosition
}

func (c *VarDeclaration) End() Position {
	if c.RParen != nil {
		return deltaPos(*c.RParen, len(")"))
	}
	for _, spec := range slices.Backward(c.Specs) {
		if spec != nil {
			if end := spec.End(); end != NoPosition {
				return end
			}
		}
	}
	if c.LParen != nil {
		return deltaPos(*c.LParen, len("("))
	}
	if c.Var != nil {
		return deltaPos(*c.Var, len("var"))
	}
	return NoPosition
}

func (c *VarDeclaration) Walk(w func(Node)) {
	for _, spec := range c.Specs {
		if spec != nil {
			w(spec)
		}
	}
}

func (*VarDeclaration) _node()            {}
func (*VarDeclaration) _parsedStatement() {}

// ===================================== Var Spec =====================================

type VarSpec struct {
	Names     []*Identifier
	Type      *Type         // optional
	EqualSign *Position     // optional if type
	Values    []*Expression // optional if type
}

var _ Node = (*VarSpec)(nil)

func (c *VarSpec) Start() Position {
	for _, name := range c.Names {
		if name != nil {
			if start := name.Start(); start != NoPosition {
				return start
			}
		}
	}
	if c.Type != nil {
		if start := c.Type.Start(); start != NoPosition {
			return start
		}
	}
	if c.EqualSign != nil {
		return *c.EqualSign
	}
	for _, value := range slices.Backward(c.Values) {
		if value != nil {
			if start := value.Start(); start != NoPosition {
				return start
			}
		}
	}
	return NoPosition
}

func (c *VarSpec) End() Position {
	for _, value := range slices.Backward(c.Values) {
		if value != nil {
			if end := value.End(); end != NoPosition {
				return end
			}
		}
	}
	if c.EqualSign != nil {
		return deltaPos(*c.EqualSign, len("="))
	}
	if c.Type != nil {
		if end := c.Type.End(); end != NoPosition {
			return end
		}
	}
	for _, name := range slices.Backward(c.Names) {
		if name != nil {
			if end := name.End(); end != NoPosition {
				return end
			}
		}
	}
	return NoPosition
}

func (c *VarSpec) Walk(w func(Node)) {
	for _, name := range c.Names {
		if name != nil {
			w(name)
		}
	}
	if c.Type != nil {
		w(c.Type)
	}
	for _, value := range c.Values {
		if value != nil {
			w(value)
		}
	}
}

func (*VarSpec) _node() {}

// ============================================================================
// Zero Coalescing Assignment
// ======================================================================================

type ZeroCoalescingAssignment struct {
	ValueExpression *Expression
	VarComma        *Position   // nil if no ok expression
	OkExpression    *Expression // optional
	Colon           *Position   // nil no declaration
	EqualSign       *Position
	Expression      *ZeroCoalescing // optional default
}

var (
	_ ParsedStatement       = (*ZeroCoalescingAssignment)(nil)
	_ ParsedSimpleStatement = (*ZeroCoalescingAssignment)(nil)
)

func (a *ZeroCoalescingAssignment) Start() Position {
	if a.ValueExpression != nil {
		if start := a.ValueExpression.Start(); start != NoPosition {
			return start
		}
	}
	if a.VarComma != nil {
		return *a.VarComma
	}
	if a.OkExpression != nil {
		if start := a.OkExpression.Start(); start != NoPosition {
			return start
		}
	}
	if a.Colon != nil {
		return *a.Colon
	}
	if a.EqualSign != nil {
		return *a.EqualSign
	}
	if a.Expression != nil {
		if start := a.Expression.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (a *ZeroCoalescingAssignment) End() Position {
	if a.Expression != nil {
		if end := a.Expression.End(); end != NoPosition {
			return end
		}
	}
	if a.EqualSign != nil {
		return deltaPos(*a.EqualSign, len("="))
	}
	if a.Colon != nil {
		return deltaPos(*a.Colon, len(":"))
	}
	if a.OkExpression != nil {
		if end := a.OkExpression.End(); end != NoPosition {
			return end
		}
	}
	if a.VarComma != nil {
		return deltaPos(*a.VarComma, len(","))
	}
	if a.ValueExpression != nil {
		if end := a.ValueExpression.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (a *ZeroCoalescingAssignment) Walk(w func(Node)) {
	if a.ValueExpression != nil {
		w(a.ValueExpression)
	}
	if a.OkExpression != nil {
		w(a.OkExpression)
	}
	if a.Expression != nil {
		w(a.Expression)
	}
}

func (*ZeroCoalescingAssignment) _node()                  {}
func (*ZeroCoalescingAssignment) _parsedStatement()       {}
func (*ZeroCoalescingAssignment) _parsedSimpleStatement() {}

// ============================================================================
// IncDec
// ======================================================================================

type IncDec struct {
	Expression *Expression
	IncrPos    *Position // nil if decrement
	DecrPos    *Position // nil if increment
}

var (
	_ ParsedStatement       = (*IncDec)(nil)
	_ ParsedSimpleStatement = (*IncDec)(nil)
)

func (i *IncDec) Start() Position {
	if i.DecrPos != nil {
		return *i.DecrPos
	}
	if i.IncrPos != nil {
		return *i.IncrPos
	}
	if i.Expression != nil {
		if start := i.Expression.Start(); start != NoPosition {
			return start
		}
	}
	return NoPosition
}

func (i *IncDec) End() Position {
	if i.DecrPos != nil {
		return deltaPos(*i.DecrPos, len("--"))
	}
	if i.IncrPos != nil {
		return deltaPos(*i.IncrPos, len("++"))
	}
	if i.Expression != nil {
		if end := i.Expression.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (i *IncDec) Walk(w func(Node)) {
	if i.Expression != nil {
		w(i.Expression)
	}
}

func (*IncDec) _node()                  {}
func (*IncDec) _parsedStatement()       {}
func (*IncDec) _parsedSimpleStatement() {}

// ============================================================================
// Short Var Declaration
// ======================================================================================

type ShortVarDeclaration struct {
	Names          []*Identifier
	ColonEqualSign *Position
	Values         []*Expression
}

var (
	_ ParsedStatement       = (*ShortVarDeclaration)(nil)
	_ ParsedSimpleStatement = (*ShortVarDeclaration)(nil)
)

func (s *ShortVarDeclaration) Start() Position {
	for _, name := range s.Names {
		if name != nil {
			if start := name.Start(); start != NoPosition {
				return start
			}
		}
	}
	if s.ColonEqualSign != nil {
		return *s.ColonEqualSign
	}
	for _, value := range slices.Backward(s.Values) {
		if value != nil {
			if start := value.Start(); start != NoPosition {
				return start
			}
		}
	}
	return NoPosition
}

func (s *ShortVarDeclaration) End() Position {
	for _, value := range slices.Backward(s.Values) {
		if value != nil {
			if end := value.End(); end != NoPosition {
				return end
			}
		}
	}
	if s.ColonEqualSign != nil {
		return deltaPos(*s.ColonEqualSign, len(":="))
	}
	for _, name := range slices.Backward(s.Names) {
		if name != nil {
			if end := name.End(); end != NoPosition {
				return end
			}
		}
	}
	return NoPosition
}

func (s *ShortVarDeclaration) Walk(w func(Node)) {
	for _, name := range s.Names {
		if name != nil {
			w(name)
		}
	}
	for _, value := range s.Values {
		if value != nil {
			w(value)
		}
	}
}

func (*ShortVarDeclaration) _node()                  {}
func (*ShortVarDeclaration) _parsedStatement()       {}
func (*ShortVarDeclaration) _parsedSimpleStatement() {}

// ============================================================================
// Label
// ======================================================================================

type Label struct {
	Name  *Identifier
	Colon *Position
}

var _ ParsedStatement = (*Label)(nil)

func (l *Label) Start() Position {
	if l.Name != nil {
		if start := l.Name.Start(); start != NoPosition {
			return start
		}
	}
	if l.Colon != nil {
		return *l.Colon
	}
	return NoPosition
}

func (l *Label) End() Position {
	if l.Colon != nil {
		return deltaPos(*l.Colon, len(":"))
	}
	if l.Name != nil {
		if end := l.Name.End(); end != NoPosition {
			return end
		}
	}
	return NoPosition
}

func (l *Label) Walk(w func(Node)) {
	if l.Name != nil {
		w(l.Name)
	}
}

func (*Label) _node()            {}
func (*Label) _parsedStatement() {}

// ============================================================================
// Assignment
// ======================================================================================

type Assignment struct {
	LHS              []*Expression
	Operator         string
	OperatorPosition *Position
	RHS              []*Expression
}

var (
	_ ParsedStatement       = (*Assignment)(nil)
	_ ParsedSimpleStatement = (*Assignment)(nil)
)

func (a *Assignment) Start() Position {
	for _, e := range a.LHS {
		if e != nil {
			if start := e.Start(); start != NoPosition {
				return start
			}
		}
	}
	if a.OperatorPosition != nil {
		return *a.OperatorPosition
	}
	for _, e := range a.RHS {
		if e != nil {
			if start := e.Start(); start != NoPosition {
				return start
			}
		}
	}
	return NoPosition
}

func (a *Assignment) End() Position {
	for _, e := range slices.Backward(a.RHS) {
		if e != nil {
			if end := e.End(); end != NoPosition {
				return end
			}
		}
	}
	if a.OperatorPosition != nil {
		return deltaPos(*a.OperatorPosition, len(a.Operator)+len("="))
	}
	for _, e := range slices.Backward(a.LHS) {
		if e != nil {
			if end := e.End(); end != NoPosition {
				return end
			}
		}
	}
	return NoPosition
}

func (a *Assignment) Walk(w func(Node)) {
	for _, e := range a.LHS {
		if e != nil {
			w(e)
		}
	}
	for _, e := range a.RHS {
		if e != nil {
			w(e)
		}
	}
}

func (*Assignment) _node()                  {}
func (*Assignment) _parsedStatement()       {}
func (*Assignment) _parsedSimpleStatement() {}
