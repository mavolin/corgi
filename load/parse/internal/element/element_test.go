package element

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDoctype(t *testing.T) {
	t.Parallel()

	in := "!doctype(html)"
	expect := &ast.Doctype{
		Doctype: &ast.Position{Line: 1, Col: 1},
		LParen:  &ast.Position{Line: 1, Col: 9},
		HTML:    &ast.Position{Line: 1, Col: 10},
		RParen:  &ast.Position{Line: 1, Col: 14},
	}

	actual := testutil.ParsesFully(t, in, Doctype())
	assert.Equal(t, expect, actual)
}

func TestElement(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.Element
	}{
		{
			name: "void",
			in:   "br",
			expect: &ast.Element{
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
			expect: &ast.Element{
				Header: &ast.ElementHeader{
					Name: &ast.ElementReference{
						Name: &ast.ElementName{
							Name:     "br",
							Position: &ast.Position{Line: 1, Col: 1},
						},
					},
					Attributes: &ast.Arguments{
						LParen: &ast.Position{Line: 1, Col: 3},
						Args: []ast.Argument{
							&ast.NamedAttribute{
								Name: &ast.AttributeReference{
									Name: &ast.AttributeName{
										Name:     "foo",
										Position: &ast.Position{Line: 1, Col: 4},
									},
								},
								EqualSign: &ast.Position{Line: 1, Col: 7},
								Value: &ast.ExpressionAttributeValue{
									Code: ast.Code{
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
			expect: &ast.Element{
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

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c.in, Element())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestHeader(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ElementHeader
	}{
		{
			name: "name",
			in:   "br",
			expect: &ast.ElementHeader{
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
			expect: &ast.ElementHeader{
				Name: &ast.ElementReference{
					Name: &ast.ElementName{
						Name:     "br",
						Position: &ast.Position{Line: 1, Col: 1},
					},
				},
				Attributes: &ast.Arguments{
					LParen: &ast.Position{Line: 1, Col: 3},
					Args: []ast.Argument{
						&ast.NamedAttribute{
							Name: &ast.AttributeReference{
								Name: &ast.AttributeName{
									Name:     "foo",
									Position: &ast.Position{Line: 1, Col: 4},
								},
							},
							EqualSign: &ast.Position{Line: 1, Col: 7},
							Value: &ast.ExpressionAttributeValue{
								Code: ast.Code{
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

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c.in, Header())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestReference(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ElementReference
	}{
		{
			name: "local",
			in:   "name",
			expect: &ast.ElementReference{
				Name: &ast.ElementName{
					Name:     "name",
					Position: &ast.Position{Line: 1, Col: 1},
				},
			},
		}, {
			name: "external",
			in:   "package1.Name",
			expect: &ast.ElementReference{
				Package: &ast.Ident{
					Ident:    "package1",
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

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in, Reference())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestName(t *testing.T) {
	t.Parallel()

	in := "br"
	expect := &ast.ElementName{
		Name:     "br",
		Position: &ast.Position{Line: 1, Col: 1},
	}

	actual := testutil.ParsesFully(t, in, Name())
	assert.Equal(t, expect, actual)
}

func TestRaw(t *testing.T) {
	t.Parallel()

	in := "!raw [ foo ]"
	expect := &ast.RawElement{
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

	actual := testutil.ParsesFully(t, in, Raw())
	assert.Equal(t, expect, actual)
}

func TestAnd(t *testing.T) {
	t.Parallel()

	in := "&(foo)"
	expect := &ast.And{
		And: &ast.Position{Line: 1, Col: 1},
		Attributes: &ast.Arguments{
			LParen: &ast.Position{Line: 1, Col: 2},
			Args: []ast.Argument{
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

	actual := testutil.ParsesFully(t, in, And())
	assert.Equal(t, expect, actual)
}
