// Package parser implements a parser for the corgi language.
//
// The parser is basically parser combinator, although each function still
// operates on the same shared Parser instance to facilitate positional
// tracking, error recovery and reporting, and other stateful operations.
// This is, of course, different from a typical parser combinator that takes in
// a string, but our approach eases implementation tremendously with no
// drawback.
//
// We use special functions for consuming whitespace, which is rolled back, if
// the next call to TryErr or Must in the same function fails.
// This is incredibly convenient, but can be tricky if you skip WS and then
// try a bunch of functions.
// For example, consider the buggy code below:
//
//	 parser.MustSkip(p, whitespace.Any())
//	 if res1, ok := parser.Try(p, a()); ok {
//			// ...
//	 } else if res2, ok := parser.Try(p, b()); ok {
//			// ...
//	 }
//
// If a fails, the consumed whitespace is rolled back and b is tried with
// whitespace in front.
// To remedy, either use TryInOrder, TryOptional*, or, if a and b return
// different types, wrap them in a Func like this:
//
//	parser.MustSkip(p, whitespace.Any())
//	v, ok := parser.TryErr(p, func(p *parser.Parser) (parentType, *diagnostic.Diagnostic) {
//		if res1, ok := parser.Try(p, a()); ok {
//			return parentType(res1), nil
//		}
//		return parser.TryErr(p, b())
//	})
package parser

import (
	"unicode/utf8"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
)

const EOF rune = 0

type Preloader func(importPath string)

type Parser struct {
	*file.File
	Preload Preloader

	state     *State
	statePool pool[State]
	posPool   pool[ast.Position]
}

type pool[T any] []*T

func (p *pool[T]) Get() *T {
	if len(*p) == 0 {
		return new(T)
	}

	s := (*p)[len(*p)-1]
	*p = (*p)[:len(*p)-1]
	return s
}

func (p *pool[T]) Put(s *T) {
	*p = append(*p, s)
}

func New(f *file.File) *Parser {
	return &Parser{
		File:      f,
		Preload:   func(string) {},
		state:     newState(),
		statePool: make(pool[State], 0, 32),
		posPool:   make(pool[ast.Position], 0, 32),
	}
}

func (p *Parser) next() rune {
	if p.Index() >= len(p.AST.Raw) {
		return EOF
	}

	r, size := utf8.DecodeRuneInString(p.AST.Raw[p.Index():])
	p.state.advance(size, r == '\n')
	return r
}

func (p *Parser) peek() rune {
	if p.Index() >= len(p.AST.Raw) {
		return EOF
	}

	r, _ := utf8.DecodeRuneInString(p.AST.Raw[p.Index():])
	return r
}

func (p *Parser) skipString(s string) {
	for range s {
		p.next()
	}
}

func (p *Parser) Line() int         { return p.state.line }
func (p *Parser) Col() int          { return p.state.col }
func (p *Parser) Pos() ast.Position { return p.state.Pos() }
func (p *Parser) PosPtr() *ast.Position {
	pos := p.posPool.Get()
	pos.Line, pos.Col = p.state.line, p.state.col
	return pos
}
func (p *Parser) Index() int { return p.state.Index() }
func (p *Parser) Inline() bool {
	return p.state.inline
}

func (p *Parser) DoInline(f func()) {
	if p.state.inline {
		f()
		return
	}

	p.state.inline = true
	f()
	p.state.inline = false
}
func (p *Parser) CaptureError(err *diagnostic.Diagnostic) { p.state.CaptureError(err) }
func (p *Parser) Errors() diagnostic.List                 { return p.state.Errors() }
func (p *Parser) CaptureComment(g *ast.CommentGroup)      { p.state.CaptureComment(g) }

func (p *Parser) CloneState() *State {
	return p.state.Clone(&p.statePool)
}

func (p *Parser) RestoreState(s *State) {
	p.statePool.Put(p.state)
	p.state = s
}

func (p *Parser) markWSStart() {
	p.state.markWSStart(&p.statePool)
}

func (p *Parser) takeWSStart() *State {
	return p.state.takeWSStart(&p.statePool)
}

