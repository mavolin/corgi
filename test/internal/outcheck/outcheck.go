// Package outcheck provides a utility to check if the output of a template
// function matches a file.
package outcheck

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Checker struct {
	t *testing.T

	actual bytes.Buffer
	expect string
}

var _ io.Writer = &Checker{}

// New creates a new Checker, that will check if the output written to it
// matches that of expectFile.
//
// The check is performed automatically at the end of the test through
// testing.T.Cleanup.
func New(t *testing.T, expectFile string) *Checker {
	t.Helper()

	expect, err := os.ReadFile(expectFile)
	require.NoError(t, err)

	c := &Checker{t: t, expect: string(expect)}

	t.Cleanup(c.check)

	return c
}

func (c *Checker) Write(data []byte) (int, error) {
	return c.actual.Write(data)
}

func (c *Checker) check() {
	c.t.Helper()
	assert.Equalf(c.t, c.expect, c.actual.String(), "output of template did not match")
}
