package ast

import (
	"slices"
	"strconv"
	"strings"
)

// Code is a sequence of Go code with corgi language extensions.
//
// Note that there may be multiple successive [GoCode] nodes in a [Code]
// object, especially if the code was generated from a ParsedStatement.
// There are no guarantees that between versions, the code will be split
// in the same way.
type Code []CodeNode

var _ Node = Code(nil)

func (c Code) Start() Position {
	for _, n := range c {
		if n != nil {
			if start := n.Start(); start != NoPosition {
				return start
			}
		}
	}
	return NoPosition
}

func (c Code) End() Position {
	for _, n := range slices.Backward(c) {
		if n != nil {
			if end := n.End(); end != NoPosition {
				return end
			}
		}
	}
	return NoPosition
}

func (c Code) Walk(w func(Node)) {
	for _, n := range c {
		if n != nil {
			w(n)
		}
	}
}

func (Code) _node() {}

type CodeNode interface {
	Node
	_codeNode()
}

// ============================================================================
// Go Code
// ======================================================================================

// GoCode is actual Go code, i.e., without any corgi language extensions.
type GoCode struct {
	Code     string
	Position *Position
}

var _ CodeNode = (*GoCode)(nil)

func (c *GoCode) Start() Position {
	if c.Position != nil {
		return *c.Position
	}
	return NoPosition
}

func (c *GoCode) End() Position {
	if c.Position != nil {
		return deltaPos(*c.Position, len([]rune(c.Code)))
	}
	return NoPosition
}
func (c *GoCode) Walk(func(Node)) {}

func (*GoCode) _node()     {}
func (*GoCode) _codeNode() {}

// ============================================================================
// Block Function
// ======================================================================================

// BlockFunction is the "built-in" block existence check function.
type BlockFunction struct {
	Block     *Position
	LParen    *Position
	BlockName *Identifier // optional for the default block
	RParen    *Position
}

var _ CodeNode = (*BlockFunction)(nil)

func (f *BlockFunction) Name() string {
	if f.BlockName != nil {
		return f.BlockName.Name
	}
	return ""
}

func (f *BlockFunction) Start() Position {
	if f.Block != nil {
		return *f.Block
	}
	if f.LParen != nil {
		return *f.LParen
	}
	if f.BlockName != nil {
		if start := f.BlockName.Start(); start != NoPosition {
			return start
		}
	}
	if f.RParen != nil {
		return *f.RParen
	}
	return NoPosition
}

func (f *BlockFunction) End() Position {
	if f.RParen != nil {
		return deltaPos(*f.RParen, len(")"))
	}
	if f.BlockName != nil {
		if end := f.BlockName.End(); end != NoPosition {
			return end
		}
	}
	if f.LParen != nil {
		return deltaPos(*f.LParen, len("("))
	}
	if f.Block != nil {
		return deltaPos(*f.Block, len("block"))
	}
	return NoPosition
}

func (f *BlockFunction) Walk(w func(Node)) {
	if f.BlockName != nil {
		w(f.BlockName)
	}
}

func (*BlockFunction) _node()     {}
func (*BlockFunction) _codeNode() {}

// ============================================================================
// Ternary
// ======================================================================================

type Ternary struct {
	QuestionMark *Position
	LParen       *Position
	Condition    *Expression
	TrueVal      *Expression
	FalseVal     *Expression
	RParen       *Position
}

var _ CodeNode = (*Ternary)(nil)

func (t *Ternary) Start() Position {
	if t.QuestionMark != nil {
		return *t.QuestionMark
	}
	if t.LParen != nil {
		return *t.LParen
	}
	if t.Condition != nil {
		if start := t.Condition.Start(); start != NoPosition {
			return start
		}
	}
	if t.TrueVal != nil {
		if start := t.TrueVal.Start(); start != NoPosition {
			return start
		}
	}
	if t.FalseVal != nil {
		if start := t.FalseVal.Start(); start != NoPosition {
			return start
		}
	}
	if t.RParen != nil {
		return *t.RParen
	}
	return NoPosition
}

func (t *Ternary) End() Position {
	if t.RParen != nil {
		return deltaPos(*t.RParen, len(")"))
	}
	if t.FalseVal != nil {
		if end := t.FalseVal.End(); end != NoPosition {
			return end
		}
	}
	if t.TrueVal != nil {
		if end := t.TrueVal.End(); end != NoPosition {
			return end
		}
	}
	if t.Condition != nil {
		if end := t.Condition.End(); end != NoPosition {
			return end
		}
	}
	if t.LParen != nil {
		return deltaPos(*t.LParen, len("("))
	}
	if t.QuestionMark != nil {
		return deltaPos(*t.QuestionMark, len("?"))
	}
	return NoPosition
}

