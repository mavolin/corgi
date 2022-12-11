package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/lexutil"
	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// ============================================================================
// Block
// ======================================================================================

// Block consumes a block directive.
//
// It assumes the next string is 'block'.
//
// It emits a [token.Block].
func Block(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("block")
	l.Emit(token.Block)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return Error(&lexerr.UnknownItemError{Expected: "a space"})
	}

	return BlockName
}

// BlockExpansionBlock consumes a block directive used inside a block expansion.
//
// It assumes the next string is 'block'.
//
// It emits a [token.Block].
func BlockExpansionBlock(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("block")
	l.Emit(token.Block)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return Error(&lexerr.UnknownItemError{Expected: "a space"})
	}

	return BlockExpansionBlockName
}

// Append consumes an append directive.
//
// It assumes the next string is 'append'.
//
// It emits a [token.Append].
func Append(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("append")
	l.Emit(token.Append)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return Error(&lexerr.UnknownItemError{Expected: "a space"})
	}

	return BlockName
}

// Prepend consumes a prepend directive.
//
// It assumes the next string is 'prepend'.
//
// It emits a [token.Prepend].
func Prepend(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("prepend")
	l.Emit(token.Prepend)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return Error(&lexerr.UnknownItemError{Expected: "a space"})
	}

	return BlockName
}

// BlockName consumes a block's name.
//
// It assumes the next string is the name of the block.
//
// It emits a [token.Ident] with the name of the block.
func BlockName(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	end := lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "a block name"})
	if end != nil {
		return end
	}

	switch l.Peek() {
	case '.':
		return DotBlock
	case ':':
		return BlockExpansion
	case '=':
		return Assign
	case ' ', '\t':
		if l.IsLineEmpty() {
			return lexutil.AssertNewlineOrEOF(l, Next)
		}

		l.IgnoreNext()
		return Text
	default:
		return lexutil.AssertNewlineOrEOF(l, Next)
	}
}

// BlockExpansionBlockName consumes the name of a block used as part of a block
// expansion.
//
// It assumes the next string is the name of the block.
//
// It emits a [token.Ident] with the name of the block.
func BlockExpansionBlockName(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	end := lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "a block name"})
	if end != nil {
		return end
	}

	switch l.Peek() {
	case ':':
		return BlockExpansion
	case '=':
		return Assign
	case ' ', '\t':
		if l.IsLineEmpty() {
			return lexutil.AssertNewlineOrEOF(l, Next)
		}

		l.IgnoreNext()
		return Text
	default:
		return lexutil.AssertNewlineOrEOF(l, Next)
	}
}
