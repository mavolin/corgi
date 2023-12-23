package safe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustedJS(t *testing.T) {
	expect := "foo"
	actual := TrustedJS(expect)
	assert.Equal(t, expect, actual.Escaped())
}

func TestTrustedJSAttr(t *testing.T) {
	expect := "foo"
	actual := TrustedJSAttr(expect)
	assert.Equal(t, expect, actual.Escaped())
}
