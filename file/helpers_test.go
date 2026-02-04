package file

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/should"
)

func TestImportConstantsValid(t *testing.T) {
	t.Parallel()
	t.Run("EscapeImport", func(t *testing.T) {
		t.Parallel()
		should.NoError(t, EscapeImport.CheckValid())
	})
	t.Run("SafeImport", func(t *testing.T) {
		t.Parallel()
		should.NoError(t, SafeImport.CheckValid())
	})
	t.Run("RuntimeImport", func(t *testing.T) {
		t.Parallel()
		should.NoError(t, RuntimeImport.CheckValid())
	})
}
