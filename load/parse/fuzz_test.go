package parse

import (
	"testing"

	"github.com/mavolin/corgi/v2/load/internal/fuzzdata"
)

func Foo() {}

// FuzzParse tests that the parser doesn't crash or hang on arbitrary input.
func FuzzParse(f *testing.F) {
	fuzzdata.AddBaseCorpus(f)

	f.Fuzz(func(t *testing.T, data string) {
		// We're just checking that Parse doesn't panic, so we ignore the return values
		_, _ = Parse(data, Options{})
	})
}
