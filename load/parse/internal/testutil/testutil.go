package testutil

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mavolin/corgi/fancyerr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func NewParser(t *testing.T, input string) *parser.Parser {
	t.Helper()

	lines := strings.Split(input, "\n")
	for i, line := range lines {
		last := len(line) - 1
		if len(line) > 0 && line[last] == '\r' {
			lines[i] = line[:last]
		}
	}

	return parser.New(&file.File{
		Name: t.Name(),
		AST: &ast.AST{
			Raw:   input,
			Lines: lines,
		},
	})
}

func ParsesFully[T any](t *testing.T, input string, f parser.Func[T]) T {
	t.Helper()

	p := NewParser(t, input)
	v := AssertNoError(t, p, f)
	AssertEOF(t, p)
	return v
}

func NoMatch[T any](t *testing.T, input string, f parser.Func[T]) {
	t.Helper()

	p := NewParser(t, input)
	_, err := f(p)
	assert.Error(t, err, "expected match error")
}

func MatchesButError[T any](t *testing.T, input string, f parser.Func[T]) T {
	t.Helper()

	p := NewParser(t, input)
	return AssertMatchesButError(t, p, f)
}

func AssertMatchesButError[T any](t *testing.T, p *parser.Parser, f parser.Func[T]) T {
	t.Helper()

	v, err := f(p)
	assert.Nil(t, err, "match error")

	assert.NotEmpty(t, p.CloneState().Errors(), "no error was captured")
	for _, err = range p.CloneState().Errors() {
		assert.NotNil(t, err, "a nil error was captured")
	}

	return v
}

func AssertNoError[T any](t *testing.T, p *parser.Parser, f parser.Func[T]) T {
	t.Helper()

	v, err := f(p)
	assert.Nil(t, err, "match error")

	for _, err = range p.CloneState().Errors() {
		if !assert.NotNil(t, err, "a nil error was captured") {
			assert.Fail(t, "unexpected error: %s", err)
		}
	}

	return v
}

func AssertEOF(t *testing.T, p *parser.Parser) {
	line, col, index := 1, 1, 0
	for _, r := range p.Raw {
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
		index++
	}
	AssertPosition(t, p, line, col, index)
}

func AssertPosition(t *testing.T, p *parser.Parser, line, col, index int) {
	assert.Equal(t, line, p.Line(), "line mismatch")
	assert.Equal(t, col, p.Col(), "col mismatch")
	assert.Equal(t, index, p.Index(), "index mismatch")
}

func CoerceFunc[I, O any](t *testing.T, in parser.Func[I]) parser.Func[O] {
	return func(p *parser.Parser) (O, *fancyerr.Error) {
		var zero O

		v, err := in(p)
		if err != nil {
			return zero, err
		}

		require.IsType(t, zero, v)
		return any(in).(O), err
	}
}

func AssertAlsoFulfils[I, O any](t *testing.T, f parser.Func[I], subTest func(*testing.T, parser.Func[O])) {
	t.Helper()
	var zeroI I
	iType := reflect.TypeOf(&zeroI).Elem()
	for iType.Kind() == reflect.Pointer {
		iType = iType.Elem()
	}
	t.Run(iType.Name(), func(t *testing.T) {
		t.Parallel()
		subTest(t, CoerceFunc[I, O](t, f))
	})
}
