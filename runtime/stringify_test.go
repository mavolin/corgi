package runtime

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestStringify(t *testing.T) {
	t.Parallel()

	testStringify(t, "hello", "hello")
	testStringify(t, "", "")

	testStringify(t, 0, "0")
	testStringify(t, 42, "42")
	testStringify(t, -42, "-42")
	testStringify(t, int8(8), "8")
	testStringify(t, int8(-8), "-8")
	testStringify(t, int16(16), "16")
	testStringify(t, int16(-16), "-16")
	testStringify(t, int32(32), "32")
	testStringify(t, int32(-32), "-32")
	testStringify(t, int64(64), "64")
	testStringify(t, int64(-64), "-64")

	testStringify(t, uint(0), "0")
	testStringify(t, uint(42), "42")
	testStringify(t, uint8(8), "8")
	testStringify(t, uint16(16), "16")
	testStringify(t, uint32(32), "32")
	testStringify(t, uint64(64), "64")

	testStringify(t, float32(0), "0")
	testStringify(t, float32(3.14), "3.14")
	testStringify(t, float32(-3.14), "-3.14")
	testStringify(t, float32(42), "42")
	testStringify(t, float64(0), "0")
	testStringify(t, float64(2.71828), "2.71828")
	testStringify(t, float64(-2.71828), "-2.71828")
	testStringify(t, float64(42), "42")

	testStringify(t, (*string)(nil), "")
	testStringify(t, pointerTo("woof"), "woof")

	testStringify(t, (*int)(nil), "")
	testStringify(t, pointerTo(42), "42")
	testStringify(t, (*int8)(nil), "")
	testStringify(t, pointerTo(int8(8)), "8")
	testStringify(t, (*int16)(nil), "")
	testStringify(t, pointerTo(int16(16)), "16")
	testStringify(t, (*int32)(nil), "")
	testStringify(t, pointerTo(int32(32)), "32")
	testStringify(t, (*int64)(nil), "")
	testStringify(t, pointerTo(int64(64)), "64")

	testStringify(t, (*uint)(nil), "")
	testStringify(t, pointerTo(uint(42)), "42")
	testStringify(t, (*uint8)(nil), "")
	testStringify(t, pointerTo(uint8(8)), "8")
	testStringify(t, (*uint16)(nil), "")
	testStringify(t, pointerTo(uint16(16)), "16")
	testStringify(t, (*uint32)(nil), "")
	testStringify(t, pointerTo(uint32(32)), "32")
	testStringify(t, (*uint64)(nil), "")
	testStringify(t, pointerTo(uint64(64)), "64")

	testStringify(t, (*float32)(nil), "")
	testStringify(t, pointerTo(float32(3.14)), "3.14")
	testStringify(t, (*float64)(nil), "")
	testStringify(t, pointerTo(float64(2.71828)), "2.71828")
}

// testStringify is a generic helper that runs a subtest for Stringify with a given input and expected output
func testStringify[T Printable](t *testing.T, in T, want string) {
	t.Helper()

	name := fmt.Sprintf("%T/", in)
	rv := reflect.ValueOf(in)
	if rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			name += "nil"
		} else {
			name += fmt.Sprintf("%#[1]v", rv.Elem().Interface())
		}
	} else {
		name += fmt.Sprintf("%#[1]v", in)
	}

	t.Run(name, func(t *testing.T) {
		t.Parallel()

		got := Stringify(in)
		should.Equal(t, got, want)
	})
}

func pointerTo[T any](t T) *T {
	return &t
}
