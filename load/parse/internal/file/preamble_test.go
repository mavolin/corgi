package file

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestPackageDirective(t *testing.T) {
	t.Parallel()

	in := "package foo"
	want := &ast.PackageDirective{
		Package: &ast.Position{Line: 1, Col: 1},
		Name:    &ast.Identifier{Name: "foo", Position: &ast.Position{Line: 1, Col: 9}},
	}

	got := parsetest.ParsesExact(t, in, PackageDirective())
	should.Equal(t, got, want)
}

func TestImport(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.Import
	}{
		{
			name: "single",
			in:   "import \"foo\"",
			want: &ast.Import{
				Import: &ast.Position{Line: 1, Col: 1},
				Specs: []*ast.ImportSpec{
					{
						Path: &ast.StaticString{
							Open:     &ast.Position{Line: 1, Col: 8},
							Quote:    '"',
							Contents: "foo",
							Close:    &ast.Position{Line: 1, Col: 12},
						},
					},
				},
			},
		}, {
			name: "multiple",
			in: "import (\n" +
				"\t\"foo\"\n" +
				"\t\"bar\"\n" +
				")",
			want: &ast.Import{
				Import: &ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 8},
				Specs: []*ast.ImportSpec{
					{
						Path: &ast.StaticString{
							Open:     &ast.Position{Line: 2, Col: 2},
							Quote:    '"',
							Contents: "foo",
							Close:    &ast.Position{Line: 2, Col: 6},
						},
					}, {
						Path: &ast.StaticString{
							Open:     &ast.Position{Line: 3, Col: 2},
							Quote:    '"',
							Contents: "bar",
							Close:    &ast.Position{Line: 3, Col: 6},
						},
					},
				},
				RParen: &ast.Position{Line: 4, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesExact(t, c.in, Import())
			should.Equal(t, got, c.want)
		})
	}
}

func TestImportSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.ImportSpec
	}{
		{
			name: "no alias",
			in:   "\"foo\"",
			want: &ast.ImportSpec{
				Path: &ast.StaticString{
					Open:     &ast.Position{Line: 1, Col: 1},
					Quote:    '"',
					Contents: "foo",
					Close:    &ast.Position{Line: 1, Col: 5},
				},
			},
		}, {
			name: "alias",
			in:   "foo \"bar\"",
			want: &ast.ImportSpec{
				Alias: &ast.Identifier{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
				Path: &ast.StaticString{
					Open:     &ast.Position{Line: 1, Col: 5},
					Quote:    '"',
					Contents: "bar",
					Close:    &ast.Position{Line: 1, Col: 9},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesExact(t, c.in, ImportSpec())
			should.Equal(t, got, c.want)
		})
	}
}
