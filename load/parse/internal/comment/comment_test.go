package comment

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestComment(t *testing.T) {
	t.Parallel()

	t.Run("line", func(t *testing.T) {
		t.Parallel()
		testLineComment(t, Comment())
	})

	t.Run("general", func(t *testing.T) {
		t.Parallel()
		testGeneralComment(t, Comment())
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

	got := parsetest.ParsesExact(t, "// foo\n", f)
	should.Equal(t, got, want)
}

func TestGeneralComment(t *testing.T) {
	t.Parallel()
	testGeneralComment(t, GeneralComment())
}

func testGeneralComment(t *testing.T, f parser.Func[*ast.Comment]) {
	t.Run("single line", func(t *testing.T) {
		t.Parallel()

		want := &ast.Comment{
			Open:    &ast.Position{Line: 1, Col: 1},
			Comment: " foo ",
			General: true,
			Close:   &ast.Position{Line: 1, Col: 8},
			Until:   ast.Position{Line: 1, Col: 10},
		}

		got := parsetest.ParsesExact(t, "/* foo */", f)
		should.Equal(t, got, want)
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

		got := parsetest.ParsesExact(t, "/* foo\n   bar */", f)
		should.Equal(t, got, want)
	})

	t.Run("missing closing", func(t *testing.T) {
		t.Parallel()

		in := "/* foo\n"
		wantError := "unclosed general comment"

		parsetest.ParsesUntilExtra(t, in, "", f,
			parsetest.Inline(), parsetest.WantErrors(wantError))
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

			got := parsetest.ParsesExact(t, "/* foo */", f, parsetest.Inline())
			should.Equal(t, got, want)
		})

		t.Run("error on multi line", func(t *testing.T) {
			t.Parallel()

			in := "/* foo\n   bar */"
			wantError := "illegal placement of multiline general comment"

			parsetest.ParsesUntilExtra(t, in, "", f,
				parsetest.Inline(), parsetest.WantErrors(wantError))
		})

		t.Run("missing closing", func(t *testing.T) {
			t.Parallel()

			in := "/* foo\n"
			wantError := "unclosed general comment"

			parsetest.ParsesUntilExtra(t, in, "", f,
				parsetest.Inline(), parsetest.WantErrors(wantError))
		})
	})
}
