package html

import (
	"testing"

	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestTagName(t *testing.T) {
	t.Parallel()

	testCases := []string{
		"div", "input", "myelement",
	}

	for _, c := range testCases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c, TagName())
			assert.Equal(t, c, actual)
		})
	}
}

func TestAttributeName(t *testing.T) {
	t.Parallel()

	testCases := []string{
		"rel", "type", "my-attribute",
	}

	for _, c := range testCases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c, AttributeName())
			assert.Equal(t, c, actual)
		})
	}
}
