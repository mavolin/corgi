package safe

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/should"
)

func TestConstantHTML(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		want := constant("Hello &amp; welcome to my site!")
		got := ConstantHTML(want).Get()
		should.Equal(t, got, string(want))
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		t.Run("contains <", func(t *testing.T) {
			t.Parallel()
			should.Panic(t, func() {
				ConstantHTML(`Hello <script>alert('xss')</script>`)
			}, "HTML must not contain literal '<'")
		})
	})
}
