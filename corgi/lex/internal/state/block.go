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
// It emits a [token.Block] item.
// If the block is named, it emits an Ident with the name of the block.
//
// Lastly, it optionally emits a DotBlock, if the line ends with a '.'.
func Block(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("block")
	l.Emit(token.Block)

	if l.Peek() == '.' {
		return DotBlock
	}

	if l.IsLineEmpty() {
		return lexutil.AssertNewlineOrEOF(l, Next)
	}

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "a block name"})

	if l.Peek() == '.' {
		return DotBlock
	}

	return lexutil.AssertNewlineOrEOF(l, Next)
}

// BlockExpansionBlock consumes a block directive used inside a block expansion.
//
// It assumes the next string is 'block'.
//
// It emits a [token.Block] item.
// If the block is named, it emits an Ident with the name of the block.
func BlockExpansionBlock(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("block")
	l.Emit(token.Block)

	if l.IsLineEmpty() {
		return lexutil.AssertNewlineOrEOF(l, Next)
	}

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	lexutil.EmitIdent(l, &lexerr.UnknownItemError{Expected: "a block name"})

	return lexutil.AssertNewlineOrEOF(l, Next)
}

// Append consumes an append directive.
//
// It assumes the next string is 'append'.
//
// It emits a [token.Append] item followed by a [token.Ident] with the
// name of the block to append.
func Append(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("append")
	l.Emit(token.Append)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	end := lexutil.EmitIdent(l, &lexerr.EOLError{After: "'append'"})
	if end != nil {
		return end
	}

	if l.Peek() == '.' {
		return DotBlock
	}

	return lexutil.AssertNewlineOrEOF(l, Next)
}

// Prepend consumes a prepend directive.
//
// It assumes the next string is 'prepend'.
//
// It emits a [token.Prepend] item followed by a [token.Ident] with the
// name of the block to append.
func Prepend(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("prepend")
	l.Emit(token.Prepend)

	if !l.IgnoreWhile(lexer.IsHorizontalWhitespace) {
		return lexutil.ErrorState(&lexerr.UnknownItemError{Expected: "a space"})
	}

	end := lexutil.EmitIdent(l, &lexerr.EOLError{After: "'prepend'"})
	if end != nil {
		return end
	}

	if l.Peek() == '.' {
		return DotBlock
	}

	return lexutil.AssertNewlineOrEOF(l, Next)
}
