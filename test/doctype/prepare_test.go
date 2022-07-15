//go:build prepare_integration_test

package doctype

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestOneDotOne(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "1.1.corgi", compile.Options{})
}

func TestBasic(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "basic.corgi", compile.Options{})
}

func TestCustomDoctype(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "custom_doctype.corgi", compile.Options{})
}

func TestCustomProlog(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "custom_prolog.corgi", compile.Options{})
}

func TestFrameset(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "frameset.corgi", compile.Options{})
}

func TestHTML(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "html.corgi", compile.Options{})
}

func TestMobile(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "mobile.corgi", compile.Options{})
}

func TestPList(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "plist.corgi", compile.Options{})
}

func TestStrict(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "strict.corgi", compile.Options{})
}

func TestTransitional(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "transitional.corgi", compile.Options{})
}

func TestXML(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "xml.corgi", compile.Options{})
}
