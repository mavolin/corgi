package safe

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/should"
)

func TestJSLiteralFromData(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		got, err := JSLiteralFromData([]int{1, 2, 3})
		if should.NoError(t, err) {
			should.Equal(t, got.Get(), "[1,2,3]")
		}
	})
}

func TestConstantScript(t *testing.T) {
	t.Parallel()

	want := constant("console.log('ok')")
	got := ConstantScript(want).Get()
	should.Equal(t, got, string(want))
}

func TestConstantScriptWithData(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		sc, err := ConstantScriptWithData("x1", 5, "console.log(x1)")
		if should.NoError(t, err) {
			should.Equal(t, sc.Get(), "var x1=5;console.log(x1)")
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		t.Run("marshal error", func(t *testing.T) {
			t.Parallel()
			// json.Marshal cannot encode channels
			_, err := ConstantScriptWithData("ch", make(chan int), "console.log(ch)")
			should.False(t, err == nil)
		})
	})
}
