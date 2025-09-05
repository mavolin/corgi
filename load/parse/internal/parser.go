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
// the next call to Try* in the same function fails.
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
// To remedy, either use TryInOrder or TryOptional*.
package parser

import (
	"math"
	"reflect"
	"slices"
	"unicode/utf8"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

const EOF rune = 0

type Preloader func(importPath string)

type Parser struct {
	*file.File
	Preload Preloader

	state    *State
	errs     diagnostic.List
	comments []*ast.CommentGroup

	statePool pool[State]

	inline    bool
	parsingWS bool

	runes []rune
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
		errs:      make(diagnostic.List, 0, 48),
		comments:  make([]*ast.CommentGroup, 0, 128),
		statePool: make(pool[State], 0, 32),
		runes:     []rune(f.AST.Raw),
	}
}

func (p *Parser) next() rune {
	if int(p.state.runeIndex) >= len(p.runes) {
		return EOF
	}

	r := p.runes[p.state.runeIndex]
	if r == utf8.RuneError {
		p.state.advance(1, r == '\n')
		p.CaptureError(&diagnostic.Diagnostic{
			Message: "invalid UTF-8 encoding",
			Primary: []diagnostic.Annotation{
				anno.Position(p.File, p.Pos(), "this byte is not valid UTF-8"),
			},
		})
		return r
	}
	p.state.advance(uint32(utf8.RuneLen(r)), r == '\n') //nolint:gosec
	return r
}

func (p *Parser) peek() rune {
	if int(p.state.runeIndex) >= len(p.runes) {
		return EOF
	}

	return p.runes[p.state.runeIndex]
}

func (p *Parser) skipString(s string) {
	for range s {
		p.next()
	}
}

func (p *Parser) Line() uint16      { return p.state.line }
func (p *Parser) Col() uint16       { return p.state.col }
func (p *Parser) Pos() ast.Position { return p.state.Pos() }
func (p *Parser) Index() int        { return int(p.state.index) }
func (p *Parser) Inline() bool      { return p.inline }

func (p *Parser) NumErrors() uint8              { return p.state.numErrs }
func (p *Parser) Errors() diagnostic.List       { return slices.Clip(p.errs) }
func (p *Parser) Comments() []*ast.CommentGroup { return p.comments }

func (p *Parser) PosPtr() *ast.Position {
	pos := p.Pos()
	return &pos
}

func (p *Parser) DoInline(f func()) {
	if p.inline {
		f()
		return
	}

	p.inline = true
	f()
	p.inline = false
}

func (p *Parser) CaptureError(err *diagnostic.Diagnostic) {
	if len(p.errs) < math.MaxUint8 {
		p.errs = append(p.errs, err)
		p.state.numErrs = uint8(len(p.errs)) //nolint:gosec
	}
}

func (p *Parser) CaptureComment(g *ast.CommentGroup) {
	if len(p.comments) < math.MaxUint16 {
		p.comments = append(p.comments, g)
		p.state.numComments = uint16(len(p.comments)) //nolint:gosec
	}
}

func (p *Parser) cloneState() *State {
	s := p.statePool.Get()
	p.state.Copy(s)
	return s
}

func (p *Parser) CloneState() *State {
	if p.state.ws != nil {
		panic("CloneState called after WS was consumed")
	}
	s := p.statePool.Get()
	p.state.Copy(s)
	return s
}

func (p *Parser) RestoreState(s *State) {
	p.statePool.Put(p.state)
	p.state = s
	p.errs = p.errs[:p.state.numErrs]
	p.comments = p.comments[:p.state.numComments]
}

// commitWS commits the whitespace, preventing rollback.
func (p *Parser) commitWS() {
	if !p.parsingWS {
		p.state.ws = nil
	}
}

func (p *Parser) markWSStart(start *State) {
	if p.state.ws != nil {
		return
	}
	p.state.ws = start
}

func (p *Parser) takeWSStart() *State {
	if p.state.ws == nil || p.parsingWS {
		return p.cloneState()
	}
	wsStart := p.state.ws
	p.state.ws = nil
	return wsStart
}

