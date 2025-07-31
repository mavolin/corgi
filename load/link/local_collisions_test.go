package link

import (
	"context"
	"testing"

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

	if should.Equal(t, 1, len(ds)) {
		if !should.Equal(t, "component defined multiple times", ds[0].Message) {
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

		if should.Equal(t, 1, len(ds)) {
			if !should.Equal(t, "element defined multiple times", ds[0].Message) {
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

		if should.Equal(t, 1, len(ds)) {
			if !should.Equal(t, "element defined multiple times", ds[0].Message) {
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

		if should.Equal(t, 1, len(ds)) {
			if !should.Equal(t, "element defined multiple times", ds[0].Message) {
				t.Log(ds[0].Short())
			}
		} else {
			t.Log(ds.Short())
		}
	})
}
