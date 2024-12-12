package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
)

func TestPackageName(t *testing.T) {
	t.Parallel()

	testutil.AssertAlsoFulfils(t, PackageName(), testIdentifier)
}
