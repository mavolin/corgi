package attribute

import (
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
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
			in:   `attr foo { * string }`,
			want: &ast.AttributeDefinition{
				Specs: []*ast.AttributeSpec{
					{
						Selector: &ast.BasicAttributeSelector{
							Name:          "foo",
							CanonicalName: "foo",
							Position:      &ast.Position{Line: 1, Col: 1 + len("attr ")},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: &ast.Position{Line: 1, Col: 1 + len("attr foo ")},
							List: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: &ast.Position{Line: 1, Col: 1 + len("attr foo { ")},
									},
									Type: &ast.AttributeTypeName{
										Name:     "string",
										Type:     attrtype.String,
										Position: &ast.Position{Line: 1, Col: 1 + len("attr foo { * ")},
									},
								},
							},
							RBrace: &ast.Position{Line: 1, Col: 1 + len("attr foo { * string ")},
						},
					},
				},
				Attr: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "single with prefix",
			in:   `attr hx- foo { * string }`,
			want: &ast.AttributeDefinition{
				Prefix: &ast.AttributeName{
					Name:          "hx-",
					CanonicalName: "hx-",
					Position:      &ast.Position{Line: 1, Col: 1 + len("attr ")},
				},
				Specs: []*ast.AttributeSpec{
					{
						Selector: &ast.BasicAttributeSelector{
							Name:          "foo",
							CanonicalName: "foo",
							Position:      &ast.Position{Line: 1, Col: 1 + len("attr hx- ")},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: &ast.Position{Line: 1, Col: 1 + len("attr hx- foo ")},
							List: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: &ast.Position{Line: 1, Col: 1 + len("attr hx- foo { ")},
									},
									Type: &ast.AttributeTypeName{
										Name:     "string",
										Type:     attrtype.String,
										Position: &ast.Position{Line: 1, Col: 1 + len("attr hx- foo { * ")},
									},
								},
							},
							RBrace: &ast.Position{Line: 1, Col: 1 + len("attr hx- foo { * string ")},
						},
					},
				},
				Attr: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "multiple",
			in: "attr (\n" +
				"\tfoo { * string }\n" +
				"\tbar { * text }\n" +
				")",
			want: &ast.AttributeDefinition{
				LParen: &ast.Position{Line: 1, Col: 1 + len("attr ")},
				Specs: []*ast.AttributeSpec{
					{
						Selector: &ast.BasicAttributeSelector{
							Name:          "foo",
							CanonicalName: "foo",
							Position:      &ast.Position{Line: 2, Col: 1 + len("\t")},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: &ast.Position{Line: 2, Col: 1 + len("\tfoo ")},
							List: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: &ast.Position{Line: 2, Col: 1 + len("\tfoo { ")},
									},
									Type: &ast.AttributeTypeName{
										Name:     "string",
										Type:     attrtype.String,
										Position: &ast.Position{Line: 2, Col: 1 + len("\tfoo { * ")},
									},
								},
							},
							RBrace: &ast.Position{Line: 2, Col: 1 + len("\tfoo { * string ")},
						},
					}, {
						Selector: &ast.BasicAttributeSelector{
							Name:          "bar",
							CanonicalName: "bar",
							Position:      &ast.Position{Line: 3, Col: 1 + len("\t")},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: &ast.Position{Line: 3, Col: 1 + len("\tbar ")},
							List: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: &ast.Position{Line: 3, Col: 1 + len("\tbar { ")},
									},
									Type: &ast.AttributeTypeName{
										Name:     "text",
										Type:     attrtype.Text,
										Position: &ast.Position{Line: 3, Col: 1 + len("\tbar { * ")},
									},
								},
							},
							RBrace: &ast.Position{Line: 3, Col: 1 + len("\tbar { * text ")},
						},
					},
				},
				RParen: &ast.Position{Line: 4, Col: 1},
				Attr:   &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "multiple with prefix",
			in: "attr hx- (\n" +
				"\tfoo { * string }\n" +
				"\tbar { * text }\n" +
				")",
			want: &ast.AttributeDefinition{
				Prefix: &ast.AttributeName{
					Name:          "hx-",
					CanonicalName: "hx-",
					Position:      &ast.Position{Line: 1, Col: 1 + len("attr ")},
				},
				LParen: &ast.Position{Line: 1, Col: 1 + len("attr hx- ")},
				Specs: []*ast.AttributeSpec{
					{
						Selector: &ast.BasicAttributeSelector{
							Name:          "foo",
							CanonicalName: "foo",
							Position:      &ast.Position{Line: 2, Col: 1 + len("\t")},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: &ast.Position{Line: 2, Col: 1 + len("\tfoo ")},
							List: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: &ast.Position{Line: 2, Col: 1 + len("\tfoo { ")},
									},
									Type: &ast.AttributeTypeName{
										Name:     "string",
										Type:     attrtype.String,
										Position: &ast.Position{Line: 2, Col: 1 + len("\tfoo { * ")},
									},
								},
							},
							RBrace: &ast.Position{Line: 2, Col: 1 + len("\tfoo { * string ")},
						},
					}, {
						Selector: &ast.BasicAttributeSelector{
							Name:          "bar",
							CanonicalName: "bar",
							Position:      &ast.Position{Line: 3, Col: 1 + len("\t")},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: &ast.Position{Line: 3, Col: 1 + len("\tbar ")},
							List: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: &ast.Position{Line: 3, Col: 1 + len("\tbar { ")},
									},
									Type: &ast.AttributeTypeName{
										Name:     "text",
										Type:     attrtype.Text,
										Position: &ast.Position{Line: 3, Col: 1 + len("\tbar { * ")},
									},
								},
							},
							RBrace: &ast.Position{Line: 3, Col: 1 + len("\tbar { * text ")},
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

			got := parsetest.ParsesExact(t, c.in, Definition())
			should.Equal(t, got, c.want, parsetest.CmpOpts...)
		})
	}
}

