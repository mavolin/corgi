package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/should"
)

func TestLinker_CheckComponentCollisions(t *testing.T) {
	t.Parallel()

	var start ast.Position
	pkg := createPackage("test")
	f := createFile(pkg, "test.corgi")

	createComponent(f, &start, "Same")
	createComponent(f, &start, "Same")
	d := Link(context.Background(), pkg, Options{})

	t.Log(d.Pretty(diagnostic.PrettyOptions{}))
	if should.Equal(t, len(d), 1) {
		should.Equal(t, d[0].Message, "component defined multiple times")
	}
}

func TestLinker_CheckElementSpecCollisions(t *testing.T) {
	t.Parallel()

	t.Run("qualified collision", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		pkg := createPackage("test")
		f := createFile(pkg, "test.corgi")

		createElementSpec(f, &start, "foo", "Same", elemtype.Normal)
		createElementSpec(f, &start, "bar", "Same", elemtype.Normal)
		d := Link(context.Background(), pkg, Options{})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "multiple elements with same qualified name")
		}
	})

	t.Run("html name collision", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		pkg := createPackage("test")
		f := createFile(pkg, "test.corgi")

		createElementSpec(f, &start, "foo", "same", elemtype.Normal)
		createElementSpec(f, &start, "foos", "ame", elemtype.Normal)
		d := Link(context.Background(), pkg, Options{})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "multiple elements with same html name")
		}
	})

	t.Run("both", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		pkg := createPackage("test")
		f := createFile(pkg, "test.corgi")

		createElementSpec(f, &start, "foo", "same", elemtype.Normal)
		createElementSpec(f, &start, "foo", "same", elemtype.Normal)
		d := Link(context.Background(), pkg, Options{})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "multiple elements with same qualified name")
		}
	})
}

func TestLinker_CheckAttributeSpecCollisions(t *testing.T) {
	t.Parallel()

	t.Run("qualified collision", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		pkg := createPackage("test")
		f := createFile(pkg, "test.corgi")

		createBasicAttributeSpec(f, &start, "foo", "Same", nil, attrtype.Innocuous)
		createBasicAttributeSpec(f, &start, "bar", "Same", nil, attrtype.Innocuous)
		d := Link(context.Background(), pkg, Options{})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "multiple attributes with same qualified selector")
		}
	})

	t.Run("html name collision", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		pkg := createPackage("test")
		f := createFile(pkg, "test.corgi")

		createBasicAttributeSpec(f, &start, "foo", "same", nil, attrtype.Innocuous)
		createBasicAttributeSpec(f, &start, "foos", "ame", nil, attrtype.Innocuous)
		d := Link(context.Background(), pkg, Options{})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "multiple attributes with same html name selector")
		}
	})

	t.Run("both", func(t *testing.T) {
		t.Parallel()

		var start ast.Position
		pkg := createPackage("test")
		f := createFile(pkg, "test.corgi")

		createBasicAttributeSpec(f, &start, "foo", "same", nil, attrtype.Innocuous)
		createBasicAttributeSpec(f, &start, "foo", "same", nil, attrtype.Innocuous)
		d := Link(context.Background(), pkg, Options{})

		t.Log(d.Pretty(diagnostic.PrettyOptions{}))
		if should.Equal(t, len(d), 1) {
			should.Equal(t, d[0].Message, "multiple attributes with same qualified selector")
		}
	})
}
