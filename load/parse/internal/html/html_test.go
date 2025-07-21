package html

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestTagName(t *testing.T) {
	t.Parallel()

	tests := []string{
		"div", "input", "myelement",
	}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c, TagName())
			should.Equal(t, c, got)
		})
	}
}

func TestAttributeName(t *testing.T) {
	t.Parallel()

	tests := []string{
		"rel", "type", "my-attribute",
	}

	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesFully(t, c, AttributeName())
			should.Equal(t, c, got)
		})
	}
}
