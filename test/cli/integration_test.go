//go:build integration_test

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileName(t *testing.T) {
	t.Parallel()

	assert.FileExists(t, "tacos.go")
}
