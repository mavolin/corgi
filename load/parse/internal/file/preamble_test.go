package file

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestPackageDirective(t *testing.T) {
	t.Parallel()

	in := "package foo"
	expect := &ast.PackageDirective{
		Package: ast.Position{Line: 1, Col: 1},
		Name:    &ast.Ident{Ident: "foo", Position: ast.Position{Line: 1, Col: 9}},
	}

	actual := testutil.ParsesFully(t, in+";", PackageDirective())
	assert.Equal(t, expect, actual)
}

func TestImport(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.Import
	}{
		{
			name: "single",
			in:   "import \"foo\"",
			expect: &ast.Import{
				Import: ast.Position{Line: 1, Col: 1},
				Specs: []*ast.ImportSpec{
					{
						Path: &ast.StaticString{
							Open:     ast.Position{Line: 1, Col: 8},
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
			expect: &ast.Import{
				Import: ast.Position{Line: 1, Col: 1},
				LParen: &ast.Position{Line: 1, Col: 8},
				Specs: []*ast.ImportSpec{
					{
						Path: &ast.StaticString{
							Open:     ast.Position{Line: 2, Col: 2},
							Quote:    '"',
							Contents: "foo",
							Close:    &ast.Position{Line: 2, Col: 6},
						},
					}, {
						Path: &ast.StaticString{
							Open:     ast.Position{Line: 3, Col: 2},
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

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in+";", Import())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestImportSpec(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ImportSpec
	}{
		{
			name: "no alias",
			in:   "\"foo\"",
			expect: &ast.ImportSpec{
				Path: &ast.StaticString{
					Open:     ast.Position{Line: 1, Col: 1},
					Quote:    '"',
					Contents: "foo",
					Close:    &ast.Position{Line: 1, Col: 5},
				},
			},
		}, {
			name: "alias",
			in:   "foo \"bar\"",
			expect: &ast.ImportSpec{
				Alias: &ast.Ident{Ident: "foo", Position: ast.Position{Line: 1, Col: 1}},
				Path: &ast.StaticString{
					Open:     ast.Position{Line: 1, Col: 5},
					Quote:    '"',
					Contents: "bar",
					Close:    &ast.Position{Line: 1, Col: 9},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in, ImportSpec())
			assert.Equal(t, c.expect, actual)
		})
	}
}
