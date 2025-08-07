package link

import (
	"context"
	"testing"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestLinker_CheckComponentCollisions(t *testing.T) {
	t.Parallel()

	pkg := createPackage("test")
	f := createFile(pkg, "test.corgi")

	createComponent(f, nil, "Same")
	createComponent(f, nil, "Same")
	ds := Link(context.Background(), pkg, Options{})

	if should.Equal(t, len(ds), 1) {
		if !should.Equal(t, ds[0].Message, "component defined multiple times") {
			t.Log(ds[0].Short())
		}
	} else {
		t.Log(ds.Short())
	}
}

func TestLinker_CheckElementSpecCollisions(t *testing.T) {
	t.Parallel()

	t.Run("qualified collision", func(t *testing.T) {
		t.Parallel()

		pkg := createPackage("test")
		f := createFile(pkg, "test.corgi")

		createElementSpec(f, nil, "foo", "Same", elemtype.Normal)
		createElementSpec(f, nil, "bar", "Same", elemtype.Normal)
		ds := Link(context.Background(), pkg, Options{})

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "multiple elements with same qualified name") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("html name collision", func(t *testing.T) {
		t.Parallel()

		pkg := createPackage("test")
		f := createFile(pkg, "test.corgi")

		createElementSpec(f, nil, "foo", "same", elemtype.Normal)
		createElementSpec(f, nil, "foos", "ame", elemtype.Normal)
		ds := Link(context.Background(), pkg, Options{})

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "multiple elements with same html name") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("both", func(t *testing.T) {
		t.Parallel()

		pkg := createPackage("test")
		f := createFile(pkg, "test.corgi")

		createElementSpec(f, nil, "foo", "same", elemtype.Normal)
		createElementSpec(f, nil, "foo", "same", elemtype.Normal)
		ds := Link(context.Background(), pkg, Options{})

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "multiple elements with same qualified name") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})
}

func TestLinker_CheckAttributeSpecCollisions(t *testing.T) {
	t.Parallel()

	t.Run("qualified collision", func(t *testing.T) {
		t.Parallel()

		pkg := createPackage("test")
		f := createFile(pkg, "test.corgi")

		createBasicAttributeSpec(f, nil, "foo", "Same", nil, attrtype.Innocuous)
		createBasicAttributeSpec(f, nil, "bar", "Same", nil, attrtype.Innocuous)
		ds := Link(context.Background(), pkg, Options{})

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "multiple attributes with same qualified selector") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("html name collision", func(t *testing.T) {
		t.Parallel()

		pkg := createPackage("test")
		f := createFile(pkg, "test.corgi")

		createBasicAttributeSpec(f, nil, "foo", "same", nil, attrtype.Innocuous)
		createBasicAttributeSpec(f, nil, "foos", "ame", nil, attrtype.Innocuous)
		ds := Link(context.Background(), pkg, Options{})

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "multiple attributes with same html name selector") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})

	t.Run("both", func(t *testing.T) {
		t.Parallel()

		pkg := createPackage("test")
		f := createFile(pkg, "test.corgi")

		createBasicAttributeSpec(f, nil, "foo", "same", nil, attrtype.Innocuous)
		createBasicAttributeSpec(f, nil, "foo", "same", nil, attrtype.Innocuous)
		ds := Link(context.Background(), pkg, Options{})

		if should.Equal(t, len(ds), 1) {
			if !should.Equal(t, ds[0].Message, "multiple attributes with same qualified selector") {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})
}
