package link

import (
	"context"
	"testing"

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
