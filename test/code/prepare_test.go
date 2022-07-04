//go:build prepare_integration_test

package code

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestInlineCode(t *testing.T) {
	compile.Compile(t, "inline_code.corgi", compile.Options{})
}

func TestCodeBlock(t *testing.T) {
	compile.Compile(t, "code_block.corgi", compile.Options{})
}

func TestGlobalCode(t *testing.T) {
	compile.Compile(t, "global_code.corgi", compile.Options{})
}

func TestInlineImport(t *testing.T) {
	compile.Compile(t, "inline_import.corgi", compile.Options{})
}

func TestImportBlock(t *testing.T) {
	compile.Compile(t, "import_block.corgi", compile.Options{})
}

func TestImportAlias(t *testing.T) {
	compile.Compile(t, "import_alias.corgi", compile.Options{})
}
