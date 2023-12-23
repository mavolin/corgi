package safe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustedHTML(t *testing.T) {
	expect := "foo"
	actual := TrustedHTML(expect)
	assert.Equal(t, expect, actual.Escaped())
}

func TestTrustedPlainAttr(t *testing.T) {
	expect := "foo"
	actual := TrustedPlainAttr(expect)
	assert.Equal(t, expect, actual.Escaped())
}
