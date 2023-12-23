package safe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConcat(t *testing.T) {
	a := HTML{val: "a"}
	b := HTML{val: "b"}

	expect := HTML{val: a.val + b.val}
	actual := Concat(a, b)
	assert.Equal(t, expect, actual)
}
