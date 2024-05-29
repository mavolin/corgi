package testutil

import (
	"strings"
	"testing"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
	parser "github.com/mavolin/corgi/load/parse/internal"
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
		AST: &ast.AST{
			Raw:   input,
			Lines: lines,
		},
	})
}

func AssertParsesFully[T any](t *testing.T, input string, f parser.Func[T]) T {
	t.Helper()

	p := NewParser(t, input)
	v := AssertNoError(t, p, f)
	AssertEOF(t, p)
	return v
}

func AssertMatchesButError[T any](t *testing.T, input string, f parser.Func[T]) T {
	t.Helper()

	p := NewParser(t, input)
	return AssertMatchButError(t, p, f)
}

func AssertMatchButError[T any](t *testing.T, p *parser.Parser, f parser.Func[T]) T {
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
	assert.Equal(t, line, p.Line(), "line mismatch")
	assert.Equal(t, col, p.Col(), "col mismatch")
	assert.Equal(t, index, p.Index(), "index mismatch")
}