type (
	// Func represents a sub-parser that can be tried to see if it matches.
	//
	// While the implementation is up to the function itself, a typical
	// indicator of whether the function matches is the presence of a unique
	// prefix, such as `comp` for a component declaration.
	//
	// If the func does not match, it should return the zero value.
	//
	// If the func matches, but the parsed value contains syntactical errors,
	// those should be captured using the `CaptureError` method of the parser.
	//
	// Funcs must not be called directly, but only using Try*.
	Func[T any] func(p *Parser) T

	// A WhitespaceFunc is a special [Func] that parses whitespace.
	// It semantically differs, in that consumed whitespace is rolled back, if
	// the next call to Try* (except for TryOptional*) fails.
	WhitespaceFunc func(p *Parser) bool
)

// Matches reports whether f would match.
// It does not consume any input.
func Matches[T any](p *Parser, f Func[T]) bool {
	restore := p.cloneState()
	CommitWS(p)
	v := f(p)
	p.RestoreState(restore)
	return !isZero(v)
}

func MatchesWS(p *Parser, f WhitespaceFunc) bool {
	restore := p.cloneState()
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

func MatchesAnyToken(p *Parser, ss ...string) bool {
	for _, s := range ss {
		if MatchesToken(p, s) {
			return true
		}
	}
	return false
}

func MatchesAnyRune(p *Parser, rs ...rune) bool {
	return MatchesRunePredicate(p, func(cmp rune) bool {
		return slices.Contains(rs, cmp)
	})
}

func MatchesRunePredicate(p *Parser, pred func(rune) bool) bool {
	return pred(p.peek())
}

func Try[T any](p *Parser, f Func[T]) T {
	var zero T

	restore := p.takeWSStart()
	v := f(p)
	if isZero(v) {
		p.RestoreState(restore)
		return zero
	}
	p.statePool.Put(restore)
	return v
}

func TryOptional[T any](p *Parser, f Func[T], ws WhitespaceFunc) T {
	var zero T

	restore := p.cloneState()
	CommitWS(p)
	v := f(p)
	if isZero(v) {
		p.RestoreState(restore)
		return zero
	}
	p.statePool.Put(restore)
	if ws != nil {
		TrySkip(p, ws)
	}
	return v
}

// TryInOrder tries all Funcs until it finds one that matches.
//
// If none match, it returns false.
func TryInOrder[T any](p *Parser, fs ...Func[T]) T {
	var zero T

	restore := p.takeWSStart()
	for _, f := range fs {
		v := Try(p, f)
		if !isZero(v) {
			p.statePool.Put(restore)
			return v
		}
	}
	p.RestoreState(restore)

	return zero
}

// TrySkip attempts to skip whitespace using the given [WhitespaceFunc].
//
// Calls to TrySkip can be stacked, so that the next call to Try* rolls back to
// the first TrySkip call in a chain of many.
//
// Even if TrySkip fails to match, it does not affect a previous restore
// point.
func TrySkip(p *Parser, f WhitespaceFunc) bool {
	restore := p.cloneState()
	if !p.parsingWS {
		p.markWSStart(restore)
		p.parsingWS = true

		matches := f(p)
		if !matches {
			p.RestoreState(restore)
		}

		p.parsingWS = false
		return matches
	}

	if !f(p) {
		p.RestoreState(restore)
		return false
	}

	p.statePool.Put(restore)
	return true
}

func CommitWS(p *Parser) {
	p.commitWS()
}

// RestoreWS restores all whitespace consumed by the last calls to [TrySkip]
// and friends.
func RestoreWS(p *Parser) {
	if p.state.ws != nil && !p.parsingWS {
		p.RestoreState(p.state.ws)
	}
}

func Collect[T any](p *Parser, f Func[T], ws WhitespaceFunc) []T {
	var ts []T
	for {
		if ws != nil {
			TrySkip(p, ws)
		}
		v := Try(p, f)
		if isZero(v) {
			break
		}
		ts = append(ts, v)
	}
	if len(ts) == 0 {
		return nil
	}
	return slices.Clip(ts)
}

// isZero reports whether the value t is the zero value of its type.
//
// We need this function, because Go has no native way to check if a generic
// type is the zero value.
// We could limit the type param to only cover comparable types, but that would
// exclude slices, which we use a lot.
func isZero[T any](t T) bool {
	return reflect.ValueOf(&t).Elem().IsZero()
}
