package ast

import "slices"

// ============================================================================
// Statement
// ======================================================================================

// A Statement represents a line of Go code with corgi enhancements.
type Statement struct {
	Code   Code
	Parsed ParsedStatement // may be nil; see doc of ParsedStatement
}

var _ Node = (*Statement)(nil)

func (s *Statement) Start() Position {
	if s.Code != nil {
		return s.Code.Start()
	}
	return Position{}
}

func (s *Statement) End() Position {
	if s.Code != nil {
		return s.Code.End()
	}
	return Position{}
}

func (s *Statement) Walk(w func(Node)) {
	if s.Code != nil {
		w(s.Code)
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
// that we have an ComponentAST representation for.
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
	Code   Code
	Parsed ParsedSimpleStatement // may be nil; see doc of ParsedStatement
}

var _ Node = (*SimpleStatement)(nil)

func (s *SimpleStatement) Start() Position {
	if s.Code != nil {
		return s.Code.Start()
	}
	return Position{}
}

func (s *SimpleStatement) End() Position {
	if s.Code != nil {
		return s.Code.End()
	}
	return Position{}
}

func (s *SimpleStatement) Walk(w func(Node)) {
	if s.Code != nil {
		w(s.Code)
	}
	if s.Parsed != nil {
		w(s.Parsed)
	}
}

func (*SimpleStatement) _node() {}

// ParsedSimpleStatement is a subset of valid Go simple statements with corgi
// enhancements that we have an ComponentAST representation for.
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
	} else if r.Error != nil {
		return r.Error.Start()
	}
	return Position{}
}

func (r *Return) End() Position {
	if r.Error != nil {
		return r.Error.End()
	} else if r.Return != nil {
		return deltaPos(*r.Return, len("return"))
	}
	return Position{}
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
	Label *Ident // optional
}

var _ ParsedStatement = (*Break)(nil)

func (b *Break) Start() Position {
	if b.Break != nil {
		return *b.Break
	} else if b.Label != nil {
		return b.Label.Start()
	}
	return Position{}
}

func (b *Break) End() Position {
	if b.Label != nil {
		return b.Label.End()
	} else if b.Break != nil {
		return deltaPos(*b.Break, len("break"))
	}
	return Position{}
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
	Label    *Ident // optional
}

var _ ParsedStatement = (*Continue)(nil)

func (c *Continue) Start() Position {
	if c.Label != nil {
		return c.Label.Start()
	} else if c.Continue != nil {
		return *c.Continue
	}
	return Position{}
}

func (c *Continue) End() Position {
	if c.Label != nil {
		return c.Label.End()
	} else if c.Continue != nil {
		return deltaPos(*c.Continue, len("continue"))
	}
	return Position{}
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
	Label       *Ident // optional
}

var _ ParsedStatement = (*Fallthrough)(nil)

func (f *Fallthrough) Start() Position {
	if f.Fallthrough != nil {
		return *f.Fallthrough
	} else if f.Label != nil {
		return f.Label.Start()
	}
	return Position{}
}

func (f *Fallthrough) End() Position {
	if f.Label != nil {
		return f.Label.End()
	} else if f.Fallthrough != nil {
		return deltaPos(*f.Fallthrough, len("fallthrough"))
	}
	return Position{}
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
		return d.Expression.Start()
	}
	return *d.Defer
}

func (d *Defer) End() Position {
	if d.Expression != nil {
		return d.Expression.End()
	} else if d.Defer != nil {
		return deltaPos(*d.Defer, len("defer"))
	}
	return Position{}
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
	} else if c.LParen != nil {
		return *c.LParen
	}
	for _, spec := range c.Specs {
		if spec != nil {
			return spec.Start()
		}
	}
	if c.RParen != nil {
		return *c.RParen
	}
	return Position{}
}

func (c *ConstDeclaration) End() Position {
	if c.RParen != nil {
		return deltaPos(*c.RParen, len(")"))
	}
	for _, spec := range slices.Backward(c.Specs) {
		if spec != nil {
			return spec.End()
		}
	}
	if c.LParen != nil {
		return deltaPos(*c.LParen, len("("))
	} else if c.Const != nil {
		return deltaPos(*c.Const, len("const"))
	}
	return Position{}
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
	Names     []*Ident
	Type      *Type // optional
	EqualSign *Position
	Values    []*Expression
}

var _ Node = (*ConstSpec)(nil)

func (c *ConstSpec) Start() Position {
	for _, name := range c.Names {
		if name != nil {
			return name.Start()
		}
	}
	if c.Type != nil {
		return c.Type.Start()
	} else if c.EqualSign != nil {
		return *c.EqualSign
	}
	for _, value := range slices.Backward(c.Values) {
		if value != nil {
			return value.Start()
		}
	}
	return Position{}
}

