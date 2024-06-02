package parser

import (
	"unicode/utf8"

	"github.com/mavolin/corgi/fancyerr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
)

const EOF rune = 0

type Parser struct {
	*file.File

	state *State
}

func New(f *file.File) *Parser {
	return &Parser{
		File:  f,
		state: newState(),
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
func (p *Parser) Index() int { return p.state.index }
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
func (p *Parser) CaptureError(err *fancyerr.Error)   { p.state.CaptureError(err) }
func (p *Parser) CaptureComment(g *ast.CommentGroup) { p.state.CaptureComment(g) }
func (p *Parser) CloneState() *State                 { return p.state.Clone() }

func (p *Parser) RestoreState(s *State) {
	p.state = s
}

// Func represents a sub-parser that can be tried to see if it matches.
//
// TokenWhile the implementation is up to the function itself, a typical
// indicator of whether the function matches is the present of a unique
// prefix, such as `comp` for a component declaration.
//
// If the func does not match, it should return an error.
//
// If the func matches, but the parsed value contains syntactical errors,
// those should be captured using the `CaptureError` method of the parser.
type Func[T any] func(p *Parser) (T, *fancyerr.Error)

// Matches reports whether f would match.
// It does not consume any input.
func Matches[T any](p *Parser, f Func[T]) bool {
	state := p.CloneState()
	_, err := f(p)
	p.RestoreState(state)
	return err == nil
}

func MatchesToken(p *Parser, s string) bool {
	start := p.Index()
	end := start + len(s)
	return end <= len(p.File.Raw) && p.Raw[start:end] == s
}

func MatchesAnyRune(p *Parser, rs ...rune) bool {
	peek := p.peek()
	for _, r := range rs {
		if peek == r {
			return true
		}
	}
	return false
}

func MatchesAnyRunePredicate(p *Parser, preds ...func(rune) bool) bool {
	r := p.peek()
	for _, pred := range preds {
		if pred(r) {
			return true
		}
	}
	return false
}

// Try tries to parse using the given [Func], ignoring an error if one occurs.
func Try[T any](p *Parser, f Func[T]) (_ T, ok bool) {
	state := p.CloneState()
	v, err := f(p)
	if err != nil {
		p.RestoreState(state)
		return v, false
	}
	return v, true
}

// TryInOrder tries all Funcs until it finds one that matches.
//
// If none match, it returns false.
func TryInOrder[T any](p *Parser, fs ...Func[T]) (T, bool) {
	state := p.CloneState()
	for _, f := range fs {
		v, err := f(p)
		if err == nil { // IS nil
			return v, true
		}
	}
	p.RestoreState(state)

	var z T
	return z, false
}

// Must tries to parse using the given [Func] and try set to false.
func Must[T any](p *Parser, f Func[T]) T {
	state := p.CloneState()
	v, err := f(p)
	if err != nil {
		p.RestoreState(state)
		p.CaptureError(err)
	}
	return v
}
