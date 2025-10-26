package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/should"
)

func TestLinker_CheckLocalOnlyMode(t *testing.T) {
	t.Parallel()

	var start ast.Position
	mainPkg := createPackage("main")
	mainF := createFile(mainPkg, "main.corgi")
	createImport(mainF, &start, "", "github.com/some/package")

	d := Link(context.Background(), mainPkg, Options{
		Importer: nil, // nil importer triggers local-only mode
	})

	t.Log(d.Pretty(diagnostic.PrettyOptions{}))
	if should.Equal(t, len(d), 1) {
		should.Equal(t, d[0].Message, "local-only mode: file contains imports")
	}

}
