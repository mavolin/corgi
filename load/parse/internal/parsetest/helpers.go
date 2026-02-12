package parsetest

import (
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
			return zero
		}

		return t
	}
}

func calcEnd(in string) (int, int, parser.ByteIndex) {
	line, col, index := 1, 1, 0
	for _, r := range in {
		if r == '\n' {
			line++
			col = 1
		} else {
			col++
		}
		index += len(string(r))
	}
	return line, col, parser.ByteIndex(index) //nolint:gosec
}
