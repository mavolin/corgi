package parsetest

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

func SkipsWhitespaceUntilIdentifier(t *testing.T, in string, f parser.WhitespaceFunc) []*ast.CommentGroup {
	t.Helper()
	return SkipsWhitespaceUntil(t, in, "abc", f)
}

func SkipsWhitespace(t *testing.T, in string, f parser.WhitespaceFunc) []*ast.CommentGroup {
	t.Helper()
	return SkipsWhitespaceUntil(t, in, "", f)
}

func SkipsWhitespaceUntil(t *testing.T, in, extra string, f parser.WhitespaceFunc) []*ast.CommentGroup {
	t.Helper()

	p := NewParser(t, in+extra)
	should.True(t, parser.TrySkip(p, f)) // match error

	if !should.True(t, len(p.Errors()) == 0) { // errors
		t.Log(p.Errors().Pretty(diagnostic.PrettyOptions{}))
	}

	shouldBeAtEnd(t, p, in)
	return p.Comments()
}
