package gocmd

import (
	"os/exec"
	"testing"

	"github.com/mavolin/corgi/v2/internal/meta"
	"github.com/mavolin/corgi/v2/internal/should"
)

func TestCmd_ListAllModules(t *testing.T) {
	goExecPath, err := exec.LookPath("go")
	if err != nil {
		t.Fatal(err)
	}

	cmd := New(goExecPath)
	result, err := cmd.ListAllModules(t.Context(), ".")
	if should.NoError(t, err) {
		should.Equal(t, result.MainModule.Path, meta.Module)
	}
}