func TestSpec(t *testing.T) {
	t.Parallel()

	in := `foo { * string }`
	want := &ast.AttributeSpec{
		Selector: &ast.BasicAttributeSelector{
			Name:          "foo",
			CanonicalName: "foo",
			Position:      &ast.Position{Line: 1, Col: 1},
		},
		Ruleset: &ast.AttributeRuleset{
			LBrace: &ast.Position{Line: 1, Col: 1 + len("foo ")},
			List: []*ast.AttributeRule{
				{
					Selector: &ast.WildcardElementSelector{
						Asterisk: &ast.Position{Line: 1, Col: 1 + len("foo { ")},
					},
					Type: &ast.AttributeTypeName{
						Name:     "string",
						Type:     attrtype.String,
						Position: &ast.Position{Line: 1, Col: 1 + len("foo { * ")},
					},
				},
			},
			RBrace: &ast.Position{Line: 1, Col: 1 + len("foo { * string ")},
		},
	}

	got := parsetest.ParsesExact(t, in, Spec())
	should.Equal(t, got, want, parsetest.CmpOpts...)
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
			in:   "{ * string }",
			want: &ast.AttributeRuleset{
				LBrace: &ast.Position{Line: 1, Col: 1},
				List: []*ast.AttributeRule{
					{
						Selector: &ast.WildcardElementSelector{
							Asterisk: &ast.Position{Line: 1, Col: 1 + len("{ ")},
						},
						Type: &ast.AttributeTypeName{
							Name:     "string",
							Type:     attrtype.String,
							Position: &ast.Position{Line: 1, Col: 1 + len("{ * ")},
						},
					},
				},
				RBrace: &ast.Position{Line: 1, Col: 1 + len("{ * string ")},
			},
		}, {
			name: "multiple rules on single line",
			in:   "{ * string; foo text }",
			want: &ast.AttributeRuleset{
				LBrace: &ast.Position{Line: 1, Col: 1},
				List: []*ast.AttributeRule{
					{
						Selector: &ast.WildcardElementSelector{
							Asterisk: &ast.Position{Line: 1, Col: 1 + len("{ ")},
						},
						Type: &ast.AttributeTypeName{
							Name:     "string",
							Type:     attrtype.String,
							Position: &ast.Position{Line: 1, Col: 1 + len("{ * ")},
						},
					}, {
						Selector: &ast.ListElementSelector{
							List: []*ast.ElementReference{
								{
									Name: &ast.ElementName{
										Name:          "foo",
										CanonicalName: "foo",
										Position:      &ast.Position{Line: 1, Col: 1 + len("{ * string; ")},
									},
								},
							},
						},
						Type: &ast.AttributeTypeName{
							Name:     "text",
							Type:     attrtype.Text,
							Position: &ast.Position{Line: 1, Col: 1 + len("{ * string; foo ")},
						},
					},
				},
				RBrace: &ast.Position{Line: 1, Col: 1 + len("{ * string; foo text ")},
			},
		}, {
			name: "single rule on multiple lines",
			in: "{\n" +
				"\t* string\n" +
				"}",
			want: &ast.AttributeRuleset{
				LBrace: &ast.Position{Line: 1, Col: 1},
				List: []*ast.AttributeRule{
					{
						Selector: &ast.WildcardElementSelector{
							Asterisk: &ast.Position{Line: 2, Col: 1 + len("\t")},
						},
						Type: &ast.AttributeTypeName{
							Name:     "string",
							Type:     attrtype.String,
							Position: &ast.Position{Line: 2, Col: 1 + len("\t* ")},
						},
					},
				},
				RBrace: &ast.Position{Line: 3, Col: 1},
			},
		}, {
			name: "multiple rules on multiple lines",
			in: "{\n" +
				"\t* string\n" +
				"\tfoo text\n" +
				"}",
			want: &ast.AttributeRuleset{
				LBrace: &ast.Position{Line: 1, Col: 1},
				List: []*ast.AttributeRule{
					{
						Selector: &ast.WildcardElementSelector{
							Asterisk: &ast.Position{Line: 2, Col: 1 + len("\t")},
						},
						Type: &ast.AttributeTypeName{
							Name:     "string",
							Type:     attrtype.String,
							Position: &ast.Position{Line: 2, Col: 1 + len("\t* ")},
						},
					}, {
						Selector: &ast.ListElementSelector{
							List: []*ast.ElementReference{
								{
									Name: &ast.ElementName{
										Name:          "foo",
										CanonicalName: "foo",
										Position:      &ast.Position{Line: 3, Col: 1 + len("\t")},
									},
								},
							},
						},
						Type: &ast.AttributeTypeName{
							Name:     "text",
							Type:     attrtype.Text,
							Position: &ast.Position{Line: 3, Col: 1 + len("\tfoo ")},
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

			got := parsetest.ParsesExact(t, c.in, Ruleset())
			should.Equal(t, got, c.want, parsetest.CmpOpts...)
		})
	}
}

func TestRule(t *testing.T) {
	t.Parallel()

	in := "* string"
	want := &ast.AttributeRule{
		Selector: &ast.WildcardElementSelector{
			Asterisk: &ast.Position{Line: 1, Col: 1},
		},
		Type: &ast.AttributeTypeName{
			Name:     "string",
			Type:     attrtype.String,
			Position: &ast.Position{Line: 1, Col: 1 + len("* ")},
		},
	}

	got := parsetest.ParsesExact(t, in, Rule())
	should.Equal(t, got, want, parsetest.CmpOpts...)
}

func TestSelector(t *testing.T) {
	t.Parallel()

	parsetest.AlsoFulfils(t, Selector(), testBasicSelector)
	parsetest.AlsoFulfils(t, Selector(), testRegexpSelector)
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
			in: "foo",
			want: &ast.BasicAttributeSelector{
				Name:          "foo",
				CanonicalName: "foo",
				Position:      &ast.Position{Line: 1, Col: 1},
			},
		}, {
			in: "foo*",
			want: &ast.BasicAttributeSelector{
				Name:          "foo",
				CanonicalName: "foo",
				Wildcard:      true,
				Position:      &ast.Position{Line: 1, Col: 1},
			},
		}, {
			in: "foo*bar",
			want: &ast.BasicAttributeSelector{
				Name:          "foo*bar",
				CanonicalName: "foo*bar",
				Position:      &ast.Position{Line: 1, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesExact(t, c.in, f)
			should.Equal(t, got, c.want)
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
		Compiled: regexp.MustCompile(`foo\d`),
		RParen:   &ast.Position{Line: 1, Col: 17},
		Regexp:   &ast.Position{Line: 1, Col: 1},
	}

	got := parsetest.ParsesExact(t, in, f)
	should.Equal(t, got, want, cmpopts.IgnoreFields(ast.RegexpAttributeSelector{}, "Compiled"))
}

func TestElementSelector(t *testing.T) {
	t.Parallel()

	parsetest.AlsoFulfils(t, ElementSelector(), testWildcardElementSelector)
	parsetest.AlsoFulfils(t, ElementSelector(), testListElementSelector)
}

func TestWildcardElementSelector(t *testing.T) {
	t.Parallel()
	testWildcardElementSelector(t, WildcardElementSelector())
}

func testWildcardElementSelector(t *testing.T, f parser.Func[*ast.WildcardElementSelector]) {
	in := "*"
	want := &ast.WildcardElementSelector{Asterisk: &ast.Position{Line: 1, Col: 1}}

	got := parsetest.ParsesExact(t, in, f)
	should.Equal(t, got, want)
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
				List: []*ast.ElementReference{
					{
						Name: &ast.ElementName{
							Name:          "foo",
							CanonicalName: "foo",
							Position:      &ast.Position{Line: 1, Col: 1},
						},
					},
				},
			},
		}, {
			name: "multiple",
			in: "foo,\n" +
				"\tbar, baz",
			want: &ast.ListElementSelector{
				List: []*ast.ElementReference{
					{
						Name: &ast.ElementName{
							Name:          "foo",
							CanonicalName: "foo",
							Position:      &ast.Position{Line: 1, Col: 1},
						},
					}, {
						Name: &ast.ElementName{
							Name:          "bar",
							CanonicalName: "bar",
							Position:      &ast.Position{Line: 2, Col: 1 + len("\t")},
						},
					}, {
						Name: &ast.ElementName{
							Name:          "baz",
							CanonicalName: "baz",
							Position:      &ast.Position{Line: 2, Col: 1 + len("\tbar, ")},
						},
					},
				},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesExact(t, c.in, f)
			should.Equal(t, got, c.want, parsetest.CmpOpts...)
		})
	}
}

func TestListElementSelectorItem(t *testing.T) {
	t.Parallel()

	in := "foo"
	want := &ast.ElementReference{
		Name: &ast.ElementName{
			Name:          "foo",
			CanonicalName: "foo",
			Position:      &ast.Position{Line: 1, Col: 1},
		},
	}

	got := parsetest.ParsesExact(t, in, elementReference)
	should.Equal(t, got, want)
}
