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
	diagnostic2 "github.com/mavolin/corgi/v2/file/diagnostic"
)

const EOF rune = 0

type Preloader func(importPath string)

type Parser struct {
	*file.File
	Preload Preloader

	state *State
}

func New(f *file.File) *Parser {
	return &Parser{
		File:    f,
		Preload: func(string) {},
		state:   newState(),
	}
}

func (p *Parser) next() rune {
	if p.Index() >= len(p.File.Raw) {
		return EOF
	}

	r, size := utf8.DecodeRuneInString(p.File.Raw[p.Index():])
	p.state.advance(size, r == '\n')
	return r
}

func (p *Parser) peek() rune {
	if p.Index() >= len(p.File.Raw) {
		return EOF
	}

	r, _ := utf8.DecodeRuneInString(p.File.Raw[p.Index():])
	return r
}

func (p *Parser) Line() int         { return p.state.line }
func (p *Parser) Col() int          { return p.state.col }
func (p *Parser) Pos() ast.Position { return p.state.Pos() }
func (p *Parser) PosPtr() *ast.Position {
	pos := p.Pos()
	return &pos
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
func (p *Parser) CaptureError(err *diagnostic2.Diagnostic) { p.state.CaptureError(err) }
func (p *Parser) Errors() diagnostic2.List                 { return p.state.Errors() }
func (p *Parser) CaptureComment(g *ast.CommentGroup)       { p.state.CaptureComment(g) }
func (p *Parser) CloneState() *State                       { return p.state.Clone() }

func (p *Parser) RestoreState(s *State) {
	p.state = s
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
	Func[T any] func(p *Parser) (T, *diagnostic2.Diagnostic)

	// A WhitespaceFunc is a special [Func] that parses whitespace.
	// It semantically differs, in that consumed whitespace is rolled back, if
	// the next call to [TryErr] or [Must] (and its derivatives) fails.
	WhitespaceFunc func(p *Parser) *diagnostic2.Diagnostic
)

// Matches reports whether f would match.
// It does not consume any input.
func Matches[T any](p *Parser, f Func[T]) bool {
	state := p.CloneState()
	p.state.ws = nil
	_, err := f(p)
	p.RestoreState(state)
	return err == nil
}

func MatchesWS(p *Parser, f WhitespaceFunc) bool {
	state := p.CloneState()
	p.state.ws = nil
	err := f(p)
	p.RestoreState(state)
	return err == nil
}

func MatchesToken(p *Parser, s string) bool {
	start := p.Index()
	end := start + len(s)
	return end <= len(p.File.Raw) && p.Raw[start:end] == s
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

func TryErr[T any](p *Parser, f Func[T]) (T, *diagnostic2.Diagnostic) {
	restore := p.state.takeWSStart()
	v, err := f(p)
	if err != nil {
		p.RestoreState(restore)
		return v, err
	}
	return v, err
}

func Try[T any](p *Parser, f Func[T]) T {
	v, _ := TryErr(p, f)
	return v
}

func TryOptionalErr[T any](p *Parser, f Func[T], ws WhitespaceFunc) (T, *diagnostic2.Diagnostic) {
	state := p.CloneState()
	p.state.ws = nil
	v, err := f(p)
	if err != nil {
		p.RestoreState(state)
		return v, err
	}
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
	state := p.CloneState()
	p.state.ws = nil
	for _, f := range fs {
		v, err := f(p)
		if err == nil { // IS nil
			return v
		}
		p.RestoreState(state)
	}
	if state.ws != nil {
		p.RestoreState(state.ws)
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
	return TrySkipErr(p, f) == nil
}

func TrySkipErr(p *Parser, f WhitespaceFunc) *diagnostic2.Diagnostic {
	state := p.CloneState()
	if !p.state.parsingWS {
		p.state.markWSStart()
		p.state.parsingWS = true
		defer func() { p.state.parsingWS = false }()
	}

	if err := f(p); err != nil {
		p.RestoreState(state)
		return err
	}
	return nil
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

// MustSkip is the [Must] equivalent of [TrySkip].
func MustSkip(p *Parser, f WhitespaceFunc) {
	if err := TrySkipErr(p, f); err != nil {
		p.CaptureError(err)
	}
}

func CommitWS(p *Parser) {
	p.state.ws = nil
}

// RestoreWS restores all whitespace consumed by the last calls to [TrySkip]
// and friends.
func RestoreWS(p *Parser) {
	if p.state.ws != nil {
		p.RestoreState(p.state.ws)
	}
}
