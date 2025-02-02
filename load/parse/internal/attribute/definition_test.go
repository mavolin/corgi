package attribute

import (
	"regexp"
	"testing"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestDefinition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.AttributeDefinition
	}{
		{
			name: "single",
			in:   `attr foo { * innocuous }`,
			expect: &ast.AttributeDefinition{
				Specs: []*ast.AttributeSpec{
					{
						Name: &ast.BasicAttributeSelector{
							Name:     "foo",
							Position: ast.Position{Line: 1, Col: 6},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: ast.Position{Line: 1, Col: 10},
							Rules: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: ast.Position{Line: 1, Col: 12},
									},
									Type: &ast.AttributeTypeName{
										Name:     "innocuous",
										Type:     attrtype.Innocuous,
										Position: ast.Position{Line: 1, Col: 14},
									},
								},
							},
							RBrace: &ast.Position{Line: 1, Col: 24},
						},
					},
				},
				Position: ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "single with prefix",
			in:   `attr hx- foo { * innocuous }`,
			expect: &ast.AttributeDefinition{
				Prefix: &ast.AttributeName{
					Name:     "hx-",
					Position: ast.Position{Line: 1, Col: 6},
				},
				Specs: []*ast.AttributeSpec{
					{
						Name: &ast.BasicAttributeSelector{
							Name:     "foo",
							Position: ast.Position{Line: 1, Col: 10},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: ast.Position{Line: 1, Col: 14},
							Rules: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: ast.Position{Line: 1, Col: 16},
									},
									Type: &ast.AttributeTypeName{
										Name:     "innocuous",
										Type:     attrtype.Innocuous,
										Position: ast.Position{Line: 1, Col: 18},
									},
								},
							},
							RBrace: &ast.Position{Line: 1, Col: 28},
						},
					},
				},
				Position: ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "multiple",
			in: "attr (\n" +
				"\tfoo { * innocuous }\n" +
				"\tbar { * text }\n" +
				")",
			expect: &ast.AttributeDefinition{
				LParen: &ast.Position{Line: 1, Col: 6},
				Specs: []*ast.AttributeSpec{
					{
						Name: &ast.BasicAttributeSelector{
							Name:     "foo",
							Position: ast.Position{Line: 2, Col: 2},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: ast.Position{Line: 2, Col: 6},
							Rules: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: ast.Position{Line: 2, Col: 8},
									},
									Type: &ast.AttributeTypeName{
										Name:     "innocuous",
										Type:     attrtype.Innocuous,
										Position: ast.Position{Line: 2, Col: 10},
									},
								},
							},
							RBrace: &ast.Position{Line: 2, Col: 20},
						},
					}, {
						Name: &ast.BasicAttributeSelector{
							Name:     "bar",
							Position: ast.Position{Line: 3, Col: 2},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: ast.Position{Line: 3, Col: 6},
							Rules: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: ast.Position{Line: 3, Col: 8},
									},
									Type: &ast.AttributeTypeName{
										Name:     "text",
										Type:     attrtype.Text,
										Position: ast.Position{Line: 3, Col: 10},
									},
								},
							},
							RBrace: &ast.Position{Line: 3, Col: 15},
						},
					},
				},
				RParen:   &ast.Position{Line: 4, Col: 1},
				Position: ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "multiple with prefix",
			in: "attr hx- (\n" +
				"\tfoo { * innocuous }\n" +
				"\tbar { * text }\n" +
				")",
			expect: &ast.AttributeDefinition{
				Prefix: &ast.AttributeName{
					Name:     "hx-",
					Position: ast.Position{Line: 1, Col: 6},
				},
				LParen: &ast.Position{Line: 1, Col: 10},
				Specs: []*ast.AttributeSpec{
					{
						Name: &ast.BasicAttributeSelector{
							Name:     "foo",
							Position: ast.Position{Line: 2, Col: 2},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: ast.Position{Line: 2, Col: 6},
							Rules: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: ast.Position{Line: 2, Col: 8},
									},
									Type: &ast.AttributeTypeName{
										Name:     "innocuous",
										Type:     attrtype.Innocuous,
										Position: ast.Position{Line: 2, Col: 10},
									},
								},
							},
							RBrace: &ast.Position{Line: 2, Col: 20},
						},
					}, {
						Name: &ast.BasicAttributeSelector{
							Name:     "bar",
							Position: ast.Position{Line: 3, Col: 2},
						},
						Ruleset: &ast.AttributeRuleset{
							LBrace: ast.Position{Line: 3, Col: 6},
							Rules: []*ast.AttributeRule{
								{
									Selector: &ast.WildcardElementSelector{
										Asterisk: ast.Position{Line: 3, Col: 8},
									},
									Type: &ast.AttributeTypeName{
										Name:     "text",
										Type:     attrtype.Text,
										Position: ast.Position{Line: 3, Col: 10},
									},
								},
							},
							RBrace: &ast.Position{Line: 3, Col: 15},
						},
					},
				},
				RParen:   &ast.Position{Line: 4, Col: 1},
				Position: ast.Position{Line: 1, Col: 1},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in+";", Definition())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestSpec(t *testing.T) {
	t.Parallel()

	in := `foo { * innocuous }`
	expect := &ast.AttributeSpec{
		Name: &ast.BasicAttributeSelector{
			Name:     "foo",
			Position: ast.Position{Line: 1, Col: 1},
		},
		Ruleset: &ast.AttributeRuleset{
			LBrace: ast.Position{Line: 1, Col: 5},
			Rules: []*ast.AttributeRule{
				{
					Selector: &ast.WildcardElementSelector{
						Asterisk: ast.Position{Line: 1, Col: 7},
					},
					Type: &ast.AttributeTypeName{
						Name:     "innocuous",
						Type:     attrtype.Innocuous,
						Position: ast.Position{Line: 1, Col: 9},
					},
				},
			},
			RBrace: &ast.Position{Line: 1, Col: 19},
		},
	}

	actual := testutil.ParsesFully(t, in, Spec())
	assert.Equal(t, expect, actual)
}

