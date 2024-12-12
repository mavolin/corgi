package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestIdentifier(t *testing.T) {
	t.Parallel()
	testIdentifier(t, Identifier())
}

func testIdentifier(t *testing.T, f parser.Func[*ast.Ident]) {
	in := "foo"
	expect := &ast.Ident{
		Ident:    "foo",
		Position: ast.Position{Line: 1, Col: 1},
	}

	actual := testutil.ParsesFully(t, in, f)
	assert.Equal(t, expect, actual)
}