func (t *Ternary) Walk(w func(Node)) {
	if t.Condition != nil {
		w(t.Condition)
	}
	if t.TrueVal != nil {
		w(t.TrueVal)
	}
	if t.FalseVal != nil {
		w(t.FalseVal)
	}
}

func (*Ternary) _node()     {}
func (*Ternary) _codeNode() {}

// ============================================================================
// String
// ======================================================================================

// String is either an [InterpretedString] or a [RawString].
type String interface {
	Node
	CodeNode
	_string()
	ConstantValue() (string, bool)
	Constant() bool
}

var (
	_ String = (*InterpretedString)(nil)
	_ String = (*RawString)(nil)
)

// ============================================================================
// InterpretedString
// ======================================================================================

// InterpretedString is a Go interpreted string literal extended to allow
// interpolation
type InterpretedString struct {
	Open     *Position
	Contents []InterpretedStringNode
	Close    *Position
}

var (
	_ CodeNode = (*InterpretedString)(nil)
	_ String   = (*InterpretedString)(nil)
)

func (s *InterpretedString) Start() Position {
	if s.Open != nil {
		return *s.Open
	}
	for _, n := range s.Contents {
		if n != nil {
			if start := n.Start(); start != NoPosition {
				return start
			}
		}
	}
	if s.Close != nil {
		return *s.Close
	}
	return NoPosition
}

func (s *InterpretedString) End() Position {
	if s.Close != nil {
		return deltaPos(*s.Close, len("\""))
	}
	for _, n := range slices.Backward(s.Contents) {
		if n != nil {
			if end := n.End(); end != NoPosition {
				return end
			}
		}
	}
	if s.Open != nil {
		return deltaPos(*s.Open, len("\""))
	}
	return NoPosition
}

func (s *InterpretedString) Walk(w func(Node)) {
	for _, n := range s.Contents {
		if n != nil {
			w(n)
		}
	}
}

func (s *InterpretedString) Constant() bool {
	for _, n := range s.Contents {
		if cn, _ := n.(ConstantInterpretedStringNode); cn == nil {
			return false
		}
	}
	return true
}

// ConstantValue unquotes the string, returning its constant value.
//
// If the string is not constant, i.e. if it contains any non-constant
// interpolation ConstantValue returns false.
// Furthermore, if the string contains any invalid escape sequences in its text
// nodes, ConstantValue again returns false.
//
// If the string is constant and valid, ConstantValue returns the unquoted
// value and true.
func (s *InterpretedString) ConstantValue() (string, bool) {
	if !s.Constant() {
		return "", false
	}

	var unq strings.Builder
	unq.Grow(int(s.End().Col - s.Start().Col))

	for _, n := range s.Contents {
		val, ok := n.(ConstantInterpretedStringNode).ConstantValue() //nolint:errcheck
		if !ok {
			return "", false
		}
		unq.WriteString(val)
	}
	return unq.String(), true
}

func (*InterpretedString) _node()     {}
func (*InterpretedString) _string()   {}
func (*InterpretedString) _codeNode() {}

// ============================================================================
// Interpreted String Node
// ======================================================================================

// InterpretedStringNode is either [StringInterpolation] or a pointer to
// [InterpretedStringText].
type InterpretedStringNode interface {
	Node
	_interpretedStringNode()
}

// if this is changed, change the comment above
var (
	_ InterpretedStringNode = (*InterpretedStringText)(nil)
	_ InterpretedStringNode = StringInterpolation(nil)
)

// ConstantInterpretedStringNode is the subset of InterpretedStringNode that
// have a constant value.
type ConstantInterpretedStringNode interface {
	InterpretedStringNode
	ConstantValue() (string, bool)
}

// ============================== Interpreted String Text ===============================

type InterpretedStringText struct {
	Text     string
	Position *Position
}

var _ InterpretedStringNode = (*InterpretedStringText)(nil)

func (t *InterpretedStringText) Start() Position {
	if t.Position != nil {
		return *t.Position
	}
	return NoPosition
}

func (t *InterpretedStringText) End() Position {
	if t.Position != nil {
		return deltaPos(*t.Position, len(t.Text))
	}
	return NoPosition
}
func (t *InterpretedStringText) Walk(func(Node)) {}

