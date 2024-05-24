package elemtype

import (
	"fmt"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

func TestHTML5(t *testing.T) {
	t.Parallel()

	spec, err := loadSpec()
	require.NoError(t, err)

	for _, e := range spec.elems {
		t.Run(e.name, func(t *testing.T) {
			t.Parallel()

			switch e.children {
			case "empty":
				// weird special case based on semantics
				if e.name == "template" {
					assert.Equal(t, HTML, HTML5(e.name), "hard code: expected template to be HTML", HTML5(e.name))
					return
				}

				typ := HTML5(e.name)
				switch typ {
				case Void, Nothing:
				default:
					assert.Failf(t, "type mismatch", "listed as empty (i.e. Void or Nothing), but got %s", typ)
				}
			case "transparent":
				assert.Equal(t, HTML, HTML5(e.name), "listed as transparent (i.e. HTML), but got %s", HTML5(e.name))
			case "text":
				assert.Equal(t, Text, HTML5(e.name), "listed as text, but got %s", HTML5(e.name))
			case "phrasing", "flow":
				assert.Equal(t, HTML, HTML5(e.name), "listed as phrasing or flow (i.e. HTML), but got %s", HTML5(e.name))
			default:
				assert.True(t, HTML5(e.name).IsValid(), "listed in spec, but not handled")
			}
		})
	}
}

type (
	elem struct {
		name     string
		children string // i.e. type
	}

	spec struct {
		elems []elem
	}
)

const specURL = "https://html.spec.whatwg.org/multipage/indices.html"

func loadSpec() (*spec, error) {
	specDoc, err := htmlquery.LoadURL(specURL)
	if err != nil {
		return nil, fmt.Errorf("could not load spec: %w", err)
	}

	var spec spec
	spec.elems, err = extractElems(specDoc)
	if err != nil {
		return nil, err
	}

	return &spec, nil
}

const h3ID = "elements-3"

func extractElems(specDoc *html.Node) ([]elem, error) {
	table := htmlquery.FindOne(specDoc, fmt.Sprintf("//h3[@id=%q]/following-sibling::table", h3ID))
	if table == nil {
		return nil, fmt.Errorf("could not find attribute table with id %q", h3ID)
	}

	rows := htmlquery.Find(table, "//tbody/tr")
	if len(rows) == 0 {
		return nil, fmt.Errorf("could not find any elements")
	}

	elems := make([]elem, 0, len(rows))
	for _, row := range rows {
		innerTxt := strings.TrimSpace(htmlquery.InnerText(htmlquery.FindOne(row, "//th")))
		switch innerTxt {
		case "MathML math", "SVG svg", "autonomous custom elements":
			continue
		}
		elems = append(elems, elem{
			name:     htmlquery.InnerText(htmlquery.FindOne(row, "//th/code")),
			children: strings.TrimSpace(htmlquery.InnerText(htmlquery.FindOne(row, "//td[4]"))),
		})
	}

	return elems, nil
}
