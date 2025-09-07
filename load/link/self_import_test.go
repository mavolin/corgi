package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_CheckSelfImport(t *testing.T) {
	t.Parallel()

	var start ast.Position
	pkgA := createPackage("pkg/a")
	fileA := createFile(pkgA, "a.corgi")
	imp := createImport(fileA, &start, "", pkgA.ImportPath)
	imp.Package = pkgA

	d := Link(context.Background(), pkgA, Options{
		Importer: ImporterFor(),
	})

	t.Log(d.Pretty(diagnostic.PrettyOptions{}))
	if should.Equal(t, len(d), 1) {
		should.Equal(t, d[0].Message, "package imports itself")
	}
}
