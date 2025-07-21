package comment

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestComment(t *testing.T) {
	t.Parallel()

	t.Run("line", func(t *testing.T) {
		t.Parallel()
		testLineComment(t, Comment())
	})

	t.Run("block", func(t *testing.T) {
		t.Parallel()
		testBlockComment(t, Comment())
	})
}

func TestLineComment(t *testing.T) {
	t.Parallel()
	testLineComment(t, LineComment())
}

func testLineComment(t *testing.T, f parser.Func[*ast.Comment]) {
	want := &ast.Comment{
		Open:    &ast.Position{Line: 1, Col: 1},
		Comment: " foo",
		General: false,
		Until:   ast.Position{Line: 1, Col: 7},
	}

	got := parsetest.ParsesFully(t, "// foo\n", f)
	should.Equal(t, want, got)
}

func TestGeneralComment(t *testing.T) {
	t.Parallel()
	testBlockComment(t, GeneralComment())
}

func testBlockComment(t *testing.T, f parser.Func[*ast.Comment]) {
	t.Run("single line", func(t *testing.T) {
		t.Parallel()

		want := &ast.Comment{
			Open:    &ast.Position{Line: 1, Col: 1},
			Comment: " foo ",
			General: true,
			Close:   &ast.Position{Line: 1, Col: 8},
			Until:   ast.Position{Line: 1, Col: 10},
		}

		got := parsetest.ParsesFully(t, "/* foo */", f)
		should.Equal(t, want, got)
	})

	t.Run("multi line", func(t *testing.T) {
		t.Parallel()

		want := &ast.Comment{
			Open:    &ast.Position{Line: 1, Col: 1},
			Comment: " foo\n   bar ",
			General: true,
			Close:   &ast.Position{Line: 2, Col: 8},
			Until:   ast.Position{Line: 2, Col: 10},
		}

		got := parsetest.ParsesFully(t, "/* foo\n   bar */", f)
		should.Equal(t, want, got)
	})

	t.Run("missing closing", func(t *testing.T) {
		t.Parallel()

		p := parsetest.NewParser(t, "/* foo")
		p.DoInline(func() {
			parsetest.AssertMatchesButError(t, p, f)
		})
	})

	t.Run("inline", func(t *testing.T) {
		t.Parallel()

		t.Run("single line", func(t *testing.T) {
			t.Parallel()

			want := &ast.Comment{
				Open:    &ast.Position{Line: 1, Col: 1},
				Comment: " foo ",
				General: true,
				Close:   &ast.Position{Line: 1, Col: 8},
				Until:   ast.Position{Line: 1, Col: 10},
			}

			p := parsetest.NewParser(t, "/* foo */")
			p.DoInline(func() {
				got := parsetest.AssertNoError(t, p, f)
				parsetest.AssertEOF(t, p)
				should.Equal(t, want, got)
			})
		})

		t.Run("error on multi line", func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, "/* foo\n   bar */")
			p.DoInline(func() {
				parsetest.AssertMatchesButError(t, p, f)
			})
		})

		t.Run("missing closing", func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, "/* foo\n*/")
			p.DoInline(func() {
				parsetest.AssertMatchesButError(t, p, f)
			})
		})
	})
}
