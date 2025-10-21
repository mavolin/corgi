package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestImplicitCodeLine(t *testing.T) {
	t.Parallel()
	parsetest.AlsoFulfils(t, ImplicitCodeLine(), func(t *testing.T, f parser.Func[*ast.ImplicitCodeLine]) {
		testParsedStatement(t, func(p *parser.Parser) *ast.Statement {
			f := f(p)
			if f == nil {
				return nil
			}

			return f.Statement
		})
	})
}

func TestExplicitCodeLine(t *testing.T) {
	t.Parallel()

	in := "- foo()"
	want := &ast.ExplicitCodeLine{
		Minus: &ast.Position{Line: 1, Col: 1},
		Statement: &ast.Statement{
			Nodes: ast.Code{
				&ast.GoCode{Code: "foo()", Position: &ast.Position{Line: 1, Col: 3}},
			},
		},
	}

	got := parsetest.ParsesUntilEOS(t, in, ExplicitCodeLine())
	should.Equal(t, got, want)
}