func (c *ConstSpec) End() Position {
	for _, value := range slices.Backward(c.Values) {
		if value != nil {
			return value.End()
		}
	}
	if c.EqualSign != nil {
		return deltaPos(*c.EqualSign, len("="))
	} else if c.Type != nil {
		return c.Type.End()
	}
	for _, name := range slices.Backward(c.Names) {
		if name != nil {
			return name.End()
		}
	}
	return Position{}
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
	} else if c.LParen != nil {
		return *c.LParen
	}
	for _, spec := range c.Specs {
		if spec != nil {
			return spec.Start()
		}
	}
	if c.RParen != nil {
		return *c.RParen
	}
	return Position{}
}

func (c *VarDeclaration) End() Position {
	if c.RParen != nil {
		return deltaPos(*c.RParen, len(")"))
	}
	for _, spec := range slices.Backward(c.Specs) {
		if spec != nil {
			return spec.End()
		}
	}
	if c.LParen != nil {
		return deltaPos(*c.LParen, len("("))
	} else if c.Var != nil {
		return deltaPos(*c.Var, len("var"))
	}
	return Position{}
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
	Names     []*Ident
	Type      *Type         // optional
	EqualSign *Position     // optional if type
	Values    []*Expression // optional if type
}

var _ Node = (*VarSpec)(nil)

func (c *VarSpec) Start() Position {
	for _, name := range c.Names {
		if name != nil {
			return name.Start()
		}
	}
	if c.Type != nil {
		return c.Type.Start()
	} else if c.EqualSign != nil {
		return *c.EqualSign
	}
	for _, value := range slices.Backward(c.Values) {
		if value != nil {
			return value.Start()
		}
	}
	return Position{}
}

func (c *VarSpec) End() Position {
	for _, value := range slices.Backward(c.Values) {
		if value != nil {
			return value.End()
		}
	}
	if c.EqualSign != nil {
		return deltaPos(*c.EqualSign, len("="))
	} else if c.Type != nil {
		return c.Type.End()
	}
	for _, name := range slices.Backward(c.Names) {
		if name != nil {
			return name.End()
		}
	}
	return Position{}
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
	switch {
	case a.ValueExpression != nil:
		return a.ValueExpression.Start()
	case a.VarComma != nil:
		return *a.VarComma
	case a.OkExpression != nil:
		return a.OkExpression.Start()
	case a.Colon != nil:
		return *a.Colon
	case a.EqualSign != nil:
		return *a.EqualSign
	case a.Expression != nil:
		return a.Expression.Start()
	}
	return Position{}
}

func (a *ZeroCoalescingAssignment) End() Position {
	switch {
	case a.Expression != nil:
		return a.Expression.End()
	case a.EqualSign != nil:
		return deltaPos(*a.EqualSign, len("="))
	case a.Colon != nil:
		return deltaPos(*a.Colon, len(":"))
	case a.OkExpression != nil:
		return a.OkExpression.End()
	case a.VarComma != nil:
		return deltaPos(*a.VarComma, len(","))
	case a.ValueExpression != nil:
		return a.ValueExpression.End()
	}
	return Position{}
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
	return *i.IncrPos
}

func (i *IncDec) End() Position {
	switch {
	case i.DecrPos != nil:
		return deltaPos(*i.DecrPos, len("--"))
	case i.IncrPos != nil:
		return deltaPos(*i.IncrPos, len("++"))
	case i.Expression != nil:
		return i.Expression.End()
	}
	return Position{}
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
	Names          []*Ident
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
			return name.Start()
		}
	}
	if s.ColonEqualSign != nil {
		return *s.ColonEqualSign
	}
	for _, value := range slices.Backward(s.Values) {
		if value != nil {
			return value.Start()
		}
	}
	return Position{}
}

func (s *ShortVarDeclaration) End() Position {
	if len(s.Values) > 0 {
		return s.Values[len(s.Values)-1].End()
	}
	return s.Names[len(s.Names)-1].End()
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
	Name  *Ident
	Colon *Position
}

var _ ParsedStatement = (*Label)(nil)

func (l *Label) Start() Position {
	if l.Name != nil {
		return l.Name.Start()
	} else if l.Colon != nil {
		return *l.Colon
	}
	return Position{}
}

func (l *Label) End() Position {
	if l.Colon != nil {
		return deltaPos(*l.Colon, len(":"))
	} else if l.Name != nil {
		return l.Name.End()
	}
	return Position{}
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
	SpecialOperator  string // optional, ("+", ">>", "*", etc.)
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
			return e.Start()
		}
	}
	if a.OperatorPosition != nil {
		return *a.OperatorPosition
	}
	for _, e := range a.RHS {
		if e != nil {
			return e.Start()
		}
	}
	return Position{}
}

func (a *Assignment) End() Position {
	for _, e := range slices.Backward(a.RHS) {
		if e != nil {
			return e.End()
		}
	}
	if a.OperatorPosition != nil {
		return deltaPos(*a.OperatorPosition, len(a.SpecialOperator)+len("="))
	}
	for _, e := range slices.Backward(a.LHS) {
		if e != nil {
			return e.End()
		}
	}
	return Position{}
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
