package file

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestPackageDirective(t *testing.T) {
	t.Parallel()

	in := "package foo"
	want := &ast.PackageDirective{
		Package: &ast.Position{Line: 1, Col: 1},
		Name: &ast.Identifier{
			Name: "foo",
			Position: &ast.Position{
				Line: 1,
				Col:  ast.Col(1 + len("package ")),
			},
		},
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
						Path: &ast.String{
							Open:  &ast.Position{Line: 1, Col: ast.Col(1 + len("import "))},
							Quote: '"',
							Contents: []ast.StringNode{
								&ast.StringText{
									Text:     "foo",
									Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`import "`))},
								},
							},
							Close: &ast.Position{Line: 1, Col: ast.Col(1 + len(`import "foo`))},
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
				LParen: &ast.Position{Line: 1, Col: ast.Col(1 + len("import "))},
				Specs: []*ast.ImportSpec{
					{
						Path: &ast.String{
							Open:  &ast.Position{Line: 2, Col: ast.Col(1 + len("\t"))},
							Quote: '"',
							Contents: []ast.StringNode{
								&ast.StringText{
									Text:     "foo",
									Position: &ast.Position{Line: 2, Col: ast.Col(1 + len("\t\""))},
								},
							},
							Close: &ast.Position{Line: 2, Col: ast.Col(1 + len("\t\"foo"))},
						},
					}, {
						Path: &ast.String{
							Open:  &ast.Position{Line: 3, Col: ast.Col(1 + len("\t"))},
							Quote: '"',
							Contents: []ast.StringNode{
								&ast.StringText{
									Text:     "bar",
									Position: &ast.Position{Line: 3, Col: ast.Col(1 + len("\t\""))},
								},
							},
							Close: &ast.Position{Line: 3, Col: ast.Col(1 + len("\t\"bar"))},
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
				Path: &ast.String{
					Open:  &ast.Position{Line: 1, Col: 1},
					Quote: '"',
					Contents: []ast.StringNode{
						&ast.StringText{
							Text:     "foo",
							Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"`))},
						},
					},
					Close: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo`))},
				},
			},
		}, {
			name: "alias",
			in:   "foo \"bar\"",
			want: &ast.ImportSpec{
				Alias: &ast.Identifier{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
				Path: &ast.String{
					Open:  &ast.Position{Line: 1, Col: ast.Col(1 + len("foo "))},
					Quote: '"',
					Contents: []ast.StringNode{
						&ast.StringText{
							Text:     "bar",
							Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`foo "`))},
						},
					},
					Close: &ast.Position{Line: 1, Col: ast.Col(1 + len(`foo "bar`))},
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
