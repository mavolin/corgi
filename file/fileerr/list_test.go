package fileerr

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAs(t *testing.T) {
	a := &Error{}
	b := &Error{}
	c := &Error{}

	z := errors.New("z")
	y := errors.New("y")
	x := errors.Join(errors.New("x.1"), errors.New("x.2"))

	test := errors.Join(a, b, errors.Join(z, c), y, x, nil)

	actualFerrs, actualErrs := As(test)
	assert.Len(t, actualFerrs, 3)
	assert.Len(t, actualErrs, 3, "errors split up too much or included nil")

	containsExact(t, actualFerrs, a)
	containsExact(t, actualFerrs, b)
	containsExact(t, actualFerrs, c)

	containsExact(t, actualErrs, z)
	containsExact(t, actualErrs, y)
	containsExact(t, actualErrs, x)
}

func containsExact[T comparable](t *testing.T, s []T, e T) {
	t.Helper()

	for _, v := range s {
		if v == e {
			return
		}
	}

	assert.Fail(t, "slice does not contain element", "slice: %v, element: %v", s, e)
}
