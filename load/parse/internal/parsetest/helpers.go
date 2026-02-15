package parsetest

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

func CalcEndPos(in string) ast.Position {
	line, col, _ := calcEnd(in)
	return ast.Position{Line: line, Col: col}
}

// NewParser create a new parser for the given input.
//
// Most tests won't need to call this directly, but will use one of the
// ParsesUntil* helpers instead.
func NewParser(t *testing.T, in string) *parser.Parser {
	t.Helper()

	lines := strings.Split(in, "\n")
	for i, line := range lines {
		last := len(line) - 1
		if len(line) > 0 && line[last] == '\r' {
			lines[i] = line[:last]
		}
	}

	return parser.New(&file.File{
		Name:  file.Name(t.Name()),
		Raw:   in,
		Lines: lines,
	}, in, ast.Position{Line: 1, Col: 1})
}

func isZero[T any](t T) bool {
	return reflect.ValueOf(&t).Elem().IsZero()
}

// coerceFunc takes a parser.Func[I] and returns a parser.Func[O], that parses
// using the input function.
func coerceFunc[I, O any](t *testing.T, in parser.Func[I]) parser.Func[O] {
	t.Helper()

	return func(p *parser.Parser) O {
		var zero O

		v := parser.Try(p, in)
		if isZero(v) {
			return zero
		}

		t, ok := any(v).(O)
		if !ok {
			panic(fmt.Sprintf("func %T did not return type %T for its input", in, zero))
		}

		return t
	}
}

// calcEnd calculates the line, column and byte index of the end of the given
// input.
func calcEnd(in string) (ast.Line, ast.Col, parser.ByteIndex) {
	line, col, index := ast.Line(1), ast.Col(1), parser.ByteIndex(0)
	for _, r := range in {
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
		index += parser.ByteIndex(len(string(r))) //nolint:gosec
	}
	return line, col, index
}
