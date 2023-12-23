package safe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustedSrcsetAttr(t *testing.T) {
	expect := "foo"
	actual := TrustedSrcsetAttr(expect)
	assert.Equal(t, expect, actual.Escaped())
}
