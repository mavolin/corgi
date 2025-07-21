package attribute

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestDefinition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.AttributeDefinition
	}{
		{
			name: "single",
			in:   `attr foo { * innocuous }`,
			want: &ast.AttributeDefinition{
				Specs: []*ast.AttributeSpec{
					{
						Selector: &ast.BasicAttributeSelector{
							Name:     "foo",
							Position: &ast.Position{Line: 1, Col: 6},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: &ast.Position{Line: 1, Col: 10},
							Rules: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: &ast.Position{Line: 1, Col: 12},
									},
									Type: &ast.AttributeTypeName{
										Name:     "innocuous",
										Type:     attrtype.Innocuous,
										Position: &ast.Position{Line: 1, Col: 14},
									},
								},
							},
							RBrace: &ast.Position{Line: 1, Col: 24},
						},
					},
				},
				Attr: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "single with prefix",
			in:   `attr hx- foo { * innocuous }`,
			want: &ast.AttributeDefinition{
				Prefix: &ast.AttributeName{
					Name:     "hx-",
					Position: &ast.Position{Line: 1, Col: 6},
				},
				Specs: []*ast.AttributeSpec{
					{
						Selector: &ast.BasicAttributeSelector{
							Name:     "foo",
							Position: &ast.Position{Line: 1, Col: 10},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: &ast.Position{Line: 1, Col: 14},
							Rules: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: &ast.Position{Line: 1, Col: 16},
									},
									Type: &ast.AttributeTypeName{
										Name:     "innocuous",
										Type:     attrtype.Innocuous,
										Position: &ast.Position{Line: 1, Col: 18},
									},
								},
							},
							RBrace: &ast.Position{Line: 1, Col: 28},
						},
					},
				},
				Attr: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "multiple",
			in: "attr (\n" +
				"\tfoo { * innocuous }\n" +
				"\tbar { * text }\n" +
				")",
			want: &ast.AttributeDefinition{
				LParen: &ast.Position{Line: 1, Col: 6},
				Specs: []*ast.AttributeSpec{
					{
						Selector: &ast.BasicAttributeSelector{
							Name:     "foo",
							Position: &ast.Position{Line: 2, Col: 2},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: &ast.Position{Line: 2, Col: 6},
							Rules: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: &ast.Position{Line: 2, Col: 8},
									},
									Type: &ast.AttributeTypeName{
										Name:     "innocuous",
										Type:     attrtype.Innocuous,
										Position: &ast.Position{Line: 2, Col: 10},
									},
								},
							},
							RBrace: &ast.Position{Line: 2, Col: 20},
						},
					}, {
						Selector: &ast.BasicAttributeSelector{
							Name:     "bar",
							Position: &ast.Position{Line: 3, Col: 2},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: &ast.Position{Line: 3, Col: 6},
							Rules: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: &ast.Position{Line: 3, Col: 8},
									},
									Type: &ast.AttributeTypeName{
										Name:     "text",
										Type:     attrtype.Text,
										Position: &ast.Position{Line: 3, Col: 10},
									},
								},
							},
							RBrace: &ast.Position{Line: 3, Col: 15},
						},
					},
				},
				RParen: &ast.Position{Line: 4, Col: 1},
				Attr:   &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "multiple with prefix",
			in: "attr hx- (\n" +
				"\tfoo { * innocuous }\n" +
				"\tbar { * text }\n" +
				")",
			want: &ast.AttributeDefinition{
				Prefix: &ast.AttributeName{
					Name:     "hx-",
					Position: &ast.Position{Line: 1, Col: 6},
				},
				LParen: &ast.Position{Line: 1, Col: 10},
				Specs: []*ast.AttributeSpec{
					{
						Selector: &ast.BasicAttributeSelector{
							Name:     "foo",
							Position: &ast.Position{Line: 2, Col: 2},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: &ast.Position{Line: 2, Col: 6},
							Rules: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: &ast.Position{Line: 2, Col: 8},
									},
									Type: &ast.AttributeTypeName{
										Name:     "innocuous",
										Type:     attrtype.Innocuous,
										Position: &ast.Position{Line: 2, Col: 10},
									},
								},
							},
							RBrace: &ast.Position{Line: 2, Col: 20},
						},
					}, {
						Selector: &ast.BasicAttributeSelector{
							Name:     "bar",
							Position: &ast.Position{Line: 3, Col: 2},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: &ast.Position{Line: 3, Col: 6},
							Rules: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: &ast.Position{Line: 3, Col: 8},
									},
									Type: &ast.AttributeTypeName{
										Name:     "text",
										Type:     attrtype.Text,
										Position: &ast.Position{Line: 3, Col: 10},
									},
								},
							},
							RBrace: &ast.Position{Line: 3, Col: 15},
						},
					},
				},
				RParen: &ast.Position{Line: 4, Col: 1},
				Attr:   &ast.Position{Line: 1, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c.in, Definition())
			should.Equal(t, c.want, got)
		})
	}
}

