package attrtype

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCombine(t *testing.T) {
	t.Parallel()

	var called bool
	expectElement, expectAttr := "el", "attr"

	expect := URL
	actual := Combine(func(element, attr string) Type {
		assert.Equal(t, expectElement, element)
		assert.Equal(t, expectAttr, attr)

		called = true
		return Unknown
	}, func(element, attr string) Type {
		assert.Equal(t, expectElement, element)
		assert.Equal(t, expectAttr, attr)

		return expect
	})(expectElement, expectAttr)
	assert.True(t, called)
	assert.Equal(t, expect, actual)
}
