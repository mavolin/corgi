package safe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustedCSSValue(t *testing.T) {
	expect := "foo"
	actual := TrustedCSSValue(expect)
	assert.Equal(t, expect, actual.Escaped())
}

func TestTrustedCSSValueAttr(t *testing.T) {
	expect := "foo"
	actual := TrustedCSSValueAttr(expect)
	assert.Equal(t, expect, actual.Escaped())
}
