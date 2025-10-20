package safe

import (
	"fmt"
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestConstantIdentifier(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		id := ConstantIdentifier("foo_bar-1")
		should.Equal(t, id.Get(), "foo_bar-1")
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		testCases := []constant{"1abc", "woof bark", "woof.bark"}
		for _, c := range testCases {
			t.Run(string(c), func(t *testing.T) {
				t.Parallel()
				should.Panic(t, func() {
					ConstantIdentifier(c)
				}, fmt.Sprintf("invalid identifier %q", c))
			})
		}
	})
}

func TestPrefixedIdentifier(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		id, err := PrefixedIdentifier("pre-", "value_1")
		if should.NoError(t, err) {
			should.Equal(t, id.Get(), "pre-value_1")
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			prefix  constant
			value   string
			wantErr string
		}{
			{"1bad", "ok", `invalid identifier prefix "1bad"`},
			{"pre-", "bad:val", `invalid identifier value "bad:val"`},
		}
		for _, c := range testCases {
			t.Run(string(c.prefix)+"|"+c.value, func(t *testing.T) {
				t.Parallel()
				_, err := PrefixedIdentifier(c.prefix, c.value)
				should.Error(t, err, c.wantErr)
			})
		}
	})
}

func TestMustPrefixedIdentifier(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()
		id := MustPrefixedIdentifier("x-", "y")
		should.Equal(t, id.Get(), "x-y")
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name    string
			prefix  constant
			value   string
			wantErr string
		}{
			{"invalid prefix", "1bad", "ok", `invalid identifier prefix "1bad"`},
			{"invalid value", "pre-", "bad:val", `invalid identifier value "bad:val"`},
		}
		for _, c := range testCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()
				should.Panic(t, func() {
					MustPrefixedIdentifier(c.prefix, c.value)
				}, c.wantErr)
			})
		}
	})
}