func TestSpec(t *testing.T) {
	t.Parallel()

	in := `foo { * innocuous }`
	want := &ast.AttributeSpec{
		Selector: &ast.BasicAttributeSelector{
			Name:     "foo",
			Position: &ast.Position{Line: 1, Col: 1},
		},
		Ruleset: &ast.AttributeRuleset{
			LBrace: &ast.Position{Line: 1, Col: 5},
			Rules: []*ast.AttributeRule{
				{
					Selector: &ast.WildcardElementSelector{
						Asterisk: &ast.Position{Line: 1, Col: 7},
					},
					Type: &ast.AttributeTypeName{
						Name:     "innocuous",
						Type:     attrtype.Innocuous,
						Position: &ast.Position{Line: 1, Col: 9},
					},
				},
			},
			RBrace: &ast.Position{Line: 1, Col: 19},
		},
	}

	got := parsetest.ParsesFully(t, in, Spec())
	should.Equal(t, want, got)
}

func TestRuleset(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want *ast.AttributeRuleset
	}{
		{
			name: "single rule on single line",
			in:   "{ * innocuous }",
			want: &ast.AttributeRuleset{
				LBrace: &ast.Position{Line: 1, Col: 1},
				Rules: []*ast.AttributeRule{
					{
						Selector: &ast.WildcardElementSelector{
							Asterisk: &ast.Position{Line: 1, Col: 3},
						},
						Type: &ast.AttributeTypeName{
							Name:     "innocuous",
							Type:     attrtype.Innocuous,
							Position: &ast.Position{Line: 1, Col: 5},
						},
					},
				},
				RBrace: &ast.Position{Line: 1, Col: 15},
			},
		}, {
			name: "multiple rules on single line",
			in:   "{ * innocuous; foo text }",
			want: &ast.AttributeRuleset{
				LBrace: &ast.Position{Line: 1, Col: 1},
				Rules: []*ast.AttributeRule{
					{
						Selector: &ast.WildcardElementSelector{
							Asterisk: &ast.Position{Line: 1, Col: 3},
						},
						Type: &ast.AttributeTypeName{
							Name:     "innocuous",
							Type:     attrtype.Innocuous,
							Position: &ast.Position{Line: 1, Col: 5},
						},
					}, {
						Selector: &ast.ListElementSelector{
							Elements: []*ast.ElementName{
								{
									Name:     "foo",
									Position: &ast.Position{Line: 1, Col: 16},
								},
							},
						},
						Type: &ast.AttributeTypeName{
							Name:     "text",
							Type:     attrtype.Text,
							Position: &ast.Position{Line: 1, Col: 20},
						},
					},
				},
				RBrace: &ast.Position{Line: 1, Col: 25},
			},
		}, {
			name: "single rule on multiple lines",
			in: "{\n" +
				"\t* innocuous\n" +
				"}",
			want: &ast.AttributeRuleset{
				LBrace: &ast.Position{Line: 1, Col: 1},
				Rules: []*ast.AttributeRule{
					{
						Selector: &ast.WildcardElementSelector{
							Asterisk: &ast.Position{Line: 2, Col: 2},
						},
						Type: &ast.AttributeTypeName{
							Name:     "innocuous",
							Type:     attrtype.Innocuous,
							Position: &ast.Position{Line: 2, Col: 4},
						},
					},
				},
				RBrace: &ast.Position{Line: 3, Col: 1},
			},
		}, {
			name: "multiple rules on multiple lines",
			in: "{\n" +
				"\t* innocuous\n" +
				"\tfoo text\n" +
				"}",
			want: &ast.AttributeRuleset{
				LBrace: &ast.Position{Line: 1, Col: 1},
				Rules: []*ast.AttributeRule{
					{
						Selector: &ast.WildcardElementSelector{
							Asterisk: &ast.Position{Line: 2, Col: 2},
						},
						Type: &ast.AttributeTypeName{
							Name:     "innocuous",
							Type:     attrtype.Innocuous,
							Position: &ast.Position{Line: 2, Col: 4},
						},
					}, {
						Selector: &ast.ListElementSelector{
							Elements: []*ast.ElementName{
								{
									Name:     "foo",
									Position: &ast.Position{Line: 3, Col: 2},
								},
							},
						},
						Type: &ast.AttributeTypeName{
							Name:     "text",
							Type:     attrtype.Text,
							Position: &ast.Position{Line: 3, Col: 6},
						},
					},
				},
				RBrace: &ast.Position{Line: 4, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c.in, Ruleset())
			should.Equal(t, c.want, got)
		})
	}
}

