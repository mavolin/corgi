package testutil

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/stretchr/testify/assert"
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
		File: &ast.File{
			Raw:   input,
			Lines: lines,
		},
	})
}

func ParsesFully[T any](t *testing.T, input string, f parser.Func[T]) T {
	t.Helper()

	p := NewParser(t, input+" 1other stuff")
	v := AssertNoError(t, p, f)

	line, col, index := CalcEnd(1, 1, 0, input)
	AssertPosition(t, p, line, col, index)

	return v
}

func NoMatch[T any](t *testing.T, input string, f parser.Func[T]) {
	t.Helper()

	p := NewParser(t, input)
	v, err := f(p)
	assert.NotNilf(t, err, "expected match error, found: %#v", v)
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
		if assert.NotNil(t, err, "a nil error was captured") {
			assert.Fail(t, "unexpected error", err.Message)
		}
	}

	return v
}

func AssertEOF(t *testing.T, p *parser.Parser) {
	line, col, index := CalcEnd(1, 1, 0, p.Raw)
	AssertPosition(t, p, line, col, index)
}

func CalcEnd(line, col, index int, input string) (int, int, int) {
	for _, r := range input {
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
		index += len(string(r))
	}
	return line, col, index
}

func CalcEndPos(in string) ast.Position {
	line, col, _ := CalcEnd(1, 1, 0, in)
	return ast.Position{Line: line, Col: col}
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

		t, ok := any(v).(O)
		if !ok {
			return zero, &fancyerr.Error{
				Message: fmt.Sprintf("expected %T, found %T", zero, v),
			}
		}

		return t, nil
	}
}

func AssertAlsoFulfils[I, O any](t *testing.T, f parser.Func[I], subTest func(*testing.T, parser.Func[O])) {
	t.Helper()

	var zeroO O
	oType := reflect.TypeOf(&zeroO).Elem()
	for oType.Kind() == reflect.Pointer {
		oType = oType.Elem()
	}

	t.Run(oType.Name(), func(t *testing.T) {
		t.Parallel()
		subTest(t, CoerceFunc[I, O](t, f))
	})
}
