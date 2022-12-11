package state

import (
	"github.com/mavolin/corgi/corgi/lex/internal/lexer"
	"github.com/mavolin/corgi/corgi/lex/internal/lexutil"
	"github.com/mavolin/corgi/corgi/lex/token"
)

// CorgiComment lexes a corgi comment, ignoring it completely.
//
// It assumes the next string is '//-'.
//
// It emits nothing.
func CorgiComment(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	l.SkipString("//-")
	l.Emit(token.CorgiComment)
	l.IgnoreWhile(lexer.IsHorizontalWhitespace)

	switch l.Next() {
	case lexer.EOF:
		return EOF
	case '\n': // either an empty comment or a block comment
		// handled after the switch
	default: // a one-line comment
		return CorgiCommentText
	}

	// we're possibly in a block comment, check if the next line is indented
	dIndent, _, err := l.ConsumeIndent(lexer.ConsumeSingleIncrease)
	if err != nil {
		return Error(err)
	}

	lexutil.EmitIndent(l, dIndent)

	if l.Peek() == lexer.EOF {
		return EOF
	}

	// it's not, just an empty comment.
	if dIndent <= 0 {
		return Next
	}

	l.Context[InCorgiCommentBlockKey] = true
	return CommentText
}

type inCorgiCommentBlockKey struct{}

var InCorgiCommentBlockKey inCorgiCommentBlockKey

func CorgiCommentText(l *lexer.Lexer[token.Token]) lexer.StateFn[token.Token] {
	if l.Peek() == lexer.EOF {
		return EOF
	}

	l.NextWhile(lexer.MatchesNot('\n'))

	// even emit empty lines
	l.Emit(token.Text)

	return lexutil.AssertNewlineOrEOF(l, Next)
}
