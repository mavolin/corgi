package parsetest

import (
	"slices"
	"testing"

	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

func shouldMatch[T any](t *testing.T, p *parser.Parser, f parser.Func[T], o options) T {
	t.Helper()

	var v T
	if o.inline {
		p.DoInline(func() { v = parser.Try(p, f) })
	} else {
		v = parser.Try(p, f)
	}
	should.False(t, isZero(v)) // parser did not match
	return v
}

func checkErrors(t *testing.T, p *parser.Parser, o options) {
	t.Helper()

	if len(o.errors) == 0 && len(p.Errors()) == 0 {
		return // fast path
	}

	shouldValidDiagnostics(t, p.Errors())

	seen := make(map[string]bool, len(o.errors))

	ds := p.Errors()
	ds.Tidy()
	for _, d := range ds {
		if slices.Contains(o.errors, d.Message) {
			seen[d.Message] = true
			continue
		}

		t.Error("unexpected error:\n" + d.Pretty(diagnostic.PrettyOptions{}))
	}

	for _, msg := range o.errors {
		if !seen[msg] {
			t.Errorf("error not seen: %q", msg)
		}
	}
}

func shouldValidDiagnostics(t *testing.T, ds diagnostic.List) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("panic: diagnostic.List.Pretty: %v", r)
		}
	}()
	ds.Pretty(diagnostic.PrettyOptions{})
}

func shouldBeAtEnd(t *testing.T, p *parser.Parser, in string) {
	t.Helper()

	wantLine, wantCol, wantIndex := calcEnd(in)
	should.Equal(t, int(p.Line()), wantLine) // end of input: incorrect line
	should.Equal(t, int(p.Col()), wantCol)   // end of input: incorrect col
	should.Equal(t, p.Index(), wantIndex)    // end of input: incorrect index
}
