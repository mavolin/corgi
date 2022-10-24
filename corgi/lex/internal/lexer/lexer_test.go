package lexer

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mavolin/corgi/corgi/lex/token"
)

func TestLexer_NextItem(t *testing.T) {
	t.Parallel()

	in := "abc"
	expect := Item[token.Token]{
		Type: token.Code,
		Val:  "abc",
		Line: 1,
		Col:  1,
	}

	l := New[token.Token](in, nil)

	for i := 0; i < len(in); i++ {
		l.Next()
	}

	go l.Emit(token.Code)

	actual := l.NextItem()
	assert.Equal(t, expect, actual)
}
