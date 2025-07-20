package file

import (
	"strings"
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFile(t *testing.T) {
	t.Parallel()

	in := "package foo\n" +
		"\n" +
		"import (\n" +
		"\t\"fmt\"\n" +
		")\n" +
		"\n" +
		"// foo barks.\n" +
		"comp foo() {\n" +
		"\t> woof\n" +
		"\t\tbark\n" +
		"}\n"
	lines := strings.Split(in, "\n")
	for i, line := range lines {
		last := len(line) - 1
		if len(line) > 0 && line[last] == '\r' {
			lines[i] = line[:last]
		}
	}
	expect := &ast.File{
		Raw:   in,
		Lines: lines,
		Package: &ast.PackageDirective{
			Package: &ast.Position{Line: 1, Col: 1},
			Name:    &ast.Identifier{Name: "foo", Position: &ast.Position{Line: 1, Col: 9}},
		},
		Imports: []*ast.Import{
			{
				Import: &ast.Position{Line: 3, Col: 1},
				LParen: &ast.Position{Line: 3, Col: 8},
				Specs: []*ast.ImportSpec{
					{
						Path: &ast.StaticString{
							Open:     &ast.Position{Line: 4, Col: 2},
							Quote:    '"',
							Contents: "fmt",
							Close:    &ast.Position{Line: 4, Col: 6},
						},
					},
				},
				RParen: &ast.Position{Line: 5, Col: 1},
			},
		},
		TopLevel: []ast.ScopeNode{
			&ast.Component{
				Comp: &ast.Position{Line: 8, Col: 1},
				Header: &ast.ComponentHeader{
					Name: &ast.Identifier{Name: "foo", Position: &ast.Position{Line: 8, Col: 6}},
					Parameters: &ast.ComponentParameters{
						LParen: &ast.Position{Line: 8, Col: 9},
						RParen: &ast.Position{Line: 8, Col: 10},
					},
				},
				Body: &ast.Scope{
					LBrace: &ast.Position{Line: 8, Col: 12},
					Nodes: []ast.ScopeNode{
						&ast.ArrowBlock{
							Arrow: &ast.Position{Line: 9, Col: 2},
							Lines: ast.TextBlock{
								ast.TextLine{
									&ast.Text{
										Text:     "woof",
										Position: &ast.Position{Line: 9, Col: 4},
									},
								},
								ast.TextLine{
									&ast.Text{
										Text:     "bark",
										Position: &ast.Position{Line: 10, Col: 3},
									},
								},
							},
						},
					},
					RBrace: &ast.Position{Line: 11, Col: 1},
				},
			},
		},
		Comments: []*ast.CommentGroup{
			{
				Comments: []*ast.Comment{
					{
						Open:    &ast.Position{Line: 7, Col: 1},
						Comment: " foo barks.",
						Until:   ast.Position{Line: 7, Col: 14},
					},
				},
			},
		},
	}

	p := testutil.NewParser(t, in)
	_, err := parser.TryErr(p, File())
	if err != nil {
		require.Fail(t, "unexpected error: %v", err)
	}

	assert.Equal(t, expect, p.File.AST)
}
