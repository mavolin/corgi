package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/should"
)

func TestLinker_CheckReservedAliases(t *testing.T) {
	t.Parallel()

	var start ast.Position
	importedPkg := createPackage("imported")
	importedF := createFile(importedPkg, "imported.corgi")
	importedComp := createComponent(importedF, &start, "Test")
	addComponent(importedPkg, importedComp)

	mainPkg := createPackage("main")
	mainF := createFile(mainPkg, "main.corgi")
	createImport(mainF, &start, "__corgi_illegal", importedPkg.CorgiImportPath)

	d := Link(context.Background(), mainPkg, Options{
		Importer: ImporterFor(importedPkg),
	})

	t.Log(d.Pretty(diagnostic.PrettyOptions{}))
	if should.Equal(t, len(d), 1) {
		should.Equal(t, d[0].Message, "import alias: cannot use `__corgi_` prefix")
	}
}
