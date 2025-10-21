package should

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

type fakeT struct {
	msgs []string
}

var _ TInterface = (*fakeT)(nil)

func (f *fakeT) Error(args ...any) { f.msgs = append(f.msgs, fmt.Sprint(args...)) }
func (f *fakeT) Helper()           {}

func TestEqual(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		got, want any
		ok        bool
		wantMsg   bool
	}{
		{"equal values", 42, 42, true, false},
		{"different values", 1, 2, false, true},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			f := &fakeT{}
			if Equal(f, c.got, c.want) != c.ok {
				t.Fatalf("Equal returned unexpected result; msgs=%v", f.msgs)
			}
			if (len(f.msgs) > 0) != c.wantMsg {
				t.Fatalf("message presence mismatch; msgs=%v", f.msgs)
			}
		})
	}
}

func TestNotEqual(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		got, want any
		ok        bool
		wantMsg   bool
	}{
		{"different values", 1, 2, true, false},
		{"equal values", 3, 3, false, true},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			f := &fakeT{}
			if NotEqual(f, c.got, c.want) != c.ok {
				t.Fatalf("NotEqual returned unexpected result; msgs=%v", f.msgs)
			}
			if (len(f.msgs) > 0) != c.wantMsg {
				t.Fatalf("message presence mismatch; msgs=%v", f.msgs)
			}
		})
	}
}

func TestPanic(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		fn          func()
		wantMessage string
		ok          bool
		wantMsg     bool
	}{
		{
			name:        "panics with expected message",
			fn:          func() { panic("boom") },
			wantMessage: "boom",
			ok:          true,
			wantMsg:     false,
		}, {
			name:        "does not panic",
			fn:          func() {},
			wantMessage: "",
			ok:          false,
			wantMsg:     true,
		}, {
			name:        "panics with different message",
			fn:          func() { panic("kapow") },
			wantMessage: "boom",
			ok:          false,
			wantMsg:     true,
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			f := &fakeT{}
			if Panic(f, c.fn, c.wantMessage) != c.ok {
				t.Fatalf("Panic returned unexpected result; msgs=%v", f.msgs)
			}
			if (len(f.msgs) > 0) != c.wantMsg {
				t.Fatalf("message presence mismatch; msgs=%v", f.msgs)
			}
		})
	}
}

func TestError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		err           error
		wantMsg       string
		ok            bool
		wantMsgOnFail bool
	}{
		{"matches message", errors.New("oh no"), "oh no", true, false},
		{"nil error", nil, "", false, true},
		{"mismatched message", errors.New("x"), "y", false, true},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			f := &fakeT{}
			if Error(f, c.err, c.wantMsg) != c.ok {
				t.Fatalf("Error returned unexpected result; msgs=%v", f.msgs)
			}
			if (len(f.msgs) > 0) != c.wantMsgOnFail {
				t.Fatalf("message presence mismatch; msgs=%v", f.msgs)
			}
		})
	}
}

func TestNoError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		err     error
		ok      bool
		wantMsg bool
	}{
		{"nil error", nil, true, false},
		{"non-nil error", errors.New("err"), false, true},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			f := &fakeT{}
			if NoError(f, c.err) != c.ok {
				t.Fatalf("NoError returned unexpected result; msgs=%v", f.msgs)
			}
			if (len(f.msgs) > 0) != c.wantMsg {
				t.Fatalf("message presence mismatch; msgs=%v", f.msgs)
			}
		})
	}
}

func TestTrue(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		got     bool
		ok      bool
		wantMsg bool
	}{
		{"is true", true, true, false},
		{"is false", false, false, true},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			f := &fakeT{}
			if True(f, c.got) != c.ok {
				t.Fatalf("True returned unexpected result; msgs=%v", f.msgs)
			}
			if (len(f.msgs) > 0) != c.wantMsg {
				t.Fatalf("message presence mismatch; msgs=%v", f.msgs)
			}
		})
	}
}

func TestFalse(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		got     bool
		ok      bool
		wantMsg bool
	}{
		{"is false", false, true, false},
		{"is true", true, false, true},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			f := &fakeT{}
			if False(f, c.got) != c.ok {
				t.Fatalf("False returned unexpected result; msgs=%v", f.msgs)
			}
			if (len(f.msgs) > 0) != c.wantMsg {
				t.Fatalf("message presence mismatch; msgs=%v", f.msgs)
			}
		})
	}
}

func TestArgs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name           string
		lines          []string
		fnName         string
		comparison     bool
		target         int
		arg1, op, arg2 string
	}{
		{
			name:   "Equal simple args",
			lines:  []string{"\tfoo.Equal(t, a, b)"},
			fnName: "Equal", comparison: false, target: 1,
			arg1: "a", op: "", arg2: "b",
		},
		{
			name:   "True with comparison",
			lines:  []string{"\tbar.True(t, x < y)"},
			fnName: "True", comparison: true, target: 1,
			arg1: "x", op: "<", arg2: "y",
		},
		{
			name:   "Equal with parens and quotes",
			lines:  []string{"\tfoo.Equal(t, (a + b), \"some,thing\")"},
			fnName: "Equal", comparison: false, target: 1,
			arg1: "(a + b)", op: "", arg2: "\"some,thing\"",
		},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			arg1, op, arg2 := args(c.lines, c.fnName, c.comparison, c.target)
			if arg1 != c.arg1 || op != c.op || arg2 != c.arg2 {
				t.Fatalf("unexpected args parsed: %q %q %q", arg1, op, arg2)
			}
		})
	}
}

func TestComment(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		content string
		target  int
		expectC string
		multi   bool
	}{
		{
			name:    "group comment above target",
			content: "package should\n// group comment\nfunc Test() {}\n",
			target:  3,
			expectC: "group comment",
			multi:   true,
		},
		{
			name:    "inline comment on target line",
			content: "package should\n\nfunc Test() {} // inline comment\n",
			target:  3,
			expectC: "inline comment",
			multi:   false,
		},
	}
	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			f, err := os.CreateTemp("", "should-comment-*.go")
			if err != nil {
				t.Fatal(err)
			}
			defer os.Remove(f.Name())
			if _, err := f.WriteString(c.content); err != nil {
				t.Fatal(err)
			}
			f.Close()

			lns := lines(f.Name())
			cc, multi := comment(lns, c.target)
			if cc != c.expectC || multi != c.multi {
				t.Fatalf("unexpected comment result: c=%q multi=%v", cc, multi)
			}
		})
	}
}