func TestRule(t *testing.T) {
	t.Parallel()

	in := "* innocuous"
	want := &ast.AttributeRule{
		Selector: &ast.WildcardElementSelector{
			Asterisk: &ast.Position{Line: 1, Col: 1},
		},
		Type: &ast.AttributeTypeName{
			Name:     "innocuous",
			Type:     attrtype.Innocuous,
			Position: &ast.Position{Line: 1, Col: 3},
		},
	}

	got := parsetest.ParsesFully(t, in, Rule())
	should.Equal(t, want, got)
}

func TestSelector(t *testing.T) {
	t.Parallel()

	parsetest.AssertAlsoFulfils(t, Selector(), testBasicSelector)
	parsetest.AssertAlsoFulfils(t, Selector(), testRegexpSelector)
}

func TestBasicSelector(t *testing.T) {
	t.Parallel()
	testBasicSelector(t, BasicSelector())
}

func testBasicSelector(t *testing.T, f parser.Func[*ast.BasicAttributeSelector]) {
	tests := []struct {
		in   string
		want *ast.BasicAttributeSelector
	}{
		{
			in:   "foo",
			want: &ast.BasicAttributeSelector{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
		}, {
			in:   "foo*",
			want: &ast.BasicAttributeSelector{Name: "foo", Wildcard: true, Position: &ast.Position{Line: 1, Col: 1}},
		}, {
			in:   "foo*bar",
			want: &ast.BasicAttributeSelector{Name: "foo*bar", Position: &ast.Position{Line: 1, Col: 1}},
		},
	}

	for _, c := range tests {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c.in, f)
			should.Equal(t, c.want, got)
		})
	}
}

func TestRegexpSelector(t *testing.T) {
	t.Parallel()
	testRegexpSelector(t, RegexpSelector())
}

func testRegexpSelector(t *testing.T, f parser.Func[*ast.RegexpAttributeSelector]) {
	in := `'regexp("foo\\d")`
	want := &ast.RegexpAttributeSelector{
		LParen: &ast.Position{Line: 1, Col: 8},
		Raw: &ast.StaticString{
			Open:     &ast.Position{Line: 1, Col: 9},
			Quote:    '"',
			Contents: `foo\\d`,
			Close:    &ast.Position{Line: 1, Col: 16},
		},
		Compiled: regexp.MustCompile("foo\\d"),
		RParen:   &ast.Position{Line: 1, Col: 17},
		Regexp:   &ast.Position{Line: 1, Col: 1},
	}

	got := parsetest.ParsesFully(t, in, f)
	should.Equal(t, want, got, cmpopts.IgnoreFields(ast.RegexpAttributeSelector{}, "Compiled"))
}

func TestElementSelector(t *testing.T) {
	t.Parallel()

	parsetest.AssertAlsoFulfils(t, ElementSelector(), testWildcardElementSelector)
	parsetest.AssertAlsoFulfils(t, ElementSelector(), testListElementSelector)
}

func TestWildcardElementSelector(t *testing.T) {
	t.Parallel()
	testWildcardElementSelector(t, WildcardElementSelector())
}

func testWildcardElementSelector(t *testing.T, f parser.Func[*ast.WildcardElementSelector]) {
	in := "*"
	want := &ast.WildcardElementSelector{Asterisk: &ast.Position{Line: 1, Col: 1}}

	got := parsetest.ParsesFully(t, in, f)
	should.Equal(t, want, got)
}

func TestListElementSelector(t *testing.T) {
	t.Parallel()
	testListElementSelector(t, ListElementSelector())
}

func testListElementSelector(t *testing.T, f parser.Func[*ast.ListElementSelector]) {
	tests := []struct {
		name string
		in   string
		want *ast.ListElementSelector
	}{
		{
			name: "single",
			in:   "foo",
			want: &ast.ListElementSelector{
				Elements: []*ast.ElementName{
					{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
				},
			},
		}, {
			name: "multiple",
			in: "foo,\n" +
				"\tbar, baz",
			want: &ast.ListElementSelector{
				Elements: []*ast.ElementName{
					{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}},
					{Name: "bar", Position: &ast.Position{Line: 2, Col: 2}},
					{Name: "baz", Position: &ast.Position{Line: 2, Col: 7}},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c.in, f)
			should.Equal(t, c.want, got)
		})
	}
}

func TestListElementSelectorItem(t *testing.T) {
	t.Parallel()

	in := "foo"
	want := &ast.ElementName{Name: "foo", Position: &ast.Position{Line: 1, Col: 1}}

	got := parsetest.ParsesFully(t, in, elementName())
	should.Equal(t, want, got)
}
