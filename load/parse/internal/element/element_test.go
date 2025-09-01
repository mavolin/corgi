package element

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestDoctype(t *testing.T) {
	t.Parallel()

	in := "!doctype(html)"
	want := &ast.Doctype{
		Doctype: &ast.Position{Line: 1, Col: 1},
		LParen:  &ast.Position{Line: 1, Col: 9},
		HTML:    &ast.Position{Line: 1, Col: 10},
		RParen:  &ast.Position{Line: 1, Col: 14},
	}

	got := parsetest.ParsesFully(t, in, Doctype())
	should.Equal(t, got, want)
}

func TestElement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.Element
	}{
		{
			name: "void",
			in:   "br",
			want: &ast.Element{
				Header: &ast.ElementHeader{
					Name: &ast.ElementReference{
						Name: &ast.ElementName{
							Name:     "br",
							Position: &ast.Position{Line: 1, Col: 1},
						},
					},
				},
			},
		}, {
			name: "void with attributes",
			in:   "br(foo=bar)",
			want: &ast.Element{
				Header: &ast.ElementHeader{
					Name: &ast.ElementReference{
						Name: &ast.ElementName{
							Name:     "br",
							Position: &ast.Position{Line: 1, Col: 1},
						},
					},
					Attributes: &ast.Arguments{
						LParen: &ast.Position{Line: 1, Col: 3},
						List: []ast.Argument{
							&ast.NamedAttribute{
								Name: &ast.AttributeReference{
									Name: &ast.AttributeName{
										Name:     "foo",
										Position: &ast.Position{Line: 1, Col: 4},
									},
								},
								EqualSign: &ast.Position{Line: 1, Col: 7},
								Value: &ast.ExpressionAttributeValue{
									Nodes: ast.Code{
										&ast.GoCode{
											Code:     "bar",
											Position: &ast.Position{Line: 1, Col: 8},
										},
									},
								},
							},
						},
						RParen: &ast.Position{Line: 1, Col: 11},
					},
				},
			},
		}, {
			name: "with body",
			in:   "div [ foo ]",
			want: &ast.Element{
				Header: &ast.ElementHeader{
					Name: &ast.ElementReference{
						Name: &ast.ElementName{
							Name:     "div",
							Position: &ast.Position{Line: 1, Col: 1},
						},
					},
				},
				Body: &ast.BracketText{
					LBracket: &ast.Position{Line: 1, Col: 5},
					Lines: ast.TextBlock{
						ast.TextLine{
							&ast.Text{
								Text:     "foo",
								Position: &ast.Position{Line: 1, Col: 7},
							},
						},
					},
					RBracket: &ast.Position{Line: 1, Col: 11},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			p := parsetest.NewParser(t, c.in+"; 1other stuff")
			got := parsetest.AssertNoError(t, p, Element())

			line, col, index := parsetest.CalcEnd(1, 1, 0, c.in)
			parsetest.AssertPosition(t, p, line, col, index)
			should.Equal(t, got, c.want)
		})
	}
}

func TestHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.ElementHeader
	}{
		{
			name: "name",
			in:   "br",
			want: &ast.ElementHeader{
				Name: &ast.ElementReference{
					Name: &ast.ElementName{
						Name:     "br",
						Position: &ast.Position{Line: 1, Col: 1},
					},
				},
			},
		}, {
			name: "name with attributes",
			in:   "br(foo=bar)",
			want: &ast.ElementHeader{
				Name: &ast.ElementReference{
					Name: &ast.ElementName{
						Name:     "br",
						Position: &ast.Position{Line: 1, Col: 1},
					},
				},
				Attributes: &ast.Arguments{
					LParen: &ast.Position{Line: 1, Col: 3},
					List: []ast.Argument{
						&ast.NamedAttribute{
							Name: &ast.AttributeReference{
								Name: &ast.AttributeName{
									Name:     "foo",
									Position: &ast.Position{Line: 1, Col: 4},
								},
							},
							EqualSign: &ast.Position{Line: 1, Col: 7},
							Value: &ast.ExpressionAttributeValue{
								Nodes: ast.Code{
									&ast.GoCode{
										Code:     "bar",
										Position: &ast.Position{Line: 1, Col: 8},
									},
								},
							},
						},
					},
					RParen: &ast.Position{Line: 1, Col: 11},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := parsetest.ParsesFully(t, c.in, Header())
			should.Equal(t, got, c.want)
		})
	}
}

func TestReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.ElementReference
	}{
		{
			name: "local",
			in:   "name",
			want: &ast.ElementReference{
				Name: &ast.ElementName{
					Name:     "name",
					Position: &ast.Position{Line: 1, Col: 1},
				},
			},
		}, {
			name: "external",
			in:   "package1.Name",
			want: &ast.ElementReference{
				Package: &ast.Identifier{
					Name:     "package1",
					Position: &ast.Position{Line: 1, Col: 1},
				},
				Dot: &ast.Position{Line: 1, Col: 9},
				Name: &ast.ElementName{
					Name:     "Name",
					Position: &ast.Position{Line: 1, Col: 10},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c.in, Reference())
			should.Equal(t, got, c.want)
		})
	}
}

func TestName(t *testing.T) {
	t.Parallel()

	in := "br"
	want := &ast.ElementName{
		Name:     "br",
		Position: &ast.Position{Line: 1, Col: 1},
	}

	got := parsetest.ParsesFully(t, in, Name())
	should.Equal(t, got, want)
}

func TestRaw(t *testing.T) {
	t.Parallel()

	in := "!raw [ foo ]"
	want := &ast.RawElement{
		Raw: &ast.Position{Line: 1, Col: 1},
		Body: &ast.BracketText{
			LBracket: &ast.Position{Line: 1, Col: 6},
			Lines: ast.TextBlock{
				ast.TextLine{
					&ast.Text{
						Text:     "foo",
						Position: &ast.Position{Line: 1, Col: 8},
					},
				},
			},
			RBracket: &ast.Position{Line: 1, Col: 12},
		},
	}

	got := parsetest.ParsesFully(t, in, Raw())
	should.Equal(t, got, want)
}

func TestAnd(t *testing.T) {
	t.Parallel()

	in := "&(foo)"
	want := &ast.And{
		And: &ast.Position{Line: 1, Col: 1},
		Attributes: &ast.Arguments{
			LParen: &ast.Position{Line: 1, Col: 2},
			List: []ast.Argument{
				&ast.NamedAttribute{
					Name: &ast.AttributeReference{
						Name: &ast.AttributeName{
							Name:     "foo",
							Position: &ast.Position{Line: 1, Col: 3},
						},
					},
				},
			},
			RParen: &ast.Position{Line: 1, Col: 6},
		},
	}

	got := parsetest.ParsesFully(t, in, And())
	should.Equal(t, got, want)
}
