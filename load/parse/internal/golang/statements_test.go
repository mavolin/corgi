package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/should"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestAssignOp(t *testing.T) {
	t.Parallel()

	tests := []string{"=", "+=", "-=", "|=", "^=", "*=", "/=", "%=", "<<=", ">>=", "&=", "&^="}
	for _, c := range tests {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			got := parsetest.ParsesExact(t, c, AssignOp())
			should.Equal(t, got, c)
		})
	}
}
