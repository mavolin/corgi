// Package parse implements a parser for corgi files.
package parse

import (
	"github.com/pkg/errors"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/lex"
	"github.com/mavolin/corgi/corgi/lex/token"
	"github.com/mavolin/corgi/internal/stack"
)

type Parser struct {
	lex  *lex.Lexer
	mode Mode

	peekedItem *lex.Item
	eof        *lex.Item

	f file.File

	context stack.Stack[Context]
}

type Mode uint8

const (
	// ModeMain represents the parsing of a main file.
	// A main file must define an output function and may extend other files.
	//
	// If the file extends another file, it may not define a doctype.
	ModeMain Mode = iota + 1
	// ModeExtend represents the parsing of an extended file.
	// Extended templates must not define an output function.
	// They may also extend other templates.
	//
	// If the file extends another file, it may not define a doctype.
	ModeExtend
	// ModeInclude represents the parsing of an included corgi file.
	// Included files may define an output function, which is ignored.
	// They must not extend other files.
	//
	// Included files also don't have global code, but only regular code.
	ModeInclude
	// ModeUse represents the parsing of a file that was imported through a use
	// directive.
	// Use files may only import packages, use other directories or files,
	// define global code, and define mixins.
	ModeUse
)

// Context is the context in which the parser is parsing the current scope.
// It limits the expected items.
//
// This is only relevant to outside callers, when parsing an included file.
type Context uint8

const (
	// ContextRegular is the context used when no other context applies.
	ContextRegular Context = iota
	// ContextMixinDefinition is the Context used when in the body of a mixin
	// definition.
	//
	// In it, token.Block and token.IfBlock items are allowed.
	ContextMixinDefinition
	// ContextMixinCall is the Context used when in the body of a mixin call.
	//
	// In it only token.If, token.Switch, token.And, token.MixinCall, and token.Block are
	// allowed.
	//
	// If inside a token.Block inside a mixin call, ContextRegular is to be used.
	ContextMixinCall
	// ContextMixinCallConditional is the same as ContextMixinCall, but only
	// specifies that we're in a conditional (token.If, or token.Switch) where
	// token.Block statements are not allowed.
	ContextMixinCallConditional
)

// New creates a new parser that parses the given input in the given
// Mode.
// Name is the name of the file that is being parsed.
func New(mode Mode, context Context, source, name, in string) *Parser {
	p := &Parser{
		lex:  lex.New(in),
		mode: mode,
		f:    file.File{Name: name, Source: source},

		context: stack.New[Context](20),
	}

	p.context.Push(context)

	return p
}

type stateFn func() (stateFn, error)

// Parse starts parsing the input.
// It returns the parsed file.File or an error.
func (p *Parser) Parse() (f *file.File, err error) {
	p.lex.Lex()

	state := p.start

	for state != nil {
		state, err = state()
		if err != nil {
			return nil, err
		}
	}

	return &p.f, nil
}

// ============================================================================
// Items
// ======================================================================================

func (p *Parser) next() (itm lex.Item) {
	if p.eof != nil {
		return *p.eof
	}

	if p.peekedItem != nil {
		itm = *p.peekedItem
		p.peekedItem = nil

		if itm.Type == token.EOF {
			p.eof = &itm
		}

		return itm
	}

	next := p.lex.NextItem()
	if next.Type == token.EOF {
		p.eof = &next
	}

	return next
}

func (p *Parser) peek() lex.Item {
	if p.eof != nil {
		return *p.eof
	}

	if p.peekedItem != nil {
		return *p.peekedItem
	}

	itm := p.lex.NextItem()
	p.peekedItem = &itm
	return itm
}

type Error struct {
	Line int
	Col  int
	Err  error
}

// error is a helper that returns an error containing the current source, file,
// line, and column.
func (p *Parser) error(itm lex.Item, err error) error {
	if err == nil {
		return nil
	}

	return errors.Wrapf(err, "%s/%s:%d:%d", p.f.Source, p.f.Name, itm.Line, itm.Col)
}

// unexpectedItem is short for:
//
//	p.error(itm, &UnexpectedItemError{Found: found.Type, Expected: expected})
func (p *Parser) unexpectedItem(found lex.Item, expected ...token.Token) error {
	if found.Err != nil {
		return p.error(found, found.Err)
	}

	return p.error(found, &UnexpectedItemError{Found: found.Type, Expected: expected})
}
