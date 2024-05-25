package attrtype

import (
	"fmt"
	"strings"
	"testing"

	"github.com/antchfx/htmlquery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/html"
)

// partial test for the core and additional attributes, but not for extensions
func TestHTMX(t *testing.T) {
	t.Parallel()

	spec, err := loadHTMXSpec()
	require.NoError(t, err)

	for _, attr := range spec.attrs {
		switch attr.name {
		case "hx-on*":
			// handled separately below
			continue
		case "hx-sse", "hx-ws", "hx-validate":
			// deprecated before latest major version of corgi
			continue
		}

		t.Run(attr.name, func(t *testing.T) {
			t.Parallel()

			switch {
			case strings.Contains(attr.description, "URL"):
				assert.Equal(t, URL, HTMX("", attr.name), "description contains 'URL': expected URL")
			case strings.Contains(attr.description, "JSON"):
				assert.Equal(t, JS, HTMX("", attr.name), "description contains 'JSON': expected JS")
			default:
				assert.True(t, HTMX("", attr.name).IsValid(), "attr listed in spec, but not handled")
			}
		})
	}

	t.Run("hx-on:*", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, JS, HTMX("", "hx-on:click"), "hard code: expected hx-on:click to be JS")
	})

}

type (
	htmxAttr struct {
		name        string
		description string
	}

	htmxSpec struct {
		attrs []htmxAttr
	}
)

const htmxSpecURL = "https://htmx.org/reference"

func loadHTMXSpec() (*htmxSpec, error) {
	specDoc, err := htmlquery.LoadURL(htmxSpecURL)
	if err != nil {
		return nil, fmt.Errorf("could not load htmx spec: %w", err)
	}

	var spec htmxSpec
	spec.attrs, err = extractHTMXAttrs("attributes", specDoc)
	if err != nil {
		return nil, err
	}

	attrs2, err := extractHTMXAttrs("attributes-additional", specDoc)
	if err != nil {
		return nil, err
	}
	spec.attrs = append(spec.attrs, attrs2...)

	return &spec, nil
}

func extractHTMXAttrs(h2ID string, specDoc *html.Node) ([]htmxAttr, error) {
	table := htmlquery.FindOne(specDoc, fmt.Sprintf("//h2[@id=%q]/following-sibling::div/table", h2ID))
	if table == nil {
		return nil, fmt.Errorf("could not find attribute table with h2 id %q", h2ID)
	}

	rows := htmlquery.Find(table, "//tbody/tr")
	if len(rows) == 0 {
		return nil, fmt.Errorf("could not find any attributes")
	}

	attrs := make([]htmxAttr, 0, len(rows))
	for _, row := range rows {
		name := htmlquery.InnerText(htmlquery.FindOne(row, "//td[1]/a/code/text()"))

		attrs = append(attrs, htmxAttr{
			name:        name,
			description: strings.TrimSpace(htmlquery.InnerText(htmlquery.FindOne(row, "//td[2]"))),
		})
	}

	return attrs, nil
}
