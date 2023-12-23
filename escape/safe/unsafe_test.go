package safe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTrustedUnsafe(t *testing.T) {
	t.Parallel()

	expect := "foo"
	actual := TrustedUnsafe(expect)
	assert.Equal(t, expect, actual.Escaped())
}
