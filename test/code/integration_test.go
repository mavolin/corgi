//go:build integration_test

package code

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/voidwriter"
)

func TestInlineCode(t *testing.T) {
	t.Parallel()

	var executed bool

	err := InlineCode(voidwriter.Writer, &executed)
	require.NoError(t, err)

	assert.True(t, executed, "code was not executed")
}

func TestCodeBlock(t *testing.T) {
	t.Parallel()

	var executed bool

	err := CodeBlock(voidwriter.Writer, &executed)
	require.NoError(t, err)

	assert.True(t, executed, "code was not executed")
}

func TestGlobalCode(t *testing.T) {
	t.Parallel()

	// this would actually give us a compiler error, so there is no point in
	// writing an error message
	assert.True(t, globalCodeExecuted)
}

func TestInlineImport(t *testing.T) {
	t.Parallel()

	// compiler error
	err := InlineImport(voidwriter.Writer)
	require.NoError(t, err)
}

func TestImportBlock(t *testing.T) {
	t.Parallel()

	// compiler error
	err := ImportBlock(voidwriter.Writer)
	require.NoError(t, err)
}

func TestImportAlias(t *testing.T) {
	t.Parallel()

	// compiler error
	err := ImportAlias(voidwriter.Writer)
	require.NoError(t, err)
}
