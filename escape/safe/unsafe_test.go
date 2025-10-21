package safe

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/should"
)

func TestConstantUnsafe(t *testing.T) {
	t.Parallel()

	wantAttr := constant("woof")
	wantValue := constant("bark")

	got := ConstantUnsafe(wantAttr, wantValue)
	should.Equal(t, got.Attr(), string(wantAttr))
	should.Equal(t, got.Get(), string(wantValue))
}

func TestUnsafeTrue(t *testing.T) {
	t.Parallel()

	wantAttr := constant("woof")

	got := UnsafeTrue(wantAttr)
	should.Equal(t, got.Attr(), string(wantAttr))
	should.True(t, got.Get())
}

func TestUnsafeFalse(t *testing.T) {
	t.Parallel()

	wantAttr := constant("woof")

	got := UnsafeFalse(wantAttr)
	should.Equal(t, got.Attr(), string(wantAttr))
	should.False(t, got.Get())
}
