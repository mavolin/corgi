package diagnostic

import (
	"errors"
	"testing"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
)

func TestDiagnostic_Pretty(t *testing.T) {
	t.Parallel()

	imp := &ast.Import{
		Import: &ast.Position{Line: 3, Col: 1},
		Specs: []*ast.ImportSpec{
			{
				Path: &ast.String{
					Open:  &ast.Position{Line: 3, Col: ast.Col(1 + len("import "))},
					Quote: '"',
					Contents: []ast.StringNode{
						&ast.StringText{
							Text:     "bar",
							Position: &ast.Position{Line: 3, Col: ast.Col(1 + len(`import "`))},
						},
					},
					Close: &ast.Position{Line: 3, Col: ast.Col(1 + len(`import "bar"`))},
				},
			},
		},
	}

	f := &file.File{
		Package: &file.Package{
			PathInModule: "bar",
		},
		Name: "foo.corgi",
		Raw: "package foo\n" +
			"\n" +
			"import \"bar\"",
		Lines: []string{
			"package foo",
			"",
			"import \"bar\"",
		},
		AST: &ast.File{
			Package: &ast.PackageDirective{
				Package: &ast.Position{Line: 1, Col: 1},
				Name: &ast.Identifier{
					Name:     "foo",
					Position: &ast.Position{Line: 1, Col: ast.Col(1 + len("package "))},
				},
			},
			Imports: []*ast.Import{imp},
		},
	}

	packageAnno := Annotation{
		File:         f,
		ContextStart: 1,
		ContextEnd:   2,
		Start:        ast.Position{Line: 1, Col: 1},
		End:          ast.Position{Line: 1, Col: ast.Col(1 + len("package foo"))},
		Annotation:   "package",
	}
	packageWordAnno := Annotation{
		File:         f,
		ContextStart: 1,
		ContextEnd:   2,
		Start:        ast.Position{Line: 1, Col: 1},
		End:          ast.Position{Line: 1, Col: ast.Col(1 + len("package"))},
		Annotation:   "package word",
	}
	packageNameAnno := Annotation{
		File:         f,
		ContextStart: 1,
		ContextEnd:   2,
		Start:        ast.Position{Line: 1, Col: ast.Col(1 + len("package "))},
		End:          ast.Position{Line: 1, Col: ast.Col(1 + len("package foo"))},
		Annotation:   "package name",
	}
	importAnno := Annotation{
		File:         f,
		ContextStart: 3,
		ContextEnd:   4,
		Start:        ast.Position{Line: 3, Col: 1},
		End:          ast.Position{Line: 3, Col: ast.Col(1 + len("import \"bar\""))},
		Annotation:   "import",
	}

	tests := []struct {
		name         string
		diag         *Diagnostic
		expectShort  string
		expectPretty string
	}{
		{
			name: "only message",
			diag: &Diagnostic{
				Message: "foo",
			},
			expectShort:  "error: foo",
			expectPretty: "error: foo",
		}, {
			name: "warning",
			diag: &Diagnostic{
				Type:    Warning,
				Message: "foo",
			},
			expectShort:  "warning: foo",
			expectPretty: "warning: foo",
		}, {
			name: "single annotation",
			diag: &Diagnostic{
				Message: "foo",
				Primary: []Annotation{packageAnno},
			},
			expectShort: "error: bar/foo.corgi:1:1: foo",
			expectPretty: "error: foo\n" +
				"  ╭─ bar/foo.corgi:1:1\n" +
				"1 │ package foo\n" +
				"  │ ^^^^^^^^^^^ package",
		}, {
			name: "multiple annotations",
			diag: &Diagnostic{
				Message: "foo",
				Primary: []Annotation{
					packageWordAnno,
					packageNameAnno,
				},
			},
			expectShort: "error: bar/foo.corgi:1:1: foo",
			expectPretty: "error: foo\n" +
				"  ╭─ bar/foo.corgi:1:1\n" +
				"1 │ package foo\n" +
				"  │ ^^^^^^^ ^^^ package name\n" +
				"  │ ╰ package word",
		}, {
			name: "multi-line annotation",
			diag: &Diagnostic{
				Message: "foo",
				Primary: []Annotation{
					{
						File:         f,
						ContextStart: 1,
						ContextEnd:   2,
						Start:        ast.Position{Line: 1, Col: 1},
						End:          ast.Position{Line: 1, Col: ast.Col(1 + len("package foo"))},
						Annotation:   "bar\nbaz",
					},
				},
			},
			expectShort: "error: bar/foo.corgi:1:1: foo",
			expectPretty: "error: foo\n" +
				"  ╭─ bar/foo.corgi:1:1\n" +
				"1 │ package foo\n" +
				"  │ ^^^^^^^^^^^\n" +
				"  │ ╰ bar\n" +
				"  │   baz",
		}, {
			name: "big context",
			diag: &Diagnostic{
				Message: "foo",
				Primary: []Annotation{
					{
						File:         f,
						ContextStart: 1,
						ContextEnd:   4,
						Start:        ast.Position{Line: 1, Col: 1},
						End:          ast.Position{Line: 1, Col: ast.Col(1 + len("package foo"))},
						Annotation:   "package",
					},
				},
			},
			expectShort: "error: bar/foo.corgi:1:1: foo",
			expectPretty: "error: foo\n" +
				"  ╭─ bar/foo.corgi:1:1\n" +
				"1 │ package foo\n" +
				"  │ ^^^^^^^^^^^ package\n" +
				"2 │ \n" +
				"3 │ import \"bar\"",
		}, {
			name: "primary and secondary",
			diag: &Diagnostic{
				Message:   "foo",
				Primary:   []Annotation{packageAnno},
				Secondary: []Annotation{importAnno},
			},
			expectShort: "error: bar/foo.corgi:1:1: foo",
			expectPretty: "error: foo\n" +
				"  ╭─ bar/foo.corgi:1:1\n" +
				"1 │ package foo\n" +
				"  │ ^^^^^^^^^^^ package\n" +
				"2 │ \n" +
				`3 │ import "bar"` + "\n" +
				"  │ ~~~~~~~~~~~~ import",
		}, {
			name: "cause",
			diag: &Diagnostic{
				Message: "foo",
				Cause:   errors.New("bar"),
			},
			expectShort: "error: foo: bar",
			expectPretty: "error: foo\n" +
				"\n" +
				"Cause: Bar",
		}, {
			name: "explanation",
			diag: &Diagnostic{
				Message:     "foo",
				Explanation: "bar",
			},
			expectShort: "error: foo",
			expectPretty: "error: foo\n" +
				"\n" +
				"bar",
		}, {
			name: "example",
			diag: &Diagnostic{
				Message:  "foo",
				Examples: []Example{{Example: "bar"}},
			},
			expectShort: "error: foo",
			expectPretty: "error: foo\n" +
				"\n" +
				"Example: bar",
		}, {
			name: "examples",
			diag: &Diagnostic{
				Message: "foo",
				Examples: []Example{
					{Example: "bar"},
					{Example: "baz"},
				},
			},
			expectShort: "error: foo",
			expectPretty: "error: foo\n" +
				"\n" +
				"Examples: bar\n" +
				"          baz",
		}, {
			name: "examples with titles",
			diag: &Diagnostic{
				Message: "foo",
				Examples: []Example{
					{Example: "qux", Title: "bar"},
					{Example: "quux", Title: "baz"},
				},
			},
			expectShort: "error: foo",
			expectPretty: "error: foo\n" +
				"\n" +
				"Examples: qux  (bar)\n" +
				"          quux (baz)",
		}, {
			name: "examples with headlines",
			diag: &Diagnostic{
				Message: "foo",
				Examples: []Example{
					{Example: "qux\nqax", Title: "bar"},
					{Example: "quux", Title: "baz"},
				},
			},
			expectShort: "error: foo",
			expectPretty: "error: foo\n" +
				"\n" +
				"Examples: *bar*\n" +
				"          qux\n" +
				"          qax\n" +
				"          *baz*\n" +
				"          quux",
		}, {
			name: "hint",
			diag: &Diagnostic{
				Message: "foo",
				Hints:   []Hint{{Hint: "bar"}},
			},
			expectShort: "error: foo",
			expectPretty: "error: foo\n" +
				"\n" +
				"Hint: bar",
		}, {
			name: "hint with example",
			diag: &Diagnostic{
				Message: "foo",
				Hints:   []Hint{{Hint: "bar", Example: "baz"}},
			},
			expectShort: "error: foo",
			expectPretty: "error: foo\n" +
				"\n" +
				"Hint: bar\n" +
				"      |> baz",
		}, {
			name: "hints",
			diag: &Diagnostic{
				Message: "foo",
				Hints: []Hint{
					{Hint: "bar"},
					{Hint: "baz"},
				},
			},
			expectShort: "error: foo",
			expectPretty: "error: foo\n" +
				"\n" +
				"Hints: bar\n" +
				"       baz",
		}, {
			name: "hints with example",
			diag: &Diagnostic{
				Message: "foo",
				Hints: []Hint{
					{Hint: "bar", Example: "baz"},
					{Hint: "qux", Example: "quux"},
				},
			},
			expectShort: "error: foo",
			expectPretty: "error: foo\n" +
				"\n" +
				"Hints: bar\n" +
				"       |> baz\n" +
				"       qux\n" +
				"       |> quux",
		}, {
			name: "docs",
			diag: &Diagnostic{
				Message: "foo",
				Docs:    "bar",
			},
			expectShort: "error: foo",
			expectPretty: "error: foo\n" +
				"\n" +
				"Docs: https://corgi.mavolin.co/!bar",
		}, {
			name: "full",
			diag: &Diagnostic{
				Message:     "foo",
				Primary:     []Annotation{packageAnno},
				Secondary:   []Annotation{importAnno},
				Cause:       errors.New("bar"),
				Explanation: "baz",
				Examples: []Example{
					{Example: "qux"},
				},
				Hints: []Hint{
					{Hint: "quux"},
				},
				Docs: "corgi",
			},
			expectShort: "error: bar/foo.corgi:1:1: foo: bar",
			expectPretty: "error: foo\n" +
				"  ╭─ bar/foo.corgi:1:1\n" +
				"1 │ package foo\n" +
				"  │ ^^^^^^^^^^^ package\n" +
				"2 │ \n" +
				`3 │ import "bar"` + "\n" +
				"  │ ~~~~~~~~~~~~ import\n" +
				"\n" +
				"Cause: Bar\n" +
				"\n" +
				"baz\n" +
				"\n" +
				"Example: qux\n" +
				"\n" +
				"Hint: quux\n" +
				"\n" +
				"Docs: https://corgi.mavolin.co/!corgi",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			// short message
			actualShort := c.diag.Short()
			should.Equal(t, actualShort, c.expectShort)

			// pretty message
			actualPretty := c.diag.Pretty(PrettyOptions{})
			should.Equal(t, actualPretty, c.expectPretty)
		})
	}
}
