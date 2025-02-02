package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestAssignOp(t *testing.T) {
	t.Parallel()

	testCases := []string{"=", "+=", "-=", "|=", "^=", "*=", "/=", "%=", "<<=", ">>=", "&=", "&^="}
	for _, c := range testCases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c, AssignOp())
			assert.Equal(t, c, actual)
		})
	}
}
