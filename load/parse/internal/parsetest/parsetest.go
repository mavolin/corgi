package parsetest

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
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
		AST: &ast.File{
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
	v := f(p)

	if !should.True(t, isZero(v)) {
		t.Logf("parsed value: %#v", v)
	}
}

func MatchesButError[T any](t *testing.T, input string, f parser.Func[T]) T {
	t.Helper()

	p := NewParser(t, input)
	return AssertMatchesButError(t, p, f)
}

func AssertMatchesButError[T any](t *testing.T, p *parser.Parser, f parser.Func[T]) T {
	t.Helper()

	v := f(p)
	should.False(t, isZero(v)) // match error

	// diagnostic.List
	errors := p.Errors()
	should.True(t, len(errors) > 0)
	should.False(t, slices.Contains(errors, nil)) // nil error was captured

	return v
}

func AssertNoError[T any](t *testing.T, p *parser.Parser, f parser.Func[T]) T {
	t.Helper()

	v := f(p)
	should.False(t, isZero(v)) // match error

	for _, err := range p.Errors() {
		should.NotEqual(t, nil, err) // diagnostic.List: nil error was captured
		should.Equal(t, nil, err)    // diagnostic.List: unexpected error
	}

	return v
}

func AssertEOF(t *testing.T, p *parser.Parser) {
	t.Helper()
	line, col, index := CalcEnd(1, 1, 0, p.AST.Raw)
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
	t.Helper()
	should.Equal(t, int(p.Line()), line) // position: line mismatch
	should.Equal(t, int(p.Col()), col)   // position: col mismatch
	should.Equal(t, p.Index(), index)    // position: index mismatch
}

func CoerceFunc[I, O any](t *testing.T, in parser.Func[I]) parser.Func[O] {
	t.Helper()

	return func(p *parser.Parser) O {
		var zero O

		v := in(p)
		if isZero(v) {
			return zero
		}

		t, ok := any(v).(O)
		if !ok {
			return zero
		}

		return t
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
		t.Helper()
		subTest(t, CoerceFunc[I, O](t, f))
	})
}

func isZero[T any](t T) bool {
	return reflect.ValueOf(&t).Elem().IsZero()
}
