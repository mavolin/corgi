package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
)

func TestFullIdent(t *testing.T) {
	t.Parallel()

	testutil.AssertAlsoFulfils(t, FullIdent(), testIdentifier)
	testutil.AssertAlsoFulfils(t, FullIdent(), testQualifiedIdent)
}
