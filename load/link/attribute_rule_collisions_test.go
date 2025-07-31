package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_CheckAttributeRuleCollisions(t *testing.T) {
	t.Parallel()

	t.Run("no duplicates", func(t *testing.T) {
		mainPkg := createPackage("main")
		mainF := createFile(mainPkg, "main.corgi")

		elem1 := createElementSpec(mainF, nil, "", "div", elemtype.Normal)
		elem2 := createElementSpec(mainF, nil, "", "span", elemtype.Text)

		attrSpec := createBasicAttributeSpec(mainF, nil, "", "foo", nil, attrtype.Innocuous)
		attrSpec.AST.Ruleset.List = nil

		rule1 := &ast.AttributeRule{
			Selector: &ast.ListElementSelector{
				List: []*ast.ElementReference{
					createElementReference(mainF, nil, "", elem1.HTMLName()).AST,
				},
			},
		}
		rule1.Type = &ast.AttributeTypeName{
			Name:     attrtype.Innocuous.String(),
			Type:     attrtype.Innocuous,
			Position: spaceAfter(attrSpec.AST),
		}
		attrSpec.AST.Ruleset.List = append(attrSpec.AST.Ruleset.List, rule1)

		rule2 := &ast.AttributeRule{
			Selector: &ast.ListElementSelector{
				List: []*ast.ElementReference{
					createElementReference(mainF, spaceAfter(attrSpec.AST), "", elem2.HTMLName()).AST,
				},
			},
		}
		rule2.Type = &ast.AttributeTypeName{
			Name:     attrtype.Innocuous.String(),
			Type:     attrtype.Innocuous,
			Position: spaceAfter(attrSpec.AST),
		}
		attrSpec.AST.Ruleset.List = append(attrSpec.AST.Ruleset.List, rule2)

		ds := Link(context.Background(), mainPkg, Options{
			Importer: ImporterFor(),
		})
		if !should.Equal(t, 0, len(ds)) {
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

				attrSpec := createBasicAttributeSpec(mainF, nil, "", "foo", nil, attrtype.Innocuous)
				attrSpec.AST.Ruleset.List = nil

				rule1 := &ast.AttributeRule{
					Selector: &ast.WildcardElementSelector{
						Asterisk: spaceAfter(attrSpec.AST),
					},
				}
				rule1.Type = &ast.AttributeTypeName{
					Name:     attrtype.Innocuous.String(),
					Type:     attrtype.Innocuous,
					Position: spaceAfter(attrSpec.AST),
				}
				attrSpec.AST.Ruleset.List = append(attrSpec.AST.Ruleset.List, rule1)

				rule2 := &ast.AttributeRule{
					Selector: &ast.WildcardElementSelector{
						Asterisk: spaceAfter(attrSpec.AST),
					},
				}
				rule2.Type = &ast.AttributeTypeName{
					Name:     attrtype.Innocuous.String(),
					Type:     attrtype.Innocuous,
					Position: spaceAfter(attrSpec.AST),
				}
				attrSpec.AST.Ruleset.List = append(attrSpec.AST.Ruleset.List, rule2)

				return mainPkg
			},
			message: "attribute definition: element selector used multiple times",
		}, {
			name: "two duplicate list element selectors",
			setup: func() *file.Package {
				mainPkg := createPackage("main")
				mainF := createFile(mainPkg, "main.corgi")

				elem := createElementSpec(mainF, nil, "", "div", elemtype.Normal)

				attrSpec := createBasicAttributeSpec(mainF, nil, "", "foo", elem, attrtype.Innocuous)

				rule2 := &ast.AttributeRule{
					Selector: &ast.ListElementSelector{
						List: []*ast.ElementReference{
							createElementReference(mainF, spaceAfter(attrSpec.AST), "", elem.HTMLName()).AST,
						},
					},
				}
				rule2.Type = &ast.AttributeTypeName{
					Name:     attrtype.Innocuous.String(),
					Type:     attrtype.Innocuous,
					Position: spaceAfter(attrSpec.AST),
				}
				attrSpec.AST.Ruleset.List = append(attrSpec.AST.Ruleset.List, rule2)

				return mainPkg
			},
			message: "attribute definition: element selector used multiple times",
		}, {
			name: "duplicate list element selector items",
			setup: func() *file.Package {
				mainPkg := createPackage("main")
				mainF := createFile(mainPkg, "main.corgi")

				elem := createElementSpec(mainF, nil, "", "div", elemtype.Normal)

				attrSpec := createBasicAttributeSpec(mainF, nil, "", "foo", nil, attrtype.Innocuous)
				attrSpec.AST.Ruleset.List = nil

				rule := &ast.AttributeRule{
					Selector: &ast.ListElementSelector{
						List: []*ast.ElementReference{
							createElementReference(mainF, spaceAfter(attrSpec.AST), "", elem.HTMLName()).AST,
							createElementReference(mainF, spaceAfter(attrSpec.AST), "", elem.HTMLName()).AST,
						},
					},
				}
				rule.Type = &ast.AttributeTypeName{
					Name:     attrtype.Innocuous.String(),
					Type:     attrtype.Innocuous,
					Position: spaceAfter(attrSpec.AST),
				}
				attrSpec.AST.Ruleset.List = append(attrSpec.AST.Ruleset.List, rule)

				return mainPkg
			},
			message: "attribute definition: element selector used multiple times",
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			pkg := c.setup()

			ds := Link(context.Background(), pkg, Options{
				Importer: ImporterFor(),
			})
			if should.Equal(t, 1, len(ds)) {
				if !should.Equal(t, c.message, ds[0].Message) {
					t.Log(ds[0].Short())
				}
			} else {
				t.Log(ds.Short())
			}
		})
	}
}
