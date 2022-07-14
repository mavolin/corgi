//go:build prepare_integration_test

package elements

import (
	"testing"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/test/internal/compile"
)

func TestEmptyElement(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "empty_element.corgi", compile.Options{})
}

func TestVoidElements(t *testing.T) {
	t.Parallel()

	compile.Compile(t, "void_elements.html.corgi", compile.Options{FileType: file.TypeHTML})
	compile.Compile(t, "void_elements.xhtml.corgi", compile.Options{FileType: file.TypeXHTML})
}

func TestAttributes(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "attributes.corgi", compile.Options{})
}

func TestBlockExpansion(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "block_expansion.corgi", compile.Options{})
}

func TestSelfClosing(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "self_closing.corgi", compile.Options{})
}

func TestMirror(t *testing.T) {
	t.Parallel()

	compile.Compile(t, "mirror.html.corgi", compile.Options{FileType: file.TypeHTML})
	compile.Compile(t, "mirror.xhtml.corgi", compile.Options{FileType: file.TypeXHTML})
	compile.Compile(t, "mirror.xml.corgi", compile.Options{FileType: file.TypeXML})
}
