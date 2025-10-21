package parsetest

import (
	"reflect"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

type options struct {
	inline bool
	errors []string
}

type Option func(o *options)

func Inline() Option {
	return func(o *options) { o.inline = true }
}

func WantErrors(wantMessages ...string) Option {
	return func(o *options) { o.errors = append(o.errors, wantMessages...) }
}

func applyOptions(opts ...Option) options {
	var o options
	for _, opt := range opts {
		opt(&o)
	}
	return o
}

// ParsesExact asserts that f matches the entire in string, but would not
// consume any extra space separated in.
func ParsesExact[T any](t *testing.T, in string, f parser.Func[T], opts ...Option) T {
	t.Helper()
	var v T
	ok := t.Run("followed by EOF", func(t *testing.T) {
		v = ParsesUntilExtra(t, in, " ", f, opts...)
	})
	if !ok {
		return v
	}

	t.Run("followed by number", func(t *testing.T) {
		v2 := ParsesUntilExtra(t, in, " 1other stuff", f, opts...)
		should.Equal(t, v2, v, cmpopts.IgnoreTypes((*regexp.Regexp)(nil))) // EOF and number results differ
	})
	t.Run("followed by identifier", func(t *testing.T) {
		v2 := ParsesUntilExtra(t, in, " otherStuff", f, opts...)
		should.Equal(t, v2, v, cmpopts.IgnoreTypes((*regexp.Regexp)(nil))) // EOF and identifier results differ
	})
	return v
}

// ParsesUntilEOS asserts that f matches the entire in string, and would
// stop at the EOS.
func ParsesUntilEOS[T any](t *testing.T, in string, f parser.Func[T], opts ...Option) T {
	t.Helper()
	var v T
	ok := t.Run("followed by EOF", func(t *testing.T) {
		v = ParsesUntilExtra(t, in, " ", f, opts...)
	})
	if !ok {
		return v
	}

	t.Run("followed by semicolon", func(t *testing.T) {
		v2 := ParsesUntilExtra(t, in, "; 1other stuff", f, opts...)
		should.Equal(t, v2, v, cmpopts.IgnoreTypes((*regexp.Regexp)(nil))) // EOF and semicolon results differ
	})
	t.Run("followed by newline", func(t *testing.T) {
		v2 := ParsesUntilExtra(t, in, "\n1other stuff", f, opts...)
		should.Equal(t, v2, v, cmpopts.IgnoreTypes((*regexp.Regexp)(nil))) // EOF and newline results differ
	})
	t.Run("followed by line comment", func(t *testing.T) {
		v2 := ParsesUntilExtra(t, in, "// comment", f, opts...)
		should.Equal(t, v2, v, cmpopts.IgnoreTypes((*regexp.Regexp)(nil))) // EOF and line comment results differ
	})
	t.Run("followed by block comment", func(t *testing.T) {
		v2 := ParsesUntilExtra(t, in, "/* comment */; 1other stuff", f, opts...)
		should.Equal(t, v2, v, cmpopts.IgnoreTypes((*regexp.Regexp)(nil))) // EOF and block comment results differ
	})
	return v
}

// ParsesUntilComma asserts that f matches the entire in string, and would
// stop at a comma or EOF.
func ParsesUntilComma[T any](t *testing.T, in string, f parser.Func[T], opts ...Option) T {
	t.Helper()
	var v T
	ok := t.Run("followed by EOF", func(t *testing.T) {
		v = ParsesUntilExtra(t, in, " ", f, opts...)
	})
	if !ok {
		return v
	}

	t.Run("followed by comma", func(t *testing.T) {
		v2 := ParsesUntilExtra(t, in, ", 1other stuff", f, opts...)
		should.Equal(t, v2, v, cmpopts.IgnoreTypes((*regexp.Regexp)(nil))) // EOF and semicolon results differ
	})
	t.Run("followed by block comment", func(t *testing.T) {
		v2 := ParsesUntilExtra(t, in, "/* comment */, 1other stuff", f, opts...)
		should.Equal(t, v2, v, cmpopts.IgnoreTypes((*regexp.Regexp)(nil))) // EOF and block comment results differ
	})
	return v
}

// ParsesUntilBody asserts that f matches the entire in string, and would
// stop at a body or EOF.
func ParsesUntilBody[T any](t *testing.T, in string, f parser.Func[T], opts ...Option) T {
	t.Helper()
	var v T
	ok := t.Run("followed by EOF", func(t *testing.T) {
		v = ParsesUntilExtra(t, in, " ", f, opts...)
	})
	if !ok {
		return v
	}

	t.Run("followed by scope", func(t *testing.T) {
		v2 := ParsesUntilExtra(t, in, " {\n\tdiv\n}", f, opts...)
		should.Equal(t, v2, v, cmpopts.IgnoreTypes((*regexp.Regexp)(nil))) // EOF and scope results differ

		t.Run("followed by block comment", func(t *testing.T) {
			v2 := ParsesUntilExtra(t, in, " /* comment */{\n\tdiv\n}", f, opts...)
			should.Equal(t, v2, v, cmpopts.IgnoreTypes((*regexp.Regexp)(nil))) // EOF and scope results differ
		})
	})
	t.Run("followed by bracket text", func(t *testing.T) {
		v2 := ParsesUntilExtra(t, in, " [\n\twoof\n]", f, opts...)
		should.Equal(t, v2, v, cmpopts.IgnoreTypes((*regexp.Regexp)(nil))) // EOF and bracket text results differ

		t.Run("followed by block comment", func(t *testing.T) {
			v2 := ParsesUntilExtra(t, in, " /* comment */[\n\twoof\n]", f, opts...)
			should.Equal(t, v2, v, cmpopts.IgnoreTypes((*regexp.Regexp)(nil))) // EOF and bracket text results differ
		})
	})
	return v
}

func ParsesUntilExtra[T any](t *testing.T, in, extra string, f parser.Func[T], opts ...Option) T {
	t.Helper()

	o := applyOptions(opts...)

	p := NewParser(t, in+extra)
	v := shouldMatch(t, p, f, o)
	checkErrors(t, p, o)
	shouldBeAtEnd(t, p, in)

	return v
}

func NoMatch[T any](t *testing.T, in string, f parser.Func[T], opts ...Option) {
	t.Helper()

	o := applyOptions(opts...)
	if len(o.errors) > 0 {
		t.Fatal("NoMatch must not want errors")
	}

	p := NewParser(t, in)

	var v T
	if o.inline {
		p.DoInline(func() { v = parser.Try(p, f) })
	} else {
		v = parser.Try(p, f)
	}

	if should.True(t, isZero(v)) {
		if !should.True(t, len(p.Errors()) == 0) { // got errors despite no match
			t.Log(p.Errors().Pretty(diagnostic.PrettyOptions{}))
		}
	} else {
		t.Logf("parsed value: %#v", v)
	}
}

func AlsoFulfils[I, O any](t *testing.T, f parser.Func[I], subTest func(*testing.T, parser.Func[O])) {
	t.Helper()

	var zeroO O
	oType := reflect.TypeOf(&zeroO).Elem()
	for oType.Kind() == reflect.Pointer {
		oType = oType.Elem()
	}

	t.Run(oType.Name(), func(t *testing.T) {
		t.Parallel()
		t.Helper()
		subTest(t, coerceFunc[I, O](t, f))
	})
}