func TestRuleset(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.AttributeRuleset
	}{
		{
			name: "single rule on single line",
			in:   "{ * innocuous }",
			expect: &ast.AttributeRuleset{
				LBrace: ast.Position{Line: 1, Col: 1},
				Rules: []*ast.AttributeRule{
					{
						Selector: &ast.WildcardElementSelector{
							Asterisk: ast.Position{Line: 1, Col: 3},
						},
						Type: &ast.AttributeTypeName{
							Name:     "innocuous",
							Type:     attrtype.Innocuous,
							Position: ast.Position{Line: 1, Col: 5},
						},
					},
				},
				RBrace: &ast.Position{Line: 1, Col: 15},
			},
		}, {
			name: "multiple rules on single line",
			in:   "{ * innocuous; foo text }",
			expect: &ast.AttributeRuleset{
				LBrace: ast.Position{Line: 1, Col: 1},
				Rules: []*ast.AttributeRule{
					{
						Selector: &ast.WildcardElementSelector{
							Asterisk: ast.Position{Line: 1, Col: 3},
						},
						Type: &ast.AttributeTypeName{
							Name:     "innocuous",
							Type:     attrtype.Innocuous,
							Position: ast.Position{Line: 1, Col: 5},
						},
					}, {
						Selector: &ast.ListElementSelector{
							Elements: []*ast.ListElementSelectorItem{
								{
									Name:     "foo",
									Position: ast.Position{Line: 1, Col: 16},
								},
							},
						},
						Type: &ast.AttributeTypeName{
							Name:     "text",
							Type:     attrtype.Text,
							Position: ast.Position{Line: 1, Col: 20},
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
			expect: &ast.AttributeRuleset{
				LBrace: ast.Position{Line: 1, Col: 1},
				Rules: []*ast.AttributeRule{
					{
						Selector: &ast.WildcardElementSelector{
							Asterisk: ast.Position{Line: 2, Col: 2},
						},
						Type: &ast.AttributeTypeName{
							Name:     "innocuous",
							Type:     attrtype.Innocuous,
							Position: ast.Position{Line: 2, Col: 4},
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
			expect: &ast.AttributeRuleset{
				LBrace: ast.Position{Line: 1, Col: 1},
				Rules: []*ast.AttributeRule{
					{
						Selector: &ast.WildcardElementSelector{
							Asterisk: ast.Position{Line: 2, Col: 2},
						},
						Type: &ast.AttributeTypeName{
							Name:     "innocuous",
							Type:     attrtype.Innocuous,
							Position: ast.Position{Line: 2, Col: 4},
						},
					}, {
						Selector: &ast.ListElementSelector{
							Elements: []*ast.ListElementSelectorItem{
								{
									Name:     "foo",
									Position: ast.Position{Line: 3, Col: 2},
								},
							},
						},
						Type: &ast.AttributeTypeName{
							Name:     "text",
							Type:     attrtype.Text,
							Position: ast.Position{Line: 3, Col: 6},
						},
					},
				},
				RBrace: &ast.Position{Line: 4, Col: 1},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in, Ruleset())
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestRule(t *testing.T) {
	t.Parallel()

	in := "* innocuous"
	expect := &ast.AttributeRule{
		Selector: &ast.WildcardElementSelector{
			Asterisk: ast.Position{Line: 1, Col: 1},
		},
		Type: &ast.AttributeTypeName{
			Name:     "innocuous",
			Type:     attrtype.Innocuous,
			Position: ast.Position{Line: 1, Col: 3},
		},
	}

	actual := testutil.ParsesFully(t, in, Rule())
	assert.Equal(t, expect, actual)
}

func TestSelector(t *testing.T) {
	t.Parallel()

	testutil.AssertAlsoFulfils(t, Selector(), testBasicSelector)
	testutil.AssertAlsoFulfils(t, Selector(), testRegexpSelector)
}

func TestBasicSelector(t *testing.T) {
	t.Parallel()
	testBasicSelector(t, BasicSelector())
}

func testBasicSelector(t *testing.T, f parser.Func[*ast.BasicAttributeSelector]) {
	testCases := []struct {
		in     string
		expect *ast.BasicAttributeSelector
	}{
		{
			in:     "foo",
			expect: &ast.BasicAttributeSelector{Name: "foo", Position: ast.Position{Line: 1, Col: 1}},
		}, {
			in:     "foo*",
			expect: &ast.BasicAttributeSelector{Name: "foo", Wildcard: true, Position: ast.Position{Line: 1, Col: 1}},
		}, {
			in:     "foo*bar",
			expect: &ast.BasicAttributeSelector{Name: "foo*bar", Position: ast.Position{Line: 1, Col: 1}},
		},
	}

	for _, c := range testCases {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestRegexpSelector(t *testing.T) {
	t.Parallel()
	testRegexpSelector(t, RegexpSelector())
}

func testRegexpSelector(t *testing.T, f parser.Func[*ast.RegexpAttributeSelector]) {
	in := `'regexp("foo\\d")`
	expect := &ast.RegexpAttributeSelector{
		LParen: &ast.Position{Line: 1, Col: 8},
		Raw: &ast.StaticString{
			Open:     ast.Position{Line: 1, Col: 9},
			Quote:    '"',
			Contents: `foo\\d`,
			Close:    &ast.Position{Line: 1, Col: 16},
		},
		Regexp:   regexp.MustCompile("foo\\d"),
		RParen:   &ast.Position{Line: 1, Col: 17},
		Position: ast.Position{Line: 1, Col: 1},
	}

	actual := testutil.ParsesFully(t, in, f)
	assert.Equal(t, expect, actual)
}

func TestElementSelector(t *testing.T) {
	t.Parallel()

	testutil.AssertAlsoFulfils(t, ElementSelector(), testWildcardElementSelector)
	testutil.AssertAlsoFulfils(t, ElementSelector(), testListElementSelector)
}

func TestWildcardElementSelector(t *testing.T) {
	t.Parallel()
	testWildcardElementSelector(t, WildcardElementSelector())
}

func testWildcardElementSelector(t *testing.T, f parser.Func[*ast.WildcardElementSelector]) {
	in := "*"
	expect := &ast.WildcardElementSelector{Asterisk: ast.Position{Line: 1, Col: 1}}

	actual := testutil.ParsesFully(t, in, f)
	assert.Equal(t, expect, actual)
}

func TestListElementSelector(t *testing.T) {
	t.Parallel()
	testListElementSelector(t, ListElementSelector())
}

func testListElementSelector(t *testing.T, f parser.Func[*ast.ListElementSelector]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.ListElementSelector
	}{
		{
			name: "single",
			in:   "foo",
			expect: &ast.ListElementSelector{
				Elements: []*ast.ListElementSelectorItem{
					{Name: "foo", Position: ast.Position{Line: 1, Col: 1}},
				},
			},
		}, {
			name: "multiple",
			in: "foo,\n" +
				"\tbar, baz",
			expect: &ast.ListElementSelector{
				Elements: []*ast.ListElementSelectorItem{
					{Name: "foo", Position: ast.Position{Line: 1, Col: 1}},
					{Name: "bar", Position: ast.Position{Line: 2, Col: 2}},
					{Name: "baz", Position: ast.Position{Line: 2, Col: 7}},
				},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}

func TestListElementSelectorItem(t *testing.T) {
	t.Parallel()

	in := "foo"
	expect := &ast.ListElementSelectorItem{Name: "foo", Position: ast.Position{Line: 1, Col: 1}}

	actual := testutil.ParsesFully(t, in, ListElementSelectorItem())
	assert.Equal(t, expect, actual)
}
