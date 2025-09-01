package parse

import (
	"fmt"
	"os"
	"testing"

	"github.com/mavolin/corgi/v2/file/diagnostic"
)

func BenchmarkParse(b *testing.B) {
	data, err := os.ReadFile("../../examples/readme/readme.corgi")
	if err != nil {
		b.Fatalf("failed to read input file: %v", err)
	}

	in := string(data)

	_, d := Parse(in, Options{})
	if len(d) > 0 {
		fmt.Println(d.Pretty(diagnostic.PrettyOptions{}))
		b.FailNow()
	}

	b.ResetTimer()
	for range b.N {
		_, _ = Parse(in, Options{})
	}
}
