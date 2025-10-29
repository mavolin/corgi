package assert

import (
	"errors"
	"testing"

	"github.com/mavolin/corgi/v2/internal/should"
)

func TestAlways(t *testing.T) {
	t.Parallel()

	t.Run("cond=true", func(t *testing.T) {
		t.Parallel()
		Always(true, "should not panic")
	})
	t.Run("cond=false", func(t *testing.T) {
		t.Parallel()
		should.Panic(t, func() {
			Always(false, "msg")
		}, "assertion failed: msg")
	})
}

func TestNoError(t *testing.T) {
	t.Parallel()

	t.Run("err=nil", func(t *testing.T) {
		t.Parallel()
		NoError(nil, "should not panic")
	})
	t.Run("err!=nil", func(t *testing.T) {
		t.Parallel()
		should.Panic(t, func() {
			NoError(errors.New("err"), "msg")
		}, "assertion failed: msg: err")
	})
}

func TestDebug(t *testing.T) {
	prev := DebugEnabled
	t.Cleanup(func() {
		DebugEnabled = prev
	})

	t.Run("DebugEnabled=false/cond=false", func(*testing.T) {
		DebugEnabled = false
		Debug(func() bool { return false }, "should not panic")
	})
	t.Run("DebugEnabled=true/cond=true", func(*testing.T) {
		DebugEnabled = true
		Debug(func() bool { return true }, "should not panic")
	})
	t.Run("DebugEnabled=true/cond=false", func(t *testing.T) {
		DebugEnabled = true
		should.Panic(t, func() {
			Debug(func() bool { return false }, "msg")
		}, "debug assertion failed: msg")
	})
}
