package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestImplicitCodeLine(t *testing.T) {
	t.Parallel()
	testutil.AssertAlsoFulfils(t, ImplicitCodeLine(), func(t *testing.T, f parser.Func[*ast.ImplicitCodeLine]) {
		testParsedStatement(t, func(p *parser.Parser) (*ast.Statement, *fancyerr.Error) {
			f, err := f(p)
			if err != nil {
				return nil, err
			}

			return f.Statement, nil
		})
	})
}

func TestExplicitCodeLine(t *testing.T) {
	t.Parallel()

	in := "- foo();"
	expect := &ast.ExplicitCodeLine{
		Minus: &ast.Position{Line: 1, Col: 1},
		Statement: &ast.Statement{
			Code: ast.Code{
				&ast.GoCode{Code: "foo()", Position: &ast.Position{Line: 1, Col: 3}},
			},
		},
	}

	actual := testutil.ParsesFully(t, in, ExplicitCodeLine())
	assert.Equal(t, expect, actual)
}
