package safe

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/should"
)

func TestConstantURL(t *testing.T) {
	t.Parallel()

	want := constant("https://example.com/path?query=1#fragment")
	got := ConstantURL(want).Get()
	should.Equal(t, got, string(want))
}

func TestConstantResourceURL(t *testing.T) {
	t.Parallel()

	want := constant("https://example.com/resource.js")
	got := ConstantResourceURL(want).Get()
	should.Equal(t, got, string(want))
}
