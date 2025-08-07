package parse

import (
	"os"
	"testing"
)

func BenchmarkParse(b *testing.B) {
	data, err := os.ReadFile("../../examples/readme/readme.corgi")
	if err != nil {
		b.Fatalf("failed to read input file: %v", err)
	}

	in := string(data)
	b.ResetTimer()
	for range b.N {
		_, _ = Parse(in, Options{})
	}
}
