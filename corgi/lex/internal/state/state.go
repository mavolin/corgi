// Package state provides the individual states for the lexer.
package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/lexutil"
	"github.com/mavolin/corgi/corgi/lex/lexerr"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// Next consumes the next directive.
func Next(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if _, ok := l.Context[InImportBlockKey]; ok {
		return NextImportBlockLine(l)
	} else if _, ok = l.Context[InUseBlockKey]; ok {
		return NextUseBlockLine(l)
	} else if _, ok = l.Context[InDotBlockKey]; ok {
		return NextDotBlockLine(l)
	}

	return NextRegular(l)
}

func NextDotBlockLine(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if l.Peek() == lexer.EOF {
		l.Next()
		return lexutil.EOFState()
	}

	dIndent, skippedLines, err := l.ConsumeIndent(lexer.ConsumeNoIncrease)
	if err != nil {
		return lexutil.ErrorState(err)
	}

	lexutil.EmitIndent(l, dIndent)

	if dIndent < 0 {
		delete(l.Context, InDotBlockKey)
		return Next
	}

	for i := 0; i < skippedLines; i++ {
		l.Emit(token.DotBlockLine)
	}

	return DotBlockLine
}

func NextImportBlockLine(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.IgnoreWhile(lexer.IsNewline)

	dIndent, _, err := l.ConsumeIndent(lexer.ConsumeAllIndents)
	if err != nil {
		return lexutil.ErrorState(err)
	}

	if dIndent > 0 { // can't increase indentation
		return lexutil.ErrorState(&lexerr.IllegalIndentationError{In: "an import block"})
	}

	lexutil.EmitIndent(l, dIndent)

	if dIndent < 0 {
		delete(l.Context, InImportBlockKey)
		return Next
	}

	switch {
	case l.Peek() == lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case l.PeekIsWord("//-"):
		return CorgiComment
	case l.Peek() == '"' || l.Peek() == '`':
		return ImportPath
	default:
		return ImportAlias
	}
}

func NextUseBlockLine(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.IgnoreWhile(lexer.IsNewline)

	dIndent, _, err := l.ConsumeIndent(lexer.ConsumeAllIndents)
	if err != nil {
		return lexutil.ErrorState(err)
	}

	if dIndent > 0 { // can't increase indentation
		return lexutil.ErrorState(&lexerr.IllegalIndentationError{In: "a use block"})
	}

	lexutil.EmitIndent(l, dIndent)

	if dIndent < 0 {
		delete(l.Context, InUseBlockKey)
		return Next
	}

	switch {
	case l.Peek() == lexer.EOF:
		l.Next()
		return lexutil.EOFState()
	case l.PeekIsWord("//-"):
		return CorgiComment
	case l.Peek() == '"' || l.Peek() == '`':
		return UsePath
	default:
		return UseAlias
	}
}

func NextRegular(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.IgnoreWhile(lexer.IsNewline)

	if l.Col() == 1 {
		dIndent, _, err := l.ConsumeIndent(lexer.ConsumeAllIndents)
		if err != nil {
			return lexutil.ErrorState(err)
		}

		lexutil.EmitIndent(l, dIndent)
	}

	// some keywords have spaces behind them to avoid confusion if they are
	// just the prefix of an element
	switch {
	case l.Peek() == lexer.EOF:
		return lexutil.EOFState()

	case l.PeekIsString("//-"):
		return CorgiComment

	case l.PeekIsWord("extend"):
		return Extend
	case l.PeekIsWord("import"):
		return Import
	case l.PeekIsWord("use"):
		return Use
	case l.PeekIsWord("func"):
		return Func

	case l.PeekIsWord("-"):
		return Code

	case l.PeekIsWord("include"):
		return Include

	case l.PeekIsWord("block"), l.PeekIsWord("block."), l.PeekIsWord("block:"):
		return Block
	case l.PeekIsWord("append"):
		return Append
	case l.PeekIsWord("prepend"):
		return Prepend

	case l.PeekIsWord("mixin"):
		return Mixin

	case l.PeekIsWord("if block"):
		return IfBlock
	case l.PeekIsWord("if"):
		return If
	case l.PeekIsWord("else if"):
		return ElseIf
	case l.PeekIsWord("else"), l.PeekIsWord("else:"): // inline element
		return Else
	case l.PeekIsWord("switch"):
		return Switch
	case l.PeekIsWord("case"):
		return Case
	case l.PeekIsWord("default"), l.PeekIsWord("default:"): // inline element
		return Default
	case l.PeekIsWord("for"):
		return For
	case l.PeekIsWord("while"):
		return While

	case l.PeekIsWord("."):
		return DotBlock
	case l.PeekIsString("+"):
		return MixinCall
	case l.PeekIsString("="), l.PeekIsString("!="):
		return Assign
	case l.PeekIsString("|"):
		return Pipe
	case l.PeekIsString(":"):
		return Filter

	case l.PeekIsString("//"):
		return Comment
	case l.PeekIsString("."), l.PeekIsString("#"):
		return Div
	case l.PeekIsString("&"):
		return Ampersand
	default:
		return Element
	}
}
