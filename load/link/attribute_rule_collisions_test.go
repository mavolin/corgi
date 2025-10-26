package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/should"
)

func TestLinker_CheckAttributeRuleCollisions(t *testing.T) {
	t.Parallel()

	t.Run("no duplicates", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")

		elem1 := createElementSpec(mainF, &start, "", "div", elemtype.Normal)
		elem2 := createElementSpec(mainF, &start, "", "span", elemtype.Text)

		attrSpec := createBasicAttributeSpec(mainF, &start, "", "foo", nil, attrtype.String)
		attrSpec.AST.Ruleset.List = nil

		rule1 := &ast.AttributeRule{
			Selector: &ast.ListElementSelector{
				List: []*ast.ElementReference{
					createElementReference(mainF, &start, "", elem1.StylizedHTMLName).AST,
				},
			},
		}
		rule1.Type = &ast.AttributeTypeName{
			Name:     attrtype.String.String(),
			Type:     attrtype.String,
			Position: spaceAfter(attrSpec.AST),
		}
		attrSpec.AST.Ruleset.List = append(attrSpec.AST.Ruleset.List, rule1)

		rule2 := &ast.AttributeRule{
			Selector: &ast.ListElementSelector{
				List: []*ast.ElementReference{
					createElementReference(mainF, spaceAfter(attrSpec.AST), "", elem2.StylizedHTMLName).AST,
				},
			},
		}
		rule2.Type = &ast.AttributeTypeName{
			Name:     attrtype.String.String(),
			Type:     attrtype.String,
			Position: spaceAfter(attrSpec.AST),
		}
		attrSpec.AST.Ruleset.List = append(attrSpec.AST.Ruleset.List, rule2)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(),
		})
		if !should.Equal(t, len(ds), 0) {
			t.Log(ds.Short())
		}
	})

	tests := []struct {
		name    string
		setup   func() *file.Package
		message string
	}{
		{
			name: "duplicate wildcard selectors",
			setup: func() *file.Package {
				mainPkg := createPackage("main")
				mainF := createFile(mainPkg, "main.corgi")

				var start ast.Position
				attrSpec := createBasicAttributeSpec(mainF, &start, "", "foo", nil, attrtype.String)
				attrSpec.AST.Ruleset.List = nil

				rule1 := &ast.AttributeRule{
					Selector: &ast.WildcardElementSelector{
						Asterisk: spaceAfter(attrSpec.AST),
					},
				}
				rule1.Type = &ast.AttributeTypeName{
					Name:     attrtype.String.String(),
					Type:     attrtype.String,
					Position: spaceAfter(rule1),
				}
				attrSpec.AST.Ruleset.List = append(attrSpec.AST.Ruleset.List, rule1)

				rule2 := &ast.AttributeRule{
					Selector: &ast.WildcardElementSelector{
						Asterisk: spaceAfter(rule1),
					},
				}
				rule2.Type = &ast.AttributeTypeName{
					Name:     attrtype.String.String(),
					Type:     attrtype.String,
					Position: spaceAfter(rule2),
				}
				attrSpec.AST.Ruleset.List = append(attrSpec.AST.Ruleset.List, rule2)

				return mainPkg
			},
			message: "attribute definition: element selector used multiple times",
		}, {
			name: "two duplicate list element selectors",
			setup: func() *file.Package {
				var start ast.Position
				mainPkg := createPackage("main")
				mainF := createFile(mainPkg, "main.corgi")

				elem := createElementSpec(mainF, &start, "", "div", elemtype.Normal)

				attrSpec := createBasicAttributeSpec(mainF, &start, "", "foo", elem, attrtype.String)

				rule2 := &ast.AttributeRule{
					Selector: &ast.ListElementSelector{
						List: []*ast.ElementReference{
							createElementReference(mainF, spaceAfter(attrSpec.AST), "", elem.StylizedHTMLName).AST,
						},
					},
				}
				rule2.Type = &ast.AttributeTypeName{
					Name:     attrtype.String.String(),
					Type:     attrtype.String,
					Position: spaceAfter(rule2.Selector),
				}
				attrSpec.AST.Ruleset.List = append(attrSpec.AST.Ruleset.List, rule2)

				return mainPkg
			},
			message: "attribute definition: element selector used multiple times",
		}, {
			name: "duplicate list element selector items",
			setup: func() *file.Package {
				var start ast.Position
				mainPkg := createPackage("main")
				mainF := createFile(mainPkg, "main.corgi")

				elem := createElementSpec(mainF, &start, "", "div", elemtype.Normal)

				attrSpec := createBasicAttributeSpec(mainF, &start, "", "foo", nil, attrtype.String)
				attrSpec.AST.Ruleset.List = nil

				selector := &ast.ListElementSelector{
					List: []*ast.ElementReference{
						createElementReference(mainF, spaceAfter(attrSpec.AST), "", elem.StylizedHTMLName).AST,
					},
				}
				selector.List = append(selector.List, createElementReference(mainF, spaceAfter(selector), "", elem.StylizedHTMLName).AST)
				rule := &ast.AttributeRule{Selector: selector}
				rule.Type = &ast.AttributeTypeName{
					Name:     attrtype.String.String(),
					Type:     attrtype.String,
					Position: spaceAfter(rule.Selector),
				}
				attrSpec.AST.Ruleset.List = append(attrSpec.AST.Ruleset.List, rule)

				return mainPkg
			},
			message: "attribute definition: element selector used multiple times",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			pkg := c.setup()

			d := Link(context.Background(), pkg, Options{
				Importer: ImporterFor(),
			})
			t.Log(d.Pretty(diagnostic.PrettyOptions{}))
			if should.Equal(t, len(d), 1) {
				should.Equal(t, d[0].Message, c.message)
			}
		})
	}
}
