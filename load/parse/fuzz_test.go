package parse

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
	"github.com/mavolin/corgi/v2/load/internal/fuzzdata"
)

func Foo() {}

// FuzzParse tests that the parser doesn't crash or hang on arbitrary input.
func FuzzParse(f *testing.F) {
	fuzzdata.AddBaseCorpus(f)

	f.Fuzz(func(t *testing.T, data string) {
		_, d := Parse(data, Options{})
		if len(d) > 0 {
			// invalid diagnostics
			should.NotPanic(t, func() { d.Pretty(diagnostic.PrettyOptions{}) })
		}
	})
}