type (
	// Func represents a sub-parser that can be tried to see if it matches.
	//
	// While the implementation is up to the function itself, a typical
	// indicator of whether the function matches is the present of a unique
	// prefix, such as `comp` for a component declaration.
	//
	// If the func does not match, it should return an error.
	//
	// If the func matches, but the parsed value contains syntactical errors,
	// those should be captured using the `CaptureError` method of the parser.
	//
	// Funcs must not be called directly, but only using [TryErr] and [Must].
	Func[T any] func(p *Parser) (T, *diagnostic.Diagnostic)

	// A WhitespaceFunc is a special [Func] that parses whitespace.
	// It semantically differs, in that consumed whitespace is rolled back, if
	// the next call to [TryErr] or [Must] (and its derivatives) fails.
	WhitespaceFunc func(p *Parser) bool
)

// Matches reports whether f would match.
// It does not consume any input.
func Matches[T any](p *Parser, f Func[T]) bool {
	restore := p.CloneState()
	CommitWS(p)
	_, err := f(p)
	p.RestoreState(restore)
	return err == nil
}

func MatchesWS(p *Parser, f WhitespaceFunc) bool {
	restore := p.CloneState()
	CommitWS(p)
	ok := f(p)
	p.RestoreState(restore)
	return ok
}

func MatchesToken(p *Parser, s string) bool {
	start := p.Index()
	end := start + len(s)
	return end <= len(p.AST.Raw) && p.AST.Raw[start:end] == s
}

func MatchesAnyRune(p *Parser, rs ...rune) bool {
	return MatchesRunePredicate(p, func(cmp rune) bool {
		for _, r := range rs {
			if r == cmp {
				return true
			}
		}
		return false
	})
}

func MatchesRunePredicate(p *Parser, pred func(rune) bool) bool {
	return pred(p.peek())
}

func TryErr[T any](p *Parser, f Func[T]) (T, *diagnostic.Diagnostic) {
	restore := p.takeWSStart()
	v, err := f(p)
	if err != nil {
		p.RestoreState(restore)
		return v, err
	}
	p.statePool.Put(restore)
	return v, err
}

func Try[T any](p *Parser, f Func[T]) T {
	v, _ := TryErr(p, f)
	return v
}

func TryOptionalErr[T any](p *Parser, f Func[T], ws WhitespaceFunc) (T, *diagnostic.Diagnostic) {
	restore := p.CloneState()
	CommitWS(p)
	v, err := f(p)
	if err != nil {
		p.RestoreState(restore)
		return v, err
	}
	p.statePool.Put(restore)
	if ws != nil {
		TrySkip(p, ws)
	}
	return v, nil
}

func TryOptional[T any](p *Parser, f Func[T], ws WhitespaceFunc) T {
	v, _ := TryOptionalErr(p, f, ws)
	return v
}

// TryInOrder tries all Funcs until it finds one that matches.
//
// If none match, it returns false.
func TryInOrder[T any](p *Parser, fs ...Func[T]) T {
	restore := p.takeWSStart()
	for _, f := range fs {
		v, err := TryErr(p, f)
		if err == nil { // IS nil
			p.statePool.Put(restore)
			return v
		}
	}
	if restore.ws != nil {
		p.RestoreState(restore.ws)
	}

	var z T
	return z
}

// TrySkip attempts to skip whitespace using the given [WhitespaceFunc].
//
// Calls to TrySkip can be stacked, so that the next call to [TryErr] or [Must]
// (and friends) rolls back to the first TrySkip call in a chain of many.
//
// Even if TrySkip fails to match, it does not affect a previous restore
// point.
func TrySkip(p *Parser, f WhitespaceFunc) bool {
	restore := p.CloneState()
	if !p.state.parsingWS {
		p.markWSStart()
		p.state.parsingWS = true
		defer func() { p.state.parsingWS = false }()
	}

	if !f(p) {
		p.RestoreState(restore)
		return false
	}
	p.statePool.Put(restore)
	return true
}

// Must tries to parse using the given [Func].
// If the func returns an error, Must captures it and returns the value
// returned by Func, most commonly the zero value.
func Must[T any](p *Parser, f Func[T]) T {
	v, err := TryErr(p, f)
	if err != nil {
		p.CaptureError(err)
	}
	return v
}

func CommitWS(p *Parser) {
	p.state.commitWS()
}

// RestoreWS restores all whitespace consumed by the last calls to [TrySkip]
// and friends.
func RestoreWS(p *Parser) {
	if p.state.ws != nil && !p.state.parsingWS {
		p.RestoreState(p.state.ws)
	}
}