// ConstantValue returns the value the text represents, i.e. it replaces the
// string escape sequences with the characters they represent.
//
// If the text contains any invalid escape sequences, ConstantValue returns
// false.
func (t *InterpretedStringText) ConstantValue() (string, bool) {
	unq, err := strconv.Unquote(`"` + t.Text + `"`)
	if err != nil {
		return "", false
	}
	return unq, true
}

func (*InterpretedStringText) _node()                  {}
func (*InterpretedStringText) _interpretedStringNode() {}

// ============================================================================
// Raw String
// ======================================================================================

// RawString is a Go raw string literal extended to allow interpolation.
type RawString struct {
	Open     *Position
	Contents []RawStringNode
	Close    *Position
}

var _ CodeNode = (*RawString)(nil)

func (s *RawString) Start() Position {
	if s.Open != nil {
		return *s.Open
	}
	for _, n := range s.Contents {
		if n != nil {
			if start := n.Start(); start != NoPosition {
				return start
			}
		}
	}
	if s.Close != nil {
		return *s.Close
	}
	return NoPosition
}

func (s *RawString) End() Position {
	if s.Close != nil {
		return deltaPos(*s.Close, len("\""))
	}
	for _, n := range slices.Backward(s.Contents) {
		if n != nil {
			if end := n.End(); end != NoPosition {
				return end
			}
		}
	}
	if s.Open != nil {
		return deltaPos(*s.Open, len("\""))
	}
	return NoPosition
}

func (s *RawString) Walk(w func(Node)) {
	for _, n := range s.Contents {
		if n != nil {
			w(n)
		}
	}
}

func (s *RawString) Constant() bool {
	for _, n := range s.Contents {
		if cn, _ := n.(ConstantRawStringNode); cn == nil {
			return false
		}
	}
	return true
}

// ConstantValue unquotes the string, returning its constant value.
//
// If the string is not constant, i.e. if it contains any non-constant
// interpolation ConstantValue returns false.
// Furthermore, if the string contains any invalid escape sequences in its text
// nodes, ConstantValue again returns false.
//
// If the string is constant and valid, ConstantValue returns the unquoted
// value and true.
func (s *RawString) ConstantValue() (string, bool) {
	if !s.Constant() {
		return "", false
	}

	var unq strings.Builder
	unq.Grow(int(s.End().Col - s.Start().Col))

	for _, n := range s.Contents {
		val, ok := n.(ConstantRawStringNode).ConstantValue() //nolint:errcheck
		if !ok {
			return "", false
		}
		unq.WriteString(val)
	}
	return unq.String(), true
}

func (*RawString) _node()     {}
func (*RawString) _string()   {}
func (*RawString) _codeNode() {}

// ============================================================================
// Raw String Node
// ======================================================================================

// RawStringNode is either [StringInterpolation] or a pointer to
// [RawStringText].
type RawStringNode interface {
	Node
	_rawStringNode()
}

// if this is changed, change the comment above
var (
	_ RawStringNode = (*RawStringText)(nil)
	_ RawStringNode = StringInterpolation(nil)
)

// ConstantRawStringNode is the subset of RawStringNode that have a constant
// value.
type ConstantRawStringNode interface {
	RawStringNode
	ConstantValue() (string, bool)
}

// ================================== Raw String Text ===================================

type RawStringText struct {
	Text     string
	Position *Position
}

var (
	_ RawStringNode         = (*RawStringText)(nil)
	_ ConstantRawStringNode = (*RawStringText)(nil)
)

func (t *RawStringText) Start() Position {
	if t.Position != nil {
		return *t.Position
	}
	return NoPosition
}

func (t *RawStringText) End() Position {
	if t.Position != nil {
		return deltaPos(*t.Position, len(t.Text))
	}
	return NoPosition
}
func (t *RawStringText) Walk(func(Node)) {}

// ConstantValue returns the value the text represents, stripped of carriage
// returns, per the Go spec's section on [string literals].
//
// It always returns true.
//
// [string literals]: https://go.dev/ref/spec#String_literals
func (t *RawStringText) ConstantValue() (string, bool) {
	return strings.ReplaceAll(t.Text, "\r", ""), true
}

func (*RawStringText) _node()          {}
func (*RawStringText) _rawStringNode() {}

// ============================================================================
// String Interpolation
// ======================================================================================

// StringInterpolation is a pointer to either [BadInterpolation],
// an [CharacterEscape], an [ExpressionInterpolation], a [CharacterReference],
// or a [ComponentCallInterpolation].
type StringInterpolation interface {
	InterpretedStringNode
	RawStringNode
	Interpolation
}

type ConstantStringInterpolation interface {
	StringInterpolation
	ConstantValue() (string, bool)
}
